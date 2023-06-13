package quickws

import (
	"encoding/binary"
	"testing"
)

func Benchmark_Mask(t *testing.B) {
	var payload [1024]byte
	var maskValue [4]byte

	for i := 0; i < len(payload); i++ {
		payload[i] = byte(i)
	}
	newMask(maskValue[:])
	key := binary.LittleEndian.Uint32(maskValue[:])
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		mask(payload[:], key)
	}
}

func Benchmark_Rand(t *testing.B) {
	var maskValue [4]byte

	for i := 0; i < t.N; i++ {
		newMask(maskValue[:])
	}
}
