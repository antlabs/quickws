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
	"errors"
	"io"
	"math"
)

const (
	// 根据5.2描述, 满打满算, 最多14字节
	maxFrameHeaderSize = 14
)

var ErrFramePayloadLength = errors.New("error frame payload length")

type frameHeader struct {
	payloadLen int64
	opcode     Opcode
	maskValue  [4]byte
	rsv1       bool
	rsv2       bool
	rsv3       bool
	mask       bool
	fin        bool
}

type frame struct {
	frameHeader
	payload []byte
}

func readFrame(r *fixedReader) (f frame, err error) {
	h, _, err := readHeader(r)
	if err != nil {
		return f, err
	}

	// 如果缓存区不够, 重新分配

	// h.payloadLen 是要读取body的总数据
	// h.w - h.r 是已经读取未处理的数据
	// 还需要读取的数据等于 h.payloadLen - (h.w - h.r)

	// 已读取未处理的数据
	readUnhandle := int64(r.w - r.r)
	if h.payloadLen-readUnhandle > r.available() {
		// 取得旧的buf
		oldBuf := r.ptr()
		// 获取新的buf
		newBuf := getBytes(int(h.payloadLen) + maxFrameHeaderSize)
		// 重置缓存区
		r.reset(newBuf)
		// 将旧的buf放回池子里
		putBytes(oldBuf)

	}

	// 返回可写的缓存区, 把已经读取的数据去掉，这里是把frame header的数据去掉
	payload := r.free()
	// 前面的reset已经保证了，buffer的大小是够的
	needRead := 0
	if h.payloadLen-readUnhandle > 0 {
		// 还需要读取的数据等于 h.payloadLen - (h.w - h.r)
		needRead = int(h.payloadLen - readUnhandle)
	}

	if r.r != 0 {
		panic("readFrame r != 0")
	}

	if needRead > 0 {
		payload = payload[:needRead]
		// 新建一对新的r w指向尾部的内存区域
		right := r.availableBuf()
		if _, err = io.ReadFull(right, payload); err != nil {
			return f, err
		}

		// right 也有可能超读, 直接加上payload的长度，会把超读的数据给丢了
		r.w += right.w
	}
	r.r += int(h.payloadLen)

	f.payload = r.bytes()[:h.payloadLen]
	f.frameHeader = h
	if h.mask {
		mask(f.payload, f.maskValue[:])
	}

	return f, nil
}

func readHeader(r io.Reader) (h frameHeader, size int, err error) {
	var headArray [maxFrameHeaderSize]byte
	head := headArray[:2]

	n, err := io.ReadFull(r, head)
	if err != nil {
		return
	}
	if n != 2 {
		err = io.ErrUnexpectedEOF
		return
	}
	size = 2

	h.fin = head[0]&(1<<7) > 0
	h.rsv1 = head[0]&(1<<6) > 0
	h.rsv2 = head[0]&(1<<5) > 0
	h.rsv3 = head[0]&(1<<4) > 0
	h.opcode = Opcode(head[0] & 0xF)

	have := 0
	h.mask = head[1]&(1<<7) > 0
	if h.mask {
		have += 4
		size += 4
	}

	h.payloadLen = int64(head[1] & 0x7F)

	switch {
	// 长度
	case h.payloadLen >= 0 && h.payloadLen <= 125:
		if h.payloadLen == 0 && !h.mask {
			return
		}
	case h.payloadLen == 126:
		// 2字节长度
		have += 2
		size += 2
	case h.payloadLen == 127:
		// 8字节长度
		have += 8
		size += 8
	default:
		// 预期之外的, 直接报错
		return h, 0, ErrFramePayloadLength
	}

	head = head[:have]
	_, err = io.ReadFull(r, head)
	if err != nil {
		return
	}

	switch h.payloadLen {
	case 126:
		h.payloadLen = int64(binary.BigEndian.Uint16(head[:2]))
		head = head[2:]
	case 127:
		h.payloadLen = int64(binary.BigEndian.Uint64(head[:8]))
		head = head[8:]
	}

	if h.mask {
		copy(h.maskValue[:], head)
	}

	return
}

// https://datatracker.ietf.org/doc/html/rfc6455#section-5.2
// (the most significant bit MUST be 0)
func writeHeader(w io.Writer, h frameHeader) (err error) {
	var head [maxFrameHeaderSize]byte

	if h.fin {
		head[0] |= 1 << 7
	}

	if h.rsv1 {
		head[0] |= 1 << 6
	}

	if h.rsv2 {
		head[0] |= 1 << 5
	}

	if h.rsv3 {
		head[0] |= 1 << 5
	}

	head[0] |= byte(h.opcode & 0xF)

	have := 2
	switch {
	case h.payloadLen <= 125:
		head[1] = byte(h.payloadLen)
	case h.payloadLen <= math.MaxUint16:
		head[1] = 126
		binary.BigEndian.PutUint16(head[2:], uint16(h.payloadLen))
		have += 2 // 2前
	default:
		head[1] = 127
		binary.BigEndian.PutUint64(head[2:], uint64(h.payloadLen))
		have += 8
	}

	if h.mask {
		head[1] |= 1 << 7
		have += copy(head[have:], h.maskValue[:])
	}

	_, err = w.Write(head[:have])
	return err
}

func writeMessgae(w io.Writer, op Opcode, writeBuf []byte, isClient bool) (err error) {
	var f frame
	f.fin = true
	f.opcode = op
	f.payload = writeBuf
	f.payloadLen = int64(len(writeBuf))
	defer func() {
		f.payload = nil
	}()
	if isClient {
		f.mask = true
		newMask(f.maskValue[:])
	}

	return writeFrame(w, f)
}

// func writeFrame(w io.Writer, f frame) (err error) {
// 	var tmpWriter fixedWriter
// 	tmpWriter.buf = getBytes(len(f.payload) + maxFrameHeaderSize)

// 	var ws io.Writer = &tmpWriter

// 	defer func() {
// 		putBytes(tmpWriter.buf)
// 	}()

// 	if err = writeHeader(ws, f.frameHeader); err != nil {
// 		return
// 	}

// 	wIndex := tmpWriter.w
// 	_, err = ws.Write(f.payload)
// 	if err != nil {
// 		return
// 	}

// 	if f.mask {
// 		mask(tmpWriter.buf[wIndex:tmpWriter.w], f.maskValue[:])
// 	}

// 	_, err = w.Write(tmpWriter.bytes())
// 	return err
// }

func writeFrame(w io.Writer, f frame) (err error) {
	buf := getBytes(len(f.payload) + maxFrameHeaderSize)

	// fmt.Printf("writeFrame getBytes buf = %p\n", buf)
	// var ws io.Writer = bytes.NewBuffer(nil)
	var ws io.Writer = bytes.NewBuffer((*buf)[0:0])

	defer func() {
		putBytes(buf)
	}()
	if err = writeHeader(ws, f.frameHeader); err != nil {
		return
	}

	tmpWriter := ws.(*bytes.Buffer)
	wIndex := tmpWriter.Len()
	_, err = ws.Write(f.payload)
	if err != nil {
		return
	}
	if f.mask {
		mask(tmpWriter.Bytes()[wIndex:], f.maskValue[:])
	}

	// fmt.Printf("writeFrame %#v\n", tmpWriter.Bytes())
	_, err = w.Write(tmpWriter.Bytes())
	return err
}
