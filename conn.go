package tinyws

import (
	"bufio"
	"io"
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
	rw.Reader.Reset(c)

	return &Conn{c: c, r: rw.Reader, w: rw.Writer}
}

func (c *Conn) nextFrame() (h frameHeader, err error) {
	h, err = readHeader(c.r)
	if err != nil {
		return
	}

	return
}

func (c *Conn) readLoop() (all []byte, op Opcode, err error) {
	var f frame
	for {
		f, err = readFrame(c.r)

		if f.opcode != Continuation && !f.opcode.isControl() {
			if err == io.EOF {
				err = nil
			}
			return f.payload, f.opcode, err
		}

		if err != nil {
			return
		}
	}
}

func (c *Conn) ReadMessage() (all []byte, op Opcode, err error) {
	return c.readLoop()
}

func (c *Conn) ReadTimeout(t time.Duration) (all []byte, op Opcode, err error) {
	c.c.SetDeadline(time.Now().Add(t))
	defer c.c.SetDeadline(time.Time{})
	return c.readLoop()
}

func (c *Conn) WriteMessage(op Opcode, data []byte) (err error) {
	var f frame
	f.opcode = op
	f.payload = data

	return writeFrame(c.w, f)
}

func (c *Conn) WriteTimeout(op Opcode, data []byte, t time.Duration) (err error) {
	c.c.SetDeadline(time.Now().Add(t))
	defer c.c.SetDeadline(time.Time{})
	return c.WriteMessage(op, data)
}
