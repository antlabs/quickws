package quickws

import (
	"testing"
)

func Benchmark_Rand(t *testing.B) {
	var maskValue [4]byte

	for i := 0; i < t.N; i++ {
		newMask(maskValue[:])
	}
}
