// Copyright 2017 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package quickws

import (
	"bytes"
	"compress/flate"
	"io"
	"sync"
	"unsafe"

	"github.com/antlabs/wsutil/bytespool"
	"github.com/antlabs/wsutil/enum"
)

const (
	minCompressionLevel     = -2 // flate.HuffmanOnly not defined in Go < 1.6
	maxCompressionLevel     = flate.BestCompression
	defaultCompressionLevel = 1
)

var (
	flateWriterPools [maxCompressionLevel - minCompressionLevel + 1]sync.Pool
	flateReaderPool  = sync.Pool{New: func() interface{} {
		return flate.NewReader(nil)
	}}
)

/*
func isValidCompressionLevel(level int) bool {
	return minCompressionLevel <= level && level <= maxCompressionLevel
}
*/

func compressNoContextTakeoverInner(w io.WriteCloser, level int) io.WriteCloser {
	p := &flateWriterPools[level-minCompressionLevel]
	fw, _ := p.Get().(*flate.Writer)
	if fw == nil {
		fw, _ = flate.NewWriter(w, level)
	} else {
		fw.Reset(w)
	}
	return &flateWriteWrapper{fw: fw, p: p}
}

func compressNoContextTakeover(payload []byte, level int) (encodeBuf *[]byte, err error) {

	encodeBuf = bytespool.GetBytes(len(payload) + enum.MaxFrameHeaderSize)

	out := wrapBuffer{Buffer: bytes.NewBuffer((*encodeBuf)[:0])}
	w := compressNoContextTakeoverInner(&out, defaultCompressionLevel)
	if _, err = io.Copy(w, bytes.NewReader(payload)); err != nil {
		return nil, err
	}

	if err = w.Close(); err != nil {
		return nil, err
	}

	if unsafe.SliceData(*encodeBuf) != unsafe.SliceData(out.Buffer.Bytes()) {
		bytespool.PutBytes(encodeBuf)
	}

	outBuf := out.Bytes()
	return &outBuf, nil
}

type flateWriteWrapper struct {
	fw *flate.Writer
	p  *sync.Pool
}

func (w *flateWriteWrapper) Write(p []byte) (int, error) {
	if w.fw == nil {
		return 0, ErrWriteClosed
	}
	return w.fw.Write(p)
}

func (w *flateWriteWrapper) Close() error {
	if w.fw == nil {
		return ErrWriteClosed
	}
	err1 := w.fw.Flush()
	w.p.Put(w.fw)
	w.fw = nil
	// if w.tw.p != [4]byte{0, 0, 0xff, 0xff} {
	// 	return errors.New("websocket: internal error, unexpected bytes at end of flate stream")
	// }
	err2 := w.fw.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
