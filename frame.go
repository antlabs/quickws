package tinyws

import "io"

const (
	// 根据5.2描述, 满打满算, 最多14字节
	maxFrameHeaderSize = 14
)

type frameHeader struct {
	payloadLen uint64
	opcode     int32
	rsv1       bool
	rsv2       bool
	rsv3       bool
	mask       bool
	fin        bool
}

func (f frameHeader) read(r io.Reader) (h frameHeader, err error) {

	head := make([]byte, 2, maxFrameHeaderSize)

	_, err := io.ReadFull(r, head)
	if err != nil {
		return h, err
	}

	h.fin = head[0]&1<<7 > 0
	h.rsv1 = head[0]&1<<6 > 0
	h.rsv2 = head[0]&1<<5 > 0
	h.rsv2 = head[0]&1<<4 > 0
	h.opcode = head[0] & 0xF

	h.mask = head[1]&1<<7 > 0
	h.payloadLen = head[1] >> 1

}
