package tinyws

import (
	"bufio"
	"net"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	maxControlFrameSize = 125
)

type Conn struct {
	r      *bufio.Reader
	w      *bufio.Writer
	c      net.Conn
	rw     sync.Mutex
	client bool
	config
}

func newConn(c net.Conn, rw *bufio.ReadWriter, client bool, conf config) *Conn {
	rw.Reader.Reset(c)

	return &Conn{c: c, r: rw.Reader, w: rw.Writer, client: client, config: conf}
}

func (c *Conn) writeErr(code StatusCode, userErr error) error {
	if err := c.WriteTimeout(Close, statusCodeToBytes(code), 2*time.Second); err != nil {
		return err
	}

	return userErr
}

func (c *Conn) readLoop() (all []byte, op Opcode, err error) {
	var f frame
	var fragmentFrame *frame

	for {
		f, err = readFrame(c.r)
		if err != nil {
			return
		}

		// 检查rsv1 rsv2 rsv3
		if f.rsv1 || f.rsv2 || f.rsv3 {
			return nil, f.opcode, c.writeErr(ProtocolError, ErrRsv123)
		}

		if fragmentFrame != nil && !f.opcode.isControl() {
			if f.opcode == 0 {
				fragmentFrame.payload = append(fragmentFrame.payload, f.payload...)
				// 分段的在这返回
				if f.fin {
					if !utf8.Valid(f.payload) {
						return nil, f.opcode, c.writeErr(DataCannotAccept, ErrTextNotUTF8)
					}

					return fragmentFrame.payload, fragmentFrame.opcode, nil
				}
				continue
			}

			return nil, f.opcode, c.writeErr(ProtocolError, ErrFrameOpcode)
		}

		// 检查opcode
		switch f.opcode {
		case Text, Binary:
			if !f.fin {
				f2 := f
				fragmentFrame = &f2
				continue
			}

			if f.opcode == Text {
				if !utf8.Valid(f.payload) {
					return nil, f.opcode, c.writeErr(DataCannotAccept, ErrTextNotUTF8)
				}
			}

			return f.payload, f.opcode, err
		case Close, Ping, Pong:
			//  对方发的控制消息太大
			if f.payloadLen > maxControlFrameSize {
				return nil, f.opcode, c.writeErr(ProtocolError, ErrMaxControlFrameSize)
			}

			if !f.fin {
				// 不能分片
				return nil, f.opcode, c.writeErr(ProtocolError, ErrNOTBeFragmented)
			}

			if f.opcode == Close {
				if !utf8.Valid(f.payload) {
					return nil, f.opcode, c.writeErr(ProtocolError, ErrTextNotUTF8)
				}

				// 回敬一个close包
				if err := c.WriteTimeout(Close, f.payload, 2*time.Second); err != nil {
					return nil, f.opcode, err
				}

				return nil, Close, bytesToCloseErrMsg(f.payload)
			}

			if f.opcode == Ping {
				// 回一个pong包
				if c.replyPing {
					if err := c.WriteTimeout(Pong, f.payload, 2*time.Second); err != nil {
						return nil, f.opcode, err
					}
				}
			}
		default:
			return nil, f.opcode, c.writeErr(ProtocolError, ErrOpcode)
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
