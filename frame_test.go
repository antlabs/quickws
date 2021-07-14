package tinyws

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	noMaskData   = []byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f}
	haveMaskData = []byte{0x81, 0x85, 0x37, 0xfa, 0x21, 0x3d, 0x7f, 0x9f, 0x4d, 0x51, 0x58}
)

func Test_Frame_Read_NoMask(t *testing.T) {
	r := bytes.NewReader(noMaskData)

	h, err := readHeader(r)
	assert.NoError(t, err)
	all, err := io.ReadAll(r)
	assert.NoError(t, err)

	//fmt.Printf("opcode:%d", h.opcode)
	assert.Equal(t, string(all), "Hello")
	assert.Equal(t, h.payloadLen, int64(len("Hello")))
}

func Test_Frame_Mask_Read_And_Write(t *testing.T) {
	r := bytes.NewReader(haveMaskData)

	f, err := readFrame(r)
	assert.NoError(t, err)

	//fmt.Printf("opcode:%d", h.opcode)
	assert.Equal(t, string(f.payload), "Hello")

	var w bytes.Buffer
	writeFrame(&w, f)
	assert.Equal(t, w.Bytes(), haveMaskData)
}

func Test_Frame_Write_NoMask(t *testing.T) {
	//br := bytes.NewReader([]byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f})

	var w bytes.Buffer
	var h frameHeader
	h.payloadLen = int64(5)
	h.opcode = 1
	h.fin = true
	writeHeader(&w, h)
	w.WriteString("Hello")
	assert.Equal(t, w.Bytes(), noMaskData)
}
