package quickws

import (
	"sync"
)

const (
	page     = 1024
	maxIndex = 10
)

func selectIndex(n int) int {
	index := n / page
	return index
}

// 分得细点，会对内存占用有一定的优化

// index =0, max_size = 1024 + 15
// index =1, max_size = 2 * 1024 + 15
// index =2, max_size = 3 * 1024 + 15
// index =3, max_size = 4 * 1024 + 15
// index =4, max_size = 5 * 1024 + 15
// index =5, max_size = 6 * 1024 + 15
// index =6, max_size = 7 * 1024 + 15
// index =7, max_size = 8 * 1024 + 15
// index =8, max_size = 9 * 1024 + 15
// index =9, max_size = 10 * 1024 + 15
var pools = [maxIndex]sync.Pool{
	{New: func() interface{} { // index 0
		buf := make([]byte, 1024+maxFrameHeaderSize)
		return &buf
	}},
	{New: func() interface{} { // index 1
		buf := make([]byte, 2*1024+maxFrameHeaderSize)
		return &buf
	}},
	// 2
	{New: func() interface{} { // index 2
		buf := make([]byte, 3*1024+maxFrameHeaderSize)
		return &buf
	}},
	// 3
	{New: func() interface{} { // index 3
		buf := make([]byte, 4*1024+maxFrameHeaderSize)
		return &buf
	}},
	{New: func() interface{} { // index 4
		buf := make([]byte, 5*1024+maxFrameHeaderSize)
		return &buf
	}},
	{New: func() interface{} { // index 5
		buf := make([]byte, 6*1024+maxFrameHeaderSize)
		return &buf
	}},
	{New: func() interface{} { // index 6
		buf := make([]byte, 7*1024+maxFrameHeaderSize)
		return &buf
	}},
	{New: func() interface{} { // index 7
		buf := make([]byte, 8*1024+maxFrameHeaderSize)
		return &buf
	}},
	{New: func() interface{} { // index 8
		buf := make([]byte, 9*1024+maxFrameHeaderSize)
		return &buf
	}},
	{New: func() interface{} { // index 9
		buf := make([]byte, 10*1024+maxFrameHeaderSize)
		return &buf
	}},
}

func getBytes(n int) []byte {
	if n < maxFrameHeaderSize {
		return pools[0].Get().([]byte)
	}

	index := selectIndex(n - maxFrameHeaderSize)
	if index >= len(pools) {
		return make([]byte, n+maxFrameHeaderSize)
	}

	return *pools[index].Get().(*[]byte)
}

func putBytes(bytes []byte) {
	if cap(bytes) < maxFrameHeaderSize {
		panic("putBytes: bytes is too small")
	}
	newN := cap(bytes) - maxFrameHeaderSize
	index := selectIndex(newN)
	pools[index].Put(&bytes)
	if index > len(pools) {
		return
	}
}
