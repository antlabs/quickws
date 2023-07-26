package quickws

import (
	"math/rand"
	"testing"
)

func Benchmark_Rand_Uint32(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = rand.Uint32()
	}
}
