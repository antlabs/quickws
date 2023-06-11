package quickws

import "fmt"

type fixedWriter struct {
	buf []byte
	w   int
}

func (fw *fixedWriter) Write(p []byte) (n int, err error) {
	if len(fw.buf[fw.w:]) < len(p) {
		panic(fmt.Sprintf("fixedWriter: buf is too small: %d:%d < %d", len(fw.buf[fw.w:]), cap(fw.buf), cap(p)))
	}
	n = copy(fw.buf[fw.w:], p)
	fw.w += n
	return n, nil
}

func (fw *fixedWriter) bytes() []byte {
	return fw.buf[:fw.w]
}
