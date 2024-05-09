// Copyright 2017 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package quickws

import (
	"bytes"
	"io"

	"github.com/antlabs/wsutil/bytespool"
	"github.com/klauspost/compress/flate"
)

var tailBytes = []byte{0x00, 0x00, 0xff, 0xff, 0x01, 0x00, 0x00, 0xff, 0xff}

type flateReadWrapper struct {
	fr io.ReadCloser
}

func (r *flateReadWrapper) Read(p []byte) (int, error) {
	if r.fr == nil {
		return 0, io.ErrClosedPipe
	}
	n, err := r.fr.Read(p)
	if err == io.EOF {
		// Preemptively place the reader back in the pool. This helps with
		// scenarios where the application does not call NextReader() soon after
		// this final read.
		r.Close()
	}
	return n, err
}

func (r *flateReadWrapper) Close() error {
	if r.fr == nil {
		return io.ErrClosedPipe
	}
	err := r.fr.Close()
	flateReaderPool.Put(r.fr)
	r.fr = nil
	return err
}

func decompressNoContextTakeover(r io.Reader) io.ReadCloser {
	fr, _ := flateReaderPool.Get().(io.ReadCloser)
	fr.(flate.Resetter).Reset(io.MultiReader(r, bytes.NewReader(tailBytes)), nil)
	return &flateReadWrapper{fr}
}

func decode(payload []byte) ([]byte, error) {
	r := bytes.NewReader(payload)
	r2 := decompressNoContextTakeover(r)
	var o bytes.Buffer
	// TODO:, 从池里面拿数据
	if _, err := io.Copy(&o, r2); err != nil {
		return nil, err
	}
	r2.Close()
	return o.Bytes(), nil
}

// 上下文-解压缩
type decompressContextTakeover struct {
	dict historyDict
	io.ReadCloser
}

// 初始化一个对象
func newDecompressContextTakeover(bit int) (*decompressContextTakeover, error) {
	size := 1 << uint(bit)
	r := flate.NewReader(nil)
	return &decompressContextTakeover{
		dict:       *NewHistoryDict(size),
		ReadCloser: r,
	}, nil
}

// 解压缩
func (d *decompressContextTakeover) decompress(payload []byte) ([]byte, error) {
	// 获取dict
	dict := d.dict.GetData()
	// 拿到接口
	frt := d.ReadCloser.(flate.Resetter)
	// 重置
	frt.Reset(io.MultiReader(bytes.NewReader(payload), bytes.NewReader(tailBytes)), dict)
	// 从池里面拿buf, 这里的2是经验值，解压缩之后是2倍的大小
	decodeBuf := bytespool.GetBytes(len(payload) * 2)
	// TODO 包装下
	out := bytes.NewBuffer((*decodeBuf)[:0])
	//
	io.Copy(out, d.ReadCloser)
	//
	d.dict.Write(out.Bytes())
	//
	return out.Bytes(), nil

}
