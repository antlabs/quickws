package quickws

import (
	"bytes"
	"testing"
)

// 正确性测试
func Test_Mask(t *testing.T) {
	key := uint32(0x12345678)
	for i := 0; i < 1000*4; i++ {
		payload := make([]byte, i)
		for j := 0; j < len(payload); j++ {
			payload[j] = byte(j)
		}

		pay1 := append([]byte(nil), payload...)
		pay2 := append([]byte(nil), payload...)

		maskFast(pay1, key)
		maskSlow(pay2, key)

		if !bytes.Equal(pay1, pay2) {
			t.Fatalf("i = %d, fast.payload != slow.payload:%v, %v", i, pay1, pay2)
		}
	}
}

// TODO 边界测试
func Test_Mask_Boundary(t *testing.T) {
}
