// Copyright 2021-2023 antlabs. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package quickws

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

	// fmt.Printf("opcode:%d", h.opcode)
	assert.Equal(t, string(all), "Hello")
	assert.Equal(t, h.payloadLen, int64(len("Hello")))
}

func Test_Frame_Mask_Read_And_Write(t *testing.T) {
	r := bytes.NewReader(haveMaskData)

	buf := make([]byte, 512)
	f, err := readFrame(r, &buf)
	assert.NoError(t, err)

	// fmt.Printf("opcode:%d", h.opcode)
	assert.Equal(t, string(f.payload[:f.payloadLen]), "Hello")

	var w bytes.Buffer
	assert.NoError(t, writeFrame(&w, f))
	assert.Equal(t, w.Bytes(), haveMaskData)
}

func Test_Frame_Write_NoMask(t *testing.T) {
	// br := bytes.NewReader([]byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f})

	var w bytes.Buffer
	var h frameHeader
	h.payloadLen = int64(5)
	h.opcode = 1
	h.fin = true
	assert.NoError(t, writeHeader(&w, h))
	_, err := w.WriteString("Hello")

	assert.NoError(t, err)
	assert.Equal(t, w.Bytes(), noMaskData)
}
