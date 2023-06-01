package quickws

import (
	"bytes"
	"testing"
)

// BenchmarkReadFrame tests the performance of the ReadFrame function.
func Benchmark_ReadFrame(b *testing.B) {
	r := bytes.NewReader(noMaskData)
	for i := 0; i < b.N; i++ {

		r.Reset(noMaskData)
		_, err := readHeader(r)
		if err != nil {
			b.Fatal(err)
		}
	}
}
