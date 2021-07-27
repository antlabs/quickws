package tinyws

import (
	"bufio"
	"io"
	"net"
	"sync"
	"time"
)

type Conn struct {
	r      *bufio.Reader
	w      *bufio.Writer
	c      net.Conn
	rw     sync.Mutex
	client bool
}

func newConn(c net.Conn, rw *bufio.ReadWriter, client bool) *Conn {
	rw.Reader.Reset(c)

	return &Conn{c: c, r: rw.Reader, w: rw.Writer, client: client}
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
	if err = c.c.SetDeadline(time.Now().Add(t)); err != nil {
		return
	}

	defer func() { _ = c.c.SetDeadline(time.Time{}) }()
	return c.readLoop()
}

func (c *Conn) WriteMessage(op Opcode, data []byte) (err error) {
	var f frame
	// 这里可以用sync.Pool优化下
	writeBuf := make([]byte, len(data))
	copy(writeBuf, data)

	f.fin = true
	f.opcode = op
	f.payload = writeBuf
	f.payloadLen = int64(len(data))
	if c.client {
		f.mask = true
		newMask(f.maskValue[:])
	}

	if err := writeFrame(c.w, f); err != nil {
		return err
	}
	return c.w.Flush()
}

func (c *Conn) WriteTimeout(op Opcode, data []byte, t time.Duration) (err error) {
	if err = c.c.SetDeadline(time.Now().Add(t)); err != nil {
		return
	}

	defer func() { _ = c.c.SetDeadline(time.Time{}) }()
	return c.WriteMessage(op, data)
}

func (c *Conn) Close() error {
	return c.c.Close()
}
