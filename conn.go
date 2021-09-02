// Copyright 2021 guonaihong. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tinyws

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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

func (c *Conn) failRsv1(op Opcode) bool {
	// 解压缩没有开启
	if !c.decompression {
		return true
	}

	// 不是text和binary
	if op != Text && op != Binary {
		return true
	}

	return false
}

func decode(payload []byte) ([]byte, error) {
	r := bytes.NewReader(payload)
	r2 := decompressNoContextTakeover(r)
	var o bytes.Buffer
	if _, err := io.Copy(&o, r2); err != nil {
		return nil, err
	}
	r2.Close()
	return o.Bytes(), nil
}

func (c *Conn) readLoop() (all []byte, op Opcode, err error) {
	var f frame
	var fragmentFrame *frame

	for {
		f, err = readFrame(c.r)
		if err != nil {
			return
		}

		op = f.opcode
		if fragmentFrame != nil {
			op = fragmentFrame.opcode
		}

		// 检查rsv1 rsv2 rsv3
		if f.rsv1 && c.failRsv1(op) || f.rsv2 || f.rsv3 {
			return nil, f.opcode, c.writeErr(ProtocolError, fmt.Errorf("%w:rsv1(%t) rsv2(%t) rsv2(%t)", ErrRsv123, f.rsv1, f.rsv2, f.rsv3))
		}

		if fragmentFrame != nil && !f.opcode.isControl() {
			if f.opcode == 0 {
				fragmentFrame.payload = append(fragmentFrame.payload, f.payload...)

				// 分段的在这返回
				if f.fin {
					//解压缩
					if fragmentFrame.rsv1 && c.decompression {
						fragmentFrame.payload, err = decode(fragmentFrame.payload)
						if err != nil {
							return
						}
					}
					// 这里的check按道理应该放到f.fin前面， 会更符合rfc的标准, 前提是utf8.Valid修改成流式解析
					// TODO utf8.Valid 修改成流式解析
					if fragmentFrame.opcode == Text && !utf8.Valid(fragmentFrame.payload) {
						c.c.Close()
						return nil, f.opcode, ErrTextNotUTF8
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

			if f.rsv1 && c.decompression {
				f.payload, err = decode(f.payload)
				if err != nil {
					return
				}
			}

			if f.opcode == Text {

				if !utf8.Valid(f.payload) {
					c.c.Close()
					return nil, f.opcode, ErrTextNotUTF8
				}
			}

			return f.payload, f.opcode, err
		case Close, Ping, Pong:
			//  对方发的控制消息太大
			if f.payloadLen > maxControlFrameSize {
				return nil, f.opcode, c.writeErr(ProtocolError, ErrMaxControlFrameSize)
			}
			// Close, Ping, Pong 不能分片
			if !f.fin {

				return nil, f.opcode, c.writeErr(ProtocolError, ErrNOTBeFragmented)
			}

			if f.opcode == Close {
				if len(f.payload) == 0 {
					return nil, f.opcode, c.writeErr(NormalClosure, ErrClosePayloadTooSmall)
				}

				if len(f.payload) < 2 {
					return nil, f.opcode, c.writeErr(ProtocolError, ErrClosePayloadTooSmall)
				}

				if !utf8.Valid(f.payload[2:]) {
					return nil, f.opcode, c.writeErr(ProtocolError, ErrTextNotUTF8)
				}

				code := binary.BigEndian.Uint16(f.payload)
				if !validCode(code) {
					return nil, f.opcode, c.writeErr(ProtocolError, ErrCloseValue)
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

type wrapBuffer struct {
	bytes.Buffer
}

func (w *wrapBuffer) Close() error {
	return nil
}

func (c *Conn) WriteMessage(op Opcode, data []byte) (err error) {
	var f frame
	// 这里可以用sync.Pool优化下
	writeBuf := make([]byte, len(data))
	copy(writeBuf, data)

	f.fin = true
	f.rsv1 = c.compression && (op == Text || op == Binary)
	if f.rsv1 {
		var out wrapBuffer
		w := compressNoContextTakeover(&out, defaultCompressionLevel)
		if _, err = io.Copy(w, bytes.NewReader(data)); err != nil {
			return
		}

		if err = w.Close(); err != nil {
			return
		}
		data = out.Bytes()
	}

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
