package quickws

import (
	"errors"
	"io"
)

var errNegativeRead = errors.New("bufio: reader returned negative count from Read")

// 固定大小的fixedReader, 所有的内存都是提前分配好的
// 标准库的bufio.Reader不能传递一个固定大小的buf, 导致控制力度会差点
type fixedReader struct {
	buf  []byte
	p    *[]byte
	rd   io.Reader // reader provided by the client
	r, w int       // buf read and write positions
	err  error
}

// newBuffer returns a new Buffer whose buffer has the specified size.
func newBuffer(r io.Reader, buf *[]byte) *fixedReader {
	return &fixedReader{
		rd:  r,
		buf: *buf,
		p:   buf,
	}
}

func (b *fixedReader) release() error {
	if b.p != nil {
		putBytes(b.p)
		b.buf = nil
		b.p = nil
	}
	return nil
}

func (b *fixedReader) readErr() error {
	err := b.err
	b.err = nil
	return err
}

// 将缓存区重置为一个新的buf
func (b *fixedReader) reset(buf *[]byte) {
	if len(*buf) < len(b.buf[b.r:b.w]) {
		panic("new buf size is too small")
	}

	copy(*buf, b.buf[b.r:b.w])
	b.w -= b.r
	b.r = 0
	b.p = buf
	b.buf = *buf
}

// 返回底层[]byte的长度
func (b *fixedReader) Len() int {
	return len(b.buf)
}

func (b *fixedReader) ptr() *[]byte {
	return b.p
}

func (b *fixedReader) bytes() []byte {
	return b.buf
}

// 返回剩余可写的缓存区大小
func (b *fixedReader) writeCap() int {
	return len(b.buf[b.w:])
}

// 返回剩余可用的缓存区大小
func (b *fixedReader) available() int64 {
	return int64(len(b.buf[b.w:]) + b.r)
}

// 左移缓存区
func (b *fixedReader) leftMove() {
	if b.r == 0 {
		return
	}
	copy(b.buf, b.buf[b.r:b.w])
	b.w -= b.r
	b.r = 0
}

// 返回可写的缓存区
func (b *fixedReader) writeCapBytes() []byte {
	return b.buf[b.w:]
}

func (b *fixedReader) cloneAvailable() *fixedReader {
	return &fixedReader{rd: b.rd, buf: b.buf[b.w:]}
}

func (b *fixedReader) Buffered() int { return b.w - b.r }

// 这和一般read接口中不一样
// 传入的p 一定会满足这个大小
func (b *fixedReader) Read(p []byte) (n int, err error) {
	if cap(b.buf) < cap(p) {
		panic("fixedReader.Reader buf size is too small: cap(b.buf) < cap(p)")
	}

	n = len(p)
	if n == 0 {
		if b.Buffered() > 0 {
			return 0, nil
		}
		return 0, b.readErr()
	}

	var n1 int
	for {

		if b.r == b.w || len(b.buf[b.r:b.w]) < len(p) {
			if b.err != nil {
				return 0, b.readErr()
			}
			if b.r == b.w {
				b.r = 0
				b.w = 0
			}
			n1, b.err = b.rd.Read(b.buf[b.w:])
			if n1 < 0 {
				panic(errNegativeRead)
			}
			if n1 == 0 {
				return 0, b.readErr()
			}
			b.w += n1
		}

		if len(b.buf[b.r:b.w]) < len(p) {
			continue
		}
		n1 = copy(p, b.buf[b.r:b.w])
		b.r += n1

		return n, nil
	}
}
