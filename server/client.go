package network

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	addr string
	conn net.Conn
}

func NewClient(remoteAddr string) *Client {
	return &Client{
		addr: remoteAddr,
	}
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	if _, err := conn.Write([]byte("CONNECT")); err != nil {
		conn.Close()
		return err
	}
	buf := make([]byte, 7)
	if _, err := conn.Read(buf); err != nil {
		return err
	}
	if string(buf) != "CONNACK" {
		return fmt.Errorf("invalid CONNECT response")
	}
	c.conn = conn
	return nil
}

func (c *Client) keepAlive() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		<-ticker.C
		if _, err := c.conn.Write([]byte{0x05}); err != nil {
			continue
		}
	}
}

func (c *Client) readLoop() {
	defer c.conn.Close()
	buf := make([]byte, 1024)
	for {
		_, err := c.conn.Read(buf)
		if err != nil {
			return
		}
		// implement server message type
		// msg := NewMessage()
	}
}

func (c *Client) Run() error {
	go c.keepAlive()
	go c.readLoop()

	errch := make(chan error)
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {

			<-ticker.C
		}
	}()
	return <-errch
}
