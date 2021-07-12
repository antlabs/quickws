package tinyws

import (
	"bufio"
	"net"
	"sync"
	"time"
)

type Conn struct {
	r  *bufio.Reader
	w  *bufio.Writer
	c  net.Conn
	rw sync.Mutex
}

func newConn(c net.Conn, rw *bufio.ReadWriter) *Conn {
	return &Conn{c: c, r: rw.Reader, w: rw.Writer}
}

func (c *Conn) nextFrame() (h frameHeader, err error) {
	h, err = readHeader(c.r)
	if err != nil {
		return
	}

	return
}

func (c *Conn) ReadMessage() (all []byte, op Opcode, err error) {
	return
}

func (c *Conn) ReadTimeout(t time.Time) (all []byte, op Opcode, err error) {
	return
}

func (c *Conn) WriteMessage(op Opcode, data []byte) (err error) {
	return
}

func (c *Conn) WriteTimeout(t time.Time) (err error) {
	return
}
