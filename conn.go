// Copyright 2021-2024 antlabs. All rights reserved.
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
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/antlabs/wsutil/bufio2"
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

// 延迟写
type delayWrite struct {
	delayNum     int32         // 控制延迟写的数量
	delayMu      sync.Mutex    // 延迟写的锁
	delayBuf     *bytes.Buffer // 延迟写的缓冲区
	delayTimeout *time.Timer   // 延迟写的定时器
	delayErr     error         // TODO 原子操作
}

type Conn struct {
	closed int32
	br     *bufio.Reader // read 和fr同时只能使用一个
	*Config
	c      net.Conn
	client bool
	once   sync.Once

	fr fixedreader.FixedReader
	// fw fixedwriter.FixedWriter
	bp bytespool.BytesPool // 实验某些特性加的字段

	delayWrite
	readHeadArray        [enum.MaxFrameHeaderSize]byte
	fragmentFramePayload []byte // 存放分片帧的缓冲区
	bufioPayload         []byte
	fragmentFrameHeader  *frame.FrameHeader
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

func newConn(c net.Conn, client bool, conf *Config, fr fixedreader.FixedReader, read *bufio.Reader, bp bytespool.BytesPool) *Conn {
	_ = setNoDelay(c, conf.tcpNoDelay)

	con := &Conn{
		c:      c,
		client: client,
		Config: conf,
		fr:     fr,
		br:     read,
		bp:     bp,
	}

	return con
}

// 返回标准库的net.Conn
func (c *Conn) NetConn() net.Conn {
	return c.c
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

func (c *Conn) ReadLoop() (err error) {
	c.OnOpen(c)

	defer func() {
		// c.OnClose(c, err)
		c.Close()
		if c.fr.IsInit() {
			defer func() {
				c.fr.Release()
				c.fr.BufPtr()
			}()
		}
	}()

	if c.br != nil {
		newSize := int(1024 * c.bufioMultipleTimesPayloadSize)
		if c.br.Size() != newSize {
			// TODO sync.Pool管理
			(*bufio2.Reader2)(unsafe.Pointer(c.br)).ResetBuf(make([]byte, newSize))
		}
		// bufio 模式才会使用payload
		c.bufioPayload = *bytespool.GetBytes(1024 + enum.MaxFrameHeaderSize)
	}

	for {
		err = c.readMessage()
		if err != nil {
			return err
		}
	}
}

func (c *Conn) StartReadLoop() {
	go func() {
		_ = c.ReadLoop()
	}()
}

func (c *Conn) readDataFromNet(headArray *[enum.MaxFrameHeaderSize]byte, bufioPayload *[]byte) (f frame.Frame, err error) {
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
		f, err = frame.ReadFrameFromReader(c.br, headArray, bufioPayload)
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
func (c *Conn) readMessage() (err error) {
	// 从网络读取数据
	f, err := c.readDataFromNet(&c.readHeadArray, &c.bufioPayload)
	if err != nil {
		return err
	}

	op := f.Opcode
	if c.fragmentFrameHeader != nil {
		op = c.fragmentFrameHeader.Opcode
	}

	rsv1 := f.GetRsv1()
	// 检查Rsv1 rsv2 Rfd, errsv3
	if rsv1 && c.failRsv1(op) || f.GetRsv2() || f.GetRsv3() {
		err = fmt.Errorf("%w:Rsv1(%t) Rsv2(%t) rsv2(%t) compression:%t", ErrRsv123, rsv1, f.GetRsv2(), f.GetRsv3(), c.compression)
		return c.writeErrAndOnClose(ProtocolError, err)
	}

	fin := f.GetFin()
	if c.fragmentFrameHeader != nil && !f.Opcode.IsControl() {
		if f.Opcode == 0 {
			c.fragmentFramePayload = append(c.fragmentFramePayload, f.Payload...)

			// 分段的在这返回
			if fin {
				// 解压缩
				if c.fragmentFrameHeader.GetRsv1() && c.decompression {
					tempBuf, err := decode(c.fragmentFramePayload)
					if err != nil {
						return err
					}
					c.fragmentFramePayload = tempBuf
				}
				// 这里的check按道理应该放到f.Fin前面， 会更符合rfc的标准, 前提是c.utf8Check修改成流式解析
				// TODO c.utf8Check 修改成流式解析
				if c.fragmentFrameHeader.Opcode == opcode.Text && !c.utf8Check(c.fragmentFramePayload) {
					c.Callback.OnClose(c, ErrTextNotUTF8)
					return ErrTextNotUTF8
				}

				c.Callback.OnMessage(c, c.fragmentFrameHeader.Opcode, c.fragmentFramePayload)
				c.fragmentFramePayload = c.fragmentFramePayload[0:0]
				c.fragmentFrameHeader = nil
			}
			return nil
		}

		c.writeErrAndOnClose(ProtocolError, ErrFrameOpcode)
		return ErrFrameOpcode
	}

	if f.Opcode == opcode.Text || f.Opcode == opcode.Binary {
		if !fin {
			prevFrame := f.FrameHeader
			// 第一次分段
			if len(c.fragmentFramePayload) == 0 {
				c.fragmentFramePayload = append(c.fragmentFramePayload, f.Payload...)
				f.Payload = nil
			}

			// 让fragmentFrame的Payload指向readBuf, readBuf 原引用直接丢弃
			c.fragmentFrameHeader = &prevFrame
			return
		}

		if rsv1 && c.decompression {
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
		return
	}

	if f.Opcode == Close || f.Opcode == Ping || f.Opcode == Pong {
		//  对方发的控制消息太大
		if f.PayloadLen > maxControlFrameSize {
			c.writeErrAndOnClose(ProtocolError, ErrMaxControlFrameSize)
			return ErrMaxControlFrameSize
		}
		// Close, Ping, Pong 不能分片
		if !fin {
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
				return
			}
		}

		if f.Opcode == Pong && c.ignorePong {
			return
		}

		c.Callback.OnMessage(c, f.Opcode, nil)
		return
	}
	// 检查Opcode
	c.writeErrAndOnClose(ProtocolError, ErrOpcode)
	return ErrOpcode
}

type wrapBuffer struct {
	bytes.Buffer
}

func (w *wrapBuffer) Close() error {
	return nil
}

func (c *Conn) WriteMessage(op Opcode, writeBuf []byte) (err error) {
	if atomic.LoadInt32(&c.closed) == 1 {
		return ErrClosed
	}

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

	var fw fixedwriter.FixedWriter
	return frame.WriteFrame(&fw, c.c, writeBuf, true, rsv1, c.client, op, maskValue)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.c.SetWriteDeadline(t)
}

func (c *Conn) WriteTimeout(op Opcode, data []byte, t time.Duration) (err error) {
	if err = c.c.SetWriteDeadline(time.Now().Add(t)); err != nil {
		return
	}

	defer func() { _ = c.c.SetWriteDeadline(time.Time{}) }()
	return c.WriteMessage(op, data)
}

func (c *Conn) WriteCloseTimeout(sc StatusCode, t time.Duration) (err error) {
	buf := statusCodeToBytes(sc)
	return c.WriteTimeout(opcode.Close, buf, t)
}

// data 不能超过125字节
func (c *Conn) WritePing(data []byte) (err error) {
	return c.WriteControl(Ping, data[:])
}

// data 不能超过125字节
func (c *Conn) WritePong(data []byte) (err error) {
	return c.WriteControl(Pong, data[:])
}

func (c *Conn) WriteControl(op Opcode, data []byte) (err error) {
	if len(data) > maxControlFrameSize {
		return ErrMaxControlFrameSize
	}
	return c.WriteMessage(op, data)
}

// 写分段数据, 目前主要是单元测试使用
func (c *Conn) writeFragment(op Opcode, writeBuf []byte, maxFragment int /*单个段最大size*/) (err error) {
	if len(writeBuf) < maxFragment {
		return c.WriteMessage(op, writeBuf)
	}

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

	var fw fixedwriter.FixedWriter
	for len(writeBuf) > 0 {
		if len(writeBuf) > maxFragment {
			if err := frame.WriteFrame(&fw, c.c, writeBuf[:maxFragment], false, rsv1, c.client, op, maskValue); err != nil {
				return err
			}
			writeBuf = writeBuf[maxFragment:]
			op = Continuation
			continue
		}
		return frame.WriteFrame(&fw, c.c, writeBuf, true, rsv1, c.client, op, maskValue)
	}
	return nil
}

func (c *Conn) Close() (err error) {
	c.once.Do(func() {
		c.bp.Free()
		err = c.c.Close()
		c.delayMu.Lock()
		if c.delayTimeout != nil {
			c.delayTimeout.Stop()
			c.delayBuf = nil
		}
		c.delayMu.Unlock()
		atomic.StoreInt32(&c.closed, 1)
	})
	return
}

func (c *Conn) writerDelayBufSafe() {
	c.delayMu.Lock()
	c.delayErr = c.writerDelayBufInner()
	c.delayMu.Unlock()
}

func (c *Conn) writerDelayBufInner() (err error) {
	if c.delayBuf == nil || c.delayBuf.Len() == 0 || atomic.LoadInt32(&c.closed) == 1 {
		return nil
	}
	_, err = c.c.Write(c.delayBuf.Bytes())
	if c.delayTimeout != nil {
		c.delayTimeout.Reset(c.maxDelayWriteDuration)
	}
	c.delayNum = 0
	c.delayBuf.Reset()
	return
}

// 对于流量场景这个版本推荐开启tcp delay 方法：WithClientTCPDelay() WithServerTCPDelay()

// 该函数目前是研究性质的尝试
// 延迟写消息, 对流量密集型的场景有用 或者开启tcp delay，
// 1. 如果缓存的消息超过了多少条数
// 2. 如果缓存的消费超过了多久的时间
// 3. TODO: 最大缓存多少字节
func (c *Conn) WriteMessageDelay(op Opcode, writeBuf []byte) (err error) {
	if atomic.LoadInt32(&c.closed) == 1 {
		return ErrClosed
	}

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

	c.delayMu.Lock()
	// 初始化缓存
	if c.delayBuf == nil && c.delayWriteInitBufferSize > 0 {

		// TODO: sync.Pool管理下, 如果size是1k 2k 3k
		delayBuf := make([]byte, 0, c.delayWriteInitBufferSize)
		c.delayBuf = bytes.NewBuffer(delayBuf)
	}
	// 初始化定时器
	if c.delayTimeout == nil && c.maxDelayWriteDuration > 0 {
		c.delayTimeout = time.AfterFunc(c.maxDelayWriteDuration, c.writerDelayBufSafe)
	}
	c.delayMu.Unlock()

	maskValue := uint32(0)
	if c.client {
		maskValue = rand.Uint32()
	}
	// 缓存的消息超过最大值, 则直接写入
	c.delayMu.Lock()
	if c.delayNum+1 == c.maxDelayWriteNum {
		err = frame.WriteFrameToBytes(c.delayBuf, writeBuf, true, rsv1, c.client, op, maskValue)
		if err != nil {
			c.delayMu.Unlock()
			return err
		}
		err = c.writerDelayBufInner()
		c.delayMu.Unlock()
		return err
	}

	// 为了平衡生产者，消费者的速度，这里不使用协程
	if c.delayBuf != nil {
		err = frame.WriteFrameToBytes(c.delayBuf, writeBuf, true, rsv1, c.client, op, maskValue)
	}
	c.delayNum++ // 对记数计+1
	c.delayMu.Unlock()
	return err
}
