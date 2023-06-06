package quickws

import (
	"sync"
)

const (
	page     = 1024
	maxIndex = 64
)

func selectIndex(n int) int {
	index := n / page
	return index
}

var pools = make([]sync.Pool, 0, maxIndex+1)

func init() {
	for i := 1; i <= maxIndex; i++ {
		j := i
		pools = append(pools, sync.Pool{
			New: func() interface{} {
				buf := make([]byte, j*page+maxFrameHeaderSize)
				return &buf
			},
		})
	}
}

func getBytes(n int) []byte {
	if n < maxFrameHeaderSize {
		return pools[0].Get().([]byte)
	}

	index := selectIndex(n - maxFrameHeaderSize - 1)
	if index >= len(pools) {
		return make([]byte, n+maxFrameHeaderSize)
	}

	return *pools[index].Get().(*[]byte)
}

func putBytes(bytes []byte) {
	if cap(bytes) < maxFrameHeaderSize {
		panic("putBytes: bytes is too small")
	}
	newN := cap(bytes) - maxFrameHeaderSize - 1
	index := selectIndex(newN)
	if index >= len(pools) {
		return
	}
	pools[index].Put(&bytes)
}
