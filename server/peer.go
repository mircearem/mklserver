package network

import (
	"context"
	"fmt"
	"io"
	"net"

	_ "github.com/mircearem/mklserver/log"
	"github.com/sirupsen/logrus"
)

type Peer struct {
	conn net.Conn
	// context for cancellation
	ctx context.Context
	// function to trigger cancelation of go routines
	cancel context.CancelFunc
	// channel to handle keep alive signal
	kach chan struct{}
}

func NewPeer(c net.Conn, ctx context.Context) *Peer {
	ctext, cancel := context.WithCancel(ctx)
	return &Peer{
		ctx:    ctext,
		cancel: cancel,
		conn:   c,
		kach:   make(chan struct{}),
	}
}

func (p *Peer) Write(b []byte) error {
	_, err := p.conn.Write(b)
	return err
}

func (p *Peer) ReadLoop(msgch chan *Message) error {
	defer p.conn.Close()
	errch := make(chan error)

	logrus.Info(fmt.Sprintf("accepting message from peer: (%s)\n", ipAddr(p.conn)))
	go func() {
		buf := make([]byte, 2048)
		for {
			n, err := p.conn.Read(buf)
			if err == io.EOF || err == io.ErrClosedPipe {
				logrus.Info()
				errch <- err
				break
			}
			msg := NewMessage(p, buf[:n])
			msgch <- msg
		}
	}()

	for {
		select {
		case <-p.ctx.Done():
			logrus.Warn(fmt.Sprintf("peer: (%s) disconnected", ipAddr(p.conn)))
			return nil
		case err := <-errch:
			return err
		}
	}
}
