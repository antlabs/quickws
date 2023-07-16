package quickws

import (
	"math/rand"
	"testing"
)

func Benchmark_Rand(t *testing.B) {
	var maskValue [4]byte

	for i := 0; i < t.N; i++ {
		newMask(maskValue[:])
	}
}

func Benchmark_Rand_Uint32(t *testing.B) {
	for i := 0; i < t.N; i++ {
		// newMask(maskValue[:])
		_ = rand.Uint32()
	}
}
