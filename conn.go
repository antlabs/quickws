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
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/antlabs/wsutil/bytespool"
	"github.com/antlabs/wsutil/enum"
	"github.com/antlabs/wsutil/fixedreader"
	"github.com/antlabs/wsutil/fixedwriter"
	"github.com/antlabs/wsutil/frame"
	"github.com/antlabs/wsutil/opcode"
)

const (
	maxControlFrameSize = 125
)

// var _ net.Conn = (*Conn)(nil)

type Conn struct {
	read   *bufio.Reader // read 和fr同时只能使用一个
	c      net.Conn
	client bool
	config
	once sync.Once

	fr fixedreader.FixedReader
	fw fixedwriter.FixedWriter
	bp bytespool.BytesPool
}

func setNoDelay(c net.Conn, noDelay bool) error {
	if tcp, ok := c.(*net.TCPConn); ok {
		return tcp.SetNoDelay(noDelay)
	}

	if tlsTCP, ok := c.(*tls.Conn); ok {
		return setNoDelay(tlsTCP.NetConn(), noDelay)
	}
	return nil
}

func newConn(c net.Conn, client bool, conf config, fr fixedreader.FixedReader, read *bufio.Reader, bp bytespool.BytesPool) *Conn {
	_ = setNoDelay(c, conf.tcpNoDelay)

	return &Conn{
		c:      c,
		client: client,
		config: conf,
		fr:     fr,
		read:   read,
		bp:     bp,
	}
}

func (c *Conn) writeErrAndOnClose(code StatusCode, userErr error) error {
	defer c.Callback.OnClose(c, userErr)
	if err := c.WriteTimeout(opcode.Close, statusCodeToBytes(code), 2*time.Second); err != nil {
		return err
	}

	return userErr
}

func (c *Conn) failRsv1(op opcode.Opcode) bool {
	// 解压缩没有开启
	if !c.decompression {
		return true
	}

	// 不是text和binary
	if op != opcode.Text && op != opcode.Binary {
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

func (c *Conn) ReadLoop() error {
	c.OnOpen(c)

	return c.readLoop()
}

func (c *Conn) StartReadLoop() {
	go func() {
		_ = c.ReadLoop()
	}()
}

func (c *Conn) readDataFromNet(headArray *[enum.MaxFrameHeaderSize]byte, payload *[]byte) (f frame.Frame, err error) {
	if c.readTimeout > 0 {
		err = c.c.SetReadDeadline(time.Now().Add(c.readTimeout))
		if err != nil {
			c.Callback.OnClose(c, err)
			return
		}
	}

	if c.fr.IsInit() {
		f, err = frame.ReadFrameFromWindows(&c.fr, headArray, c.windowsMultipleTimesPayloadSize)
	} else {
		f, err = frame.ReadFrameFromReader(c.read, headArray, payload)
	}
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

// 读取websocket frame.Frame的循环
func (c *Conn) readLoop() error {
	var f frame.Frame
	var fragmentFrameHeader *frame.FrameHeader

	defer c.Close()

	var err error
	var op opcode.Opcode

	if c.fr.IsInit() {
		defer func() {
			c.fr.Release()
			c.fr.BufPtr()
		}()
	}

	var fragmentFrameBuf []byte
	var headArray [enum.MaxFrameHeaderSize]byte

	var payload []byte
	if c.read != nil {
		payload = *bytespool.GetBytes(1024 + enum.MaxFrameHeaderSize)
	}
	for {

		// 从网络读取数据
		f, err = c.readDataFromNet(&headArray, &payload)
		if err != nil {
			return err
		}

		op = f.Opcode
		if fragmentFrameHeader != nil {
			op = fragmentFrameHeader.Opcode
		}

		// 检查Rsv1 rsv2 Rsv3
		if f.Rsv1 && c.failRsv1(op) || f.Rsv2 || f.Rsv3 {
			err = fmt.Errorf("%w:Rsv1(%t) Rsv2(%t) rsv2(%t) compression:%t", ErrRsv123, f.Rsv1, f.Rsv2, f.Rsv3, c.compression)
			return c.writeErrAndOnClose(ProtocolError, err)
		}

		if fragmentFrameHeader != nil && !f.Opcode.IsControl() {
			if f.Opcode == 0 {
				fragmentFrameBuf = append(fragmentFrameBuf, f.Payload...)

				// 分段的在这返回
				if f.Fin {
					// 解压缩
					if fragmentFrameHeader.Rsv1 && c.decompression {
						tmpeBuf, err := decode(fragmentFrameBuf)
						if err != nil {
							return err
						}
						fragmentFrameBuf = tmpeBuf
					}
					// 这里的check按道理应该放到f.Fin前面， 会更符合rfc的标准, 前提是c.utf8Check修改成流式解析
					// TODO c.utf8Check 修改成流式解析
					if fragmentFrameHeader.Opcode == opcode.Text && !c.utf8Check(fragmentFrameBuf) {
						c.Callback.OnClose(c, ErrTextNotUTF8)
						return ErrTextNotUTF8
					}

					c.Callback.OnMessage(c, fragmentFrameHeader.Opcode, fragmentFrameBuf)
					fragmentFrameBuf = fragmentFrameBuf[0:0]
					fragmentFrameHeader = nil
				}
				continue
			}

			c.writeErrAndOnClose(ProtocolError, ErrFrameOpcode)
			return ErrFrameOpcode
		}

		// 检查Opcode
		switch f.Opcode {
		case opcode.Text, opcode.Binary:
			if !f.Fin {
				prevFrame := f.FrameHeader
				// 第一次分段
				if len(fragmentFrameBuf) == 0 {
					fragmentFrameBuf = append(fragmentFrameBuf, f.Payload...)
					f.Payload = nil
				}

				// 让fragmentFrame的Payload指向readBuf, readBuf 原引用直接丢弃
				fragmentFrameHeader = &prevFrame
				continue
			}

			if f.Rsv1 && c.decompression {
				// 不分段的解压缩
				f.Payload, err = decode(f.Payload)
				if err != nil {
					return err
				}
			}

			if f.Opcode == opcode.Text {
				if !c.utf8Check(f.Payload) {
					c.c.Close()
					c.Callback.OnClose(c, ErrTextNotUTF8)
					return ErrTextNotUTF8
				}
			}

			c.Callback.OnMessage(c, f.Opcode, f.Payload)
		case Close, Ping, Pong:
			//  对方发的控制消息太大
			if f.PayloadLen > maxControlFrameSize {
				c.writeErrAndOnClose(ProtocolError, ErrMaxControlFrameSize)
				return ErrMaxControlFrameSize
			}
			// Close, Ping, Pong 不能分片
			if !f.Fin {
				c.writeErrAndOnClose(ProtocolError, ErrNOTBeFragmented)
				return ErrNOTBeFragmented
			}

			if f.Opcode == Close {
				if len(f.Payload) == 0 {
					return c.writeErrAndOnClose(NormalClosure, ErrClosePayloadTooSmall)
				}

				if len(f.Payload) < 2 {
					return c.writeErrAndOnClose(ProtocolError, ErrClosePayloadTooSmall)
				}

				if !c.utf8Check(f.Payload[2:]) {
					return c.writeErrAndOnClose(ProtocolError, ErrTextNotUTF8)
				}

				code := binary.BigEndian.Uint16(f.Payload)
				if !validCode(code) {
					return c.writeErrAndOnClose(ProtocolError, ErrCloseValue)
				}

				// 回敬一个close包
				if err := c.WriteTimeout(Close, f.Payload, 2*time.Second); err != nil {
					return err
				}

				err = bytesToCloseErrMsg(f.Payload)
				c.Callback.OnClose(c, err)
				return err
			}

			if f.Opcode == Ping {
				// 回一个pong包
				if c.replyPing {
					if err := c.WriteTimeout(Pong, f.Payload, 2*time.Second); err != nil {
						c.Callback.OnClose(c, err)
						return err
					}
					c.Callback.OnMessage(c, f.Opcode, f.Payload)
					continue
				}
			}

			if f.Opcode == Pong && c.ignorePong {
				continue
			}

			c.Callback.OnMessage(c, f.Opcode, nil)
		default:
			c.writeErrAndOnClose(ProtocolError, ErrOpcode)
			return ErrOpcode
		}

	}
}

type wrapBuffer struct {
	bytes.Buffer
}

func (w *wrapBuffer) Close() error {
	return nil
}

func (c *Conn) WriteMessage2(op Opcode, writeBuf []byte) (err error) {
	if op == opcode.Text {
		if !c.utf8Check(writeBuf) {
			return ErrTextNotUTF8
		}
	}

	rsv1 := c.compression && (op == opcode.Text || op == opcode.Binary)
	if rsv1 {
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

	// f.Opcode = op
	// f.PayloadLen = int64(len(writeBuf))
	maskValue := uint32(0)
	if c.client {
		maskValue = rand.Uint32()
	}

	return frame.WriteFrame2(&c.fw, c.c, writeBuf, rsv1, c.client, op, maskValue)
}

func (c *Conn) WriteMessage(op Opcode, writeBuf []byte) (err error) {
	var f frame.FrameHeader

	if op == opcode.Text {
		if !c.utf8Check(writeBuf) {
			return ErrTextNotUTF8
		}
	}

	f.Fin = true
	f.Rsv1 = c.compression && (op == opcode.Text || op == opcode.Binary)
	if f.Rsv1 {
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

	f.Opcode = op
	f.PayloadLen = int64(len(writeBuf))
	if c.client {
		f.Mask = true
		newMask(f.MaskValue[:])
	}

	return frame.WriteFrame(c.c, f, writeBuf, &c.fw)
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
	return c.WriteTimeout(opcode.Close, buf, t)
}

func (c *Conn) Close() (err error) {
	c.once.Do(func() {
		c.bp.Free()
		err = c.c.Close()
	})
	return
}
