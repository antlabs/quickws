package quickws

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func splitString(s string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

func Test_Reader_Small(t *testing.T) {
	var out bytes.Buffer

	tmp := append([]byte(nil), testTextMessage64kb...)
	err := writeMessgae(&out, Text, tmp, true)
	hexString := hex.EncodeToString(out.Bytes())
	// 在每两个字符之间插入空格
	spacedHexString := strings.Join(splitString(hexString, 2), ", ")
	fmt.Printf("header: %+v\n", spacedHexString[:100])
	assert.NoError(t, err)

	r := newBuffer(&out, getBytes(1024+maxFrameHeaderSize))

	f, err := readFrame(r)
	fmt.Printf("header: %+v\n", f.frameHeader)
	assert.NoError(t, err)
	// err = os.WriteFile("./test_reader.dat", f.payload, 0o644)

	assert.NoError(t, err)
	assert.Equal(t, f.payload, testTextMessage64kb)
}

func Test_Reader_WriteMulti_ReadOne(t *testing.T) {
	var out bytes.Buffer

	for i := 1024 * 63; i <= 1024*63+1; i++ {
		need := make([]byte, 0, i)
		got := make([]byte, 0, i)
		for j := 0; j < i; j++ {
			need = append(need, byte(j))
			got = append(got, byte(j))
		}

		for j := 0; j < 1; j++ {
			err := writeMessgae(&out, Text, need, true)
			assert.NoError(t, err)
			err = writeMessgae(&out, Text, need, true)
			assert.NoError(t, err)
		}
		fmt.Printf("i = %d, need: len(%d), write.size:%d\n", i, len(need), out.Len())

		b := getBytes(1024 + maxFrameHeaderSize)
		r := newBuffer(&out, b)
		for j := 0; j < 2; j++ {

			f, err := readFrame(r)
			assert.NoError(t, err)

			assert.NoError(t, err)
			// TODO
			if j == 0 {
				continue
			}
			if !bytes.Equal(f.payload, got) {
				t.Fatalf("bad test index:%d\n", i)
				return
			}
			// assert.Equal(t, f.payload, got, fmt.Sprintf("index:%d", i))
			if err != nil {
				return
			}
		}
		putBytes(b)
		out.Reset()
	}
}

// 测试只写一次数据包，但是分多次读取
func Test_Reader_WriteOne_ReadMulti(t *testing.T) {
	var out bytes.Buffer

	for i := 1031; i <= 1024*64; i++ {
		need := make([]byte, 0, i)
		got := make([]byte, 0, i)
		for j := 0; j < i; j++ {
			need = append(need, byte(j))
			got = append(got, byte(j))
		}

		err := writeMessgae(&out, Text, need, true)
		assert.NoError(t, err)

		b := getBytes(1024 + maxFrameHeaderSize)
		r := newBuffer(&out, b)

		f, err := readFrame(r)
		assert.NoError(t, err)
		if err != nil {
			return
		}
		putBytes(b)
		out.Reset()

		assert.NoError(t, err)
		// TODO
		if i == 0 {
			continue
		}
		assert.Equal(t, f.payload, got, fmt.Sprintf("index:%d", i))
	}
}

func Test_Reset(t *testing.T) {
	var out bytes.Buffer
	out.Write([]byte("1234"))
	r := newBuffer(&out, getBytes(1024+maxFrameHeaderSize))

	small := make([]byte, 2)

	r.Read(small)
	r.reset(getBytes(1024*2 + maxFrameHeaderSize))
	assert.Equal(t, r.bytes()[:2], []byte("34"))
	assert.Equal(t, r.free()[:2], []byte{0, 0})
}

func Test_Reader_WriteMulti_ReadOne_64512(t *testing.T) {
	var out bytes.Buffer

	for i := 64512; i <= 64512; i++ {
		need := make([]byte, 0, i)
		got := make([]byte, 0, i)
		for j := 0; j < i; j++ {
			need = append(need, byte(j))
			got = append(got, byte(j))
		}

		for j := 0; j < 1; j++ {
			err := writeMessgae(&out, Text, need, true)
			assert.NoError(t, err)
			err = writeMessgae(&out, Text, need, true)
			assert.NoError(t, err)
		}
		fmt.Printf("i = %d, need: len(%d), write.size:%d\n", i, len(need), out.Len())

		b := getBytes(1024 + maxFrameHeaderSize)
		r := newBuffer(&out, b)
		for j := 0; j < 2; j++ {

			f, err := readFrame(r)
			assert.NoError(t, err)

			assert.NoError(t, err)
			// TODO
			if j == 0 {
				continue
			}
			if !bytes.Equal(f.payload, got) {
				t.Fatalf("bad test index:%d\n", i)
				return
			}
			// assert.Equal(t, f.payload, got, fmt.Sprintf("index:%d", i))
			if err != nil {
				return
			}
		}
		putBytes(b)
		out.Reset()
	}
}

func Test_Reader_WriteMulti_ReadOne_65536(t *testing.T) {
	var out bytes.Buffer

	for i := 65536; i <= 64512; i++ {
		need := make([]byte, 0, i)
		got := make([]byte, 0, i)
		for j := 0; j < i; j++ {
			need = append(need, byte(j))
			got = append(got, byte(j))
		}

		for j := 0; j < 1; j++ {
			err := writeMessgae(&out, Text, need, true)
			assert.NoError(t, err)
			err = writeMessgae(&out, Text, need, true)
			assert.NoError(t, err)
		}
		fmt.Printf("i = %d, need: len(%d), write.size:%d\n", i, len(need), out.Len())

		b := getBytes(1024 + maxFrameHeaderSize)
		r := newBuffer(&out, b)
		for j := 0; j < 2; j++ {

			f, err := readFrame(r)
			if err != io.EOF {
				assert.NoError(t, err)
			}

			if !bytes.Equal(f.payload, got) {
				t.Fatalf("bad test index:%d\n", i)
				return
			}
			// assert.Equal(t, f.payload, got, fmt.Sprintf("index:%d", i))
			if err != nil {
				return
			}
		}
		putBytes(b)
		out.Reset()
	}
}
