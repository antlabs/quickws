package quickws

import "unsafe"

var mask func(payload []byte, key uint32)

func init() {
	i := uint32(1)
	b := *(*bool)(unsafe.Pointer(&i))

	if b {
		// 小端机器
		mask = maskFast
	} else {
		// 大端机器
		mask = maskSlow
	}
}
