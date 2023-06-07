package quickws

import (
	"errors"
	"io"
)

var errNegativeRead = errors.New("bufio: reader returned negative count from Read")

// 固定大小的fixedReader, 所有的内存都是提前分配好的
// 标准库的bufio.Reader不能传递一个固定大小的buf, 导致控制力度会差点
type fixedReader struct {
	buf          []byte
	rd           io.Reader // reader provided by the client
	r, w         int       // buf read and write positions
	err          error
	lastByte     int // last byte read for UnreadByte; -1 means invalid
	lastRuneSize int // size of last rune read for UnreadRune; -1 means invalid
}

// newBuffer returns a new Buffer whose buffer has the specified size.
func newBuffer(r io.Reader, buf []byte) *fixedReader {
	return &fixedReader{
		rd:  r,
		buf: buf,
	}
}

func (b *fixedReader) readErr() error {
	err := b.err
	b.err = nil
	return err
}

// 将缓存区重置为一个新的buf
func (b *fixedReader) reset(buf []byte) {
	if len(buf) < len(b.buf[b.r:b.w]) {
		panic("new buf size is too small")
	}

	copy(buf, b.buf[b.r:b.w])
	b.w = b.w - b.r
	b.r = 0
	b.buf = buf
}

func (b *fixedReader) bytes() []byte {
	return b.buf
}

// 返回可写的缓存区
func (b *fixedReader) free() []byte {
	r := b.r
	copy(b.buf, b.buf[r:])
	b.w -= r
	b.r = 0
	return b.buf[b.w:]
}

func (b *fixedReader) availableBuf() *fixedReader {
	return &fixedReader{rd: b.rd, buf: b.buf[b.w:]}
}

// 返回剩余可用的缓存区大小
func (b *fixedReader) available() int64 {
	return int64(len(b.buf[b.w:]) + b.r)
}

func (b *fixedReader) Buffered() int { return b.w - b.r }

func (b *fixedReader) Read(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		if b.Buffered() > 0 {
			return 0, nil
		}
		return 0, b.readErr()
	}
	if b.r == b.w {
		if b.err != nil {
			return 0, b.readErr()
		}
		if len(p) >= len(b.buf) {
			// Large read, empty buffer.
			// Read directly into p to avoid copy.
			n, b.err = b.rd.Read(p)
			if n < 0 {
				panic(errNegativeRead)
			}
			if n > 0 {
				b.lastByte = int(p[n-1])
				b.lastRuneSize = -1
			}
			return n, b.readErr()
		}
		// One read.
		// Do not use b.fill, which will loop.
		b.r = 0
		b.w = 0
		n, b.err = b.rd.Read(b.buf)
		if n < 0 {
			panic(errNegativeRead)
		}
		if n == 0 {
			return 0, b.readErr()
		}
		b.w += n
	}

	// copy as much as we can
	// Note: if the slice panics here, it is probably because
	// the underlying reader returned a bad count. See issue 49795.
	n = copy(p, b.buf[b.r:b.w])
	b.r += n
	b.lastByte = int(b.buf[b.r-1])
	b.lastRuneSize = -1
	return n, nil
}
