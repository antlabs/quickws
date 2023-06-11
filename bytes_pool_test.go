package quickws

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func Test_Index(t *testing.T) {
	for i := 0; i <= 1024+maxFrameHeaderSize; i++ {
		i2 := i
		if i2 >= maxFrameHeaderSize {
			i2 -= (maxFrameHeaderSize + 1)
		}
		index := selectIndex(i2)
		assert.Equal(t, index, 0)
	}

	for i := 1024 + maxFrameHeaderSize + 1; i <= 2*1024+maxFrameHeaderSize; i++ {
		i2 := i
		i2 -= (maxFrameHeaderSize + 1)
		index := selectIndex(i2)
		assert.Equal(t, index, 1)
	}

	for i := 1024*2 + maxFrameHeaderSize + 1; i <= 3*1024+maxFrameHeaderSize; i++ {
		i2 := i
		i2 -= (maxFrameHeaderSize + 1)
		index := selectIndex(i2)
		assert.Equal(t, index, 2)
	}
}

func Test_GetBytes_Address(t *testing.T) {
	var m map[unsafe.Pointer]bool
	for i := 0; i < 10; i++ {
		p := getBytes(1)
		if m[unsafe.Pointer(p)] {
			t.Fatal("duplicate pointer")
		}
	}
}
