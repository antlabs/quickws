// Copyright 2021-2023 antlabs. All rights reserved.
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

package quickws

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
	"unicode/utf8"
)

const (
	maxControlFrameSize = 125
)

// var _ net.Conn = (*Conn)(nil)

type Conn struct {
	c      net.Conn
	client bool
	config
}

func newConn(c net.Conn, client bool, conf config) *Conn {
	return &Conn{c: c, client: client, config: conf}
}

func (c *Conn) writeErrAndOnClose(code StatusCode, userErr error) error {
	defer c.Callback.OnClose(c, userErr)
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

func (c *Conn) ReadLoop() {
	c.OnOpen(c)

	c.readLoop()
}

func (c *Conn) readDataFromNet(fixedBuf *fixedReader, headArray *[maxFrameHeaderSize]byte) (f frame, err error) {
	if c.readTimeout > 0 {
		err = c.c.SetReadDeadline(time.Now().Add(c.readTimeout))
		if err != nil {
			c.Callback.OnClose(c, err)
		}
	}

	f, err = readFrame(fixedBuf, headArray)
	if err != nil {
		c.Callback.OnClose(c, err)
		return
	}

	if c.readTimeout > 0 {
		if err = c.c.SetReadDeadline(time.Time{}); err != nil {
			c.Callback.OnClose(c, err)
		}
	}
	return
}

// 读取websocket frame的循环
func (c *Conn) readLoop() {
	var f frame
	var fragmentFrameHeader *frameHeader

	defer c.Close()

	var err error
	var op Opcode

	// 默认最小1k + 14
	fixedBuf := newBuffer(c.c, getBytes(1024+maxFrameHeaderSize))
	defer fixedBuf.release()

	var fragmentFrameBuf []byte
	var headArray [maxFrameHeaderSize]byte
	for {

		// 从网络读取数据
		f, err = c.readDataFromNet(fixedBuf, &headArray)
		if err != nil {
			return
		}

		op = f.opcode
		if fragmentFrameHeader != nil {
			op = fragmentFrameHeader.opcode
		}

		// 检查rsv1 rsv2 rsv3
		if f.rsv1 && c.failRsv1(op) || f.rsv2 || f.rsv3 {
			err = fmt.Errorf("%w:rsv1(%t) rsv2(%t) rsv2(%t)", ErrRsv123, f.rsv1, f.rsv2, f.rsv3)
			c.writeErrAndOnClose(ProtocolError, err)
			return
		}

		if fragmentFrameHeader != nil && !f.opcode.isControl() {
			if f.opcode == 0 {
				fragmentFrameBuf = append(fragmentFrameBuf, f.payload...)

				// 分段的在这返回
				if f.fin {
					// 解压缩
					if fragmentFrameHeader.rsv1 && c.decompression {
						tmpeBuf, err := decode(fragmentFrameBuf)
						if err != nil {
							return
						}
						fragmentFrameBuf = tmpeBuf
					}
					// 这里的check按道理应该放到f.fin前面， 会更符合rfc的标准, 前提是utf8.Valid修改成流式解析
					// TODO utf8.Valid 修改成流式解析
					if fragmentFrameHeader.opcode == Text && !utf8.Valid(fragmentFrameBuf) {
						c.Callback.OnClose(c, ErrTextNotUTF8)
						return
					}

					c.Callback.OnMessage(c, fragmentFrameHeader.opcode, fragmentFrameBuf)
				}
				continue
			}

			c.writeErrAndOnClose(ProtocolError, ErrFrameOpcode)
			return
		}

		// 检查opcode
		switch f.opcode {
		case Text, Binary:
			if !f.fin {
				prevFrame := f.frameHeader
				// 第一次分段
				if fragmentFrameBuf == nil {
					fragmentFrameBuf = f.payload
					f.payload = nil

					newBuf := getBytes(len(fixedBuf.bytes()))
					fixedBuf.reset(newBuf)
				}

				// 让fragmentFrame的payload指向readBuf, readBuf 原引用直接丢弃
				fragmentFrameHeader = &prevFrame
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
					c.Callback.OnClose(c, ErrTextNotUTF8)
					return
				}
			}

			c.Callback.OnMessage(c, f.opcode, f.payload)
		case Close, Ping, Pong:
			//  对方发的控制消息太大
			if f.payloadLen > maxControlFrameSize {
				c.writeErrAndOnClose(ProtocolError, ErrMaxControlFrameSize)
				return
			}
			// Close, Ping, Pong 不能分片
			if !f.fin {
				c.writeErrAndOnClose(ProtocolError, ErrNOTBeFragmented)
				return
			}

			if f.opcode == Close {
				if len(f.payload) == 0 {
					c.writeErrAndOnClose(NormalClosure, ErrClosePayloadTooSmall)
					return
				}

				if len(f.payload) < 2 {
					c.writeErrAndOnClose(ProtocolError, ErrClosePayloadTooSmall)
					return
				}

				if !utf8.Valid(f.payload[2:]) {
					c.writeErrAndOnClose(ProtocolError, ErrTextNotUTF8)
					return
				}

				code := binary.BigEndian.Uint16(f.payload)
				if !validCode(code) {
					c.writeErrAndOnClose(ProtocolError, ErrCloseValue)
					return
				}

				// 回敬一个close包
				if err := c.WriteTimeout(Close, f.payload, 2*time.Second); err != nil {
					return
				}

				c.Callback.OnClose(c, bytesToCloseErrMsg(f.payload))
				return
			}

			if f.opcode == Ping {
				// 回一个pong包
				if c.replyPing {
					if err := c.WriteTimeout(Pong, f.payload, 2*time.Second); err != nil {
						c.Callback.OnClose(c, err)
						return
					}
					c.Callback.OnMessage(c, f.opcode, f.payload)
					continue
				}
			}

			if f.opcode == Pong && c.ignorePong {
				continue
			}

			c.Callback.OnMessage(c, f.opcode, nil)
		default:
			c.writeErrAndOnClose(ProtocolError, ErrOpcode)
			return
		}

	}
}

type wrapBuffer struct {
	bytes.Buffer
}

func (w *wrapBuffer) Close() error {
	return nil
}

func (c *Conn) WriteMessage(op Opcode, writeBuf []byte) (err error) {
	var f frame

	if op == Text {
		if !utf8.Valid(writeBuf) {
			return ErrTextNotUTF8
		}
	}

	f.fin = true
	f.rsv1 = c.compression && (op == Text || op == Binary)
	if f.rsv1 {
		var out wrapBuffer
		w := compressNoContextTakeover(&out, defaultCompressionLevel)
		if _, err = io.Copy(w, bytes.NewReader(writeBuf)); err != nil {
			return
		}

		if err = w.Close(); err != nil {
			return
		}
		writeBuf = out.Bytes()
	}

	f.opcode = op
	f.payload = writeBuf
	f.payloadLen = int64(len(writeBuf))
	if c.client {
		f.mask = true
		newMask(f.maskValue[:])
	}

	return writeFrame(c.c, f)
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.c.SetDeadline(t)
}

func (c *Conn) WriteTimeout(op Opcode, data []byte, t time.Duration) (err error) {
	if err = c.c.SetDeadline(time.Now().Add(t)); err != nil {
		return
	}

	defer func() { _ = c.c.SetDeadline(time.Time{}) }()
	return c.WriteMessage(op, data)
}

func (c *Conn) WriteCloseTimeout(sc StatusCode, t time.Duration) (err error) {
	buf := statusCodeToBytes(sc)
	return c.WriteTimeout(Close, buf, t)
}

func (c *Conn) Close() error {
	return c.c.Close()
}
