package tinyws

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var noMaskData = []byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f}

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
