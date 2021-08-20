package tinyws

import (
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

func readFrame(r io.Reader) (f frame, err error) {
	h, err := readHeader(r)
	if err != nil {
		return f, err
	}

	f.payload = make([]byte, h.payloadLen)

	if _, err = io.ReadFull(r, f.payload); err != nil {
		return f, err
	}

	f.frameHeader = h
	if h.mask {
		mask(f.payload, f.maskValue[:])
	}
	return f, nil
}

func readHeader(r io.Reader) (h frameHeader, err error) {

	head := make([]byte, 2, maxFrameHeaderSize)

	_, err = io.ReadFull(r, head)
	if err != nil {
		return
	}

	h.fin = head[0]&(1<<7) > 0
	h.rsv1 = head[0]&(1<<6) > 0
	h.rsv2 = head[0]&(1<<5) > 0
	h.rsv3 = head[0]&(1<<4) > 0
	h.opcode = Opcode(head[0] & 0xF)

	have := 0
	h.mask = head[1]&(1<<7) > 0
	if h.mask {
		have += 4
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
	case h.payloadLen == 127:
		// 8字节长度
		have += 8
	default:
		//预期之外的, 直接报错
		return h, ErrFramePayloadLength
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
	head := make([]byte, maxFrameHeaderSize)

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

func writeFrame(w io.Writer, f frame) (err error) {
	if err = writeHeader(w, f.frameHeader); err != nil {
		return
	}
	if f.mask {
		mask(f.payload, f.maskValue[:])
	}
	_, err = w.Write(f.payload)
	return
}
