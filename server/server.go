package network

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type ServerConfig struct {
	listenAddr       string
	handshakeTimeout time.Duration
	keepAlive        time.Duration
}

func NewServerConfig() ServerConfig {
	return ServerConfig{
		listenAddr:       "3000",
		handshakeTimeout: 2 * time.Second,
		keepAlive:        15 * time.Second,
	}
}

func (s ServerConfig) WithListenAddr(addr string) ServerConfig {
	s.listenAddr = addr
	return s
}

func (s ServerConfig) WithHandshakeTimeout(t time.Duration) ServerConfig {
	s.handshakeTimeout = t
	return s
}

func (s ServerConfig) WithKeepAlive(t time.Duration) ServerConfig {
	s.keepAlive = t
	return s
}

type Server struct {
	ctx       context.Context
	config    ServerConfig
	ln        net.Listener
	mu        sync.RWMutex
	peers     map[string]*Peer
	addPeerch chan *Peer
	delPeerch chan *Peer
	msgCh     chan *Message
}

func NewServer(config ServerConfig, ctx context.Context) (*Server, error) {
	ln, err := net.Listen("tcp", config.listenAddr)
	if err != nil {
		return nil, err
	}

	return &Server{
		ctx:       ctx,
		config:    config,
		ln:        ln,
		peers:     make(map[string]*Peer),
		addPeerch: make(chan *Peer),
		delPeerch: make(chan *Peer),
		msgCh:     make(chan *Message, 1024),
	}, nil
}

func (s *Server) Run() {
	go s.acceptLoop()
	for {
		select {
		case peer := <-s.addPeerch:
			if err := s.handshake(peer); err != nil {
				logrus.Warn(err)
				peer.conn.Close()
				continue
			}
			s.addPeer(peer)
			go s.handlePeer(peer)
		case peer := <-s.delPeerch:
			logrus.Warn(fmt.Sprintf("received drop signal for peer: (%s)", ipAddr(peer.conn)))
			s.delPeer(peer)
		case msg := <-s.msgCh:
			s.handleMessage(msg)
		}
	}
}

func (s *Server) acceptLoop() {
	logrus.Info(fmt.Sprintf("accepting connections on: (%s)", s.ln.Addr().String()))
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			logrus.Info(fmt.Sprintf("accept error: (%s)\n", err.Error()))
			continue
		}
		// log new connection
		logrus.Info(fmt.Sprintf("new connection from (%s)\n", ipAddr(conn)))
		p := NewPeer(conn, s.ctx)
		s.addPeerch <- p
	}
}

func (s *Server) handshake(p *Peer) error {
	// Check if the peer was already onboarded
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.peers[ipAddr(p.conn)]; ok {
		msg := fmt.Sprintf("peer: (%s) already connected", ipAddr(p.conn))
		return errors.New(msg)
	}
	// Signal timeout after 2 seconds
	errch := make(chan error)
	donech := make(chan struct{})

	timeoutch := time.After(s.config.handshakeTimeout)

	go func() {
		defer func() { close(donech) }()
		buf := make([]byte, 7)
		// Check for error when reading from the connection
		n, err := p.conn.Read(buf)
		if err != nil {
			msg := fmt.Sprintf("onboard error: (%s)", err.Error())
			errch <- errors.New(msg)
			return
		}
		if n != 7 || string(buf) != "CONNECT" {
			msg := fmt.Sprintf("invalid CONNECT request from peer: (%s)", ipAddr(p.conn))
			errch <- errors.New(msg)
			return
		}
		// Write CONNACK response
		if err := p.Write([]byte("CONNACK")); err != nil {
			msg := fmt.Sprintf("unable to send CONNACK response to peer: (%s)", ipAddr(p.conn))
			errch <- errors.New(msg)
			return
		}
	}()

	select {
	case <-timeoutch:
		return fmt.Errorf("peer: (%s), handshake timeout", ipAddr(p.conn))
	case err := <-errch:
		return err
	case <-donech:
		return nil
	}
}

// Move this to the handlePeer function
func (s *Server) handlePeer(p *Peer) {
	go s.handleKeepAlive(p)
	p.ReadLoop(s.msgCh)
}

// Add a new peer to the map
func (s *Server) addPeer(p *Peer) {
	s.mu.Lock()
	s.peers[ipAddr(p.conn)] = p
	s.mu.Unlock()
	logrus.Info(fmt.Sprintf("successfully registered peer: (%s)", ipAddr(p.conn)))
}

// Remove a peer from the map
func (s *Server) delPeer(p *Peer) {
	s.mu.Lock()
	delete(s.peers, ipAddr(p.conn))
	s.mu.Unlock()
	p.cancel()
}

// Function to handle the messages coming from the peer
func (s *Server) handleMessage(m *Message) error {
	// 1. Check what kind of message was received
	n := len(m.Payload)
	switch n {
	// Send signal to server to keep the connection alive
	case 1:
		m.Sender.kach <- struct{}{}
		logrus.Info(fmt.Sprintf("received KA packet from peer: (%s)", ipAddr(m.Sender.conn)))
	default:
		logrus.WithFields(logrus.Fields{
			"sender":  m.Sender,
			"payload": m.Payload,
		})
	}
	return nil
}

// Run in separate go routine and listen for keep
// alive packets (heartbeat) send by the peer
func (s *Server) handleKeepAlive(p *Peer) {
	ticker := time.NewTicker(s.config.keepAlive)
	for {
		select {
		// Drop connection, remove from map
		case <-ticker.C:
			logrus.Warn(fmt.Sprintf("peer: (%s) has timed out, dropping connection", ipAddr(p.conn)))
			s.delPeerch <- p
			return
		// New packet received, send back ACK
		case <-p.kach:
			p.Write([]byte{0x06})
			ticker.Reset(s.config.keepAlive)
		}
	}
}

func ipAddr(c net.Conn) string {
	addr := c.RemoteAddr().String()
	return addr
	// return strings.Split(addr, ":")[0]
}
