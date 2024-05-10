// Copyright 2021-2024 antlabs. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package quickws

import (
	"bytes"
	"io"
	"unsafe"

	"github.com/antlabs/wsutil/bytespool"
	"github.com/klauspost/compress/flate"
)

var tailBytes = []byte{0x00, 0x00, 0xff, 0xff, 0x01, 0x00, 0x00, 0xff, 0xff}

// 无上下文-解压缩
func decompressNoContextTakeover(r io.Reader) io.ReadCloser {
	fr, _ := flateReaderPool.Get().(io.ReadCloser)
	fr.(flate.Resetter).Reset(io.MultiReader(r, bytes.NewReader(tailBytes)), nil)
	return &flateReadWrapper{fr}
}

// 无上下文-解压缩
func decodeNoTontext(payload []byte) (*[]byte, error) {
	pr := bytes.NewReader(payload)
	r := decompressNoContextTakeover(pr)

	// 从池里面拿buf, 这里的2是经验值，解压缩之后是2倍的大小
	decodeBuf := bytespool.GetBytes(len(payload) * 2)
	// 包装下
	out := bytes.NewBuffer((*decodeBuf)[:0])
	// 解压缩
	if _, err := io.Copy(out, r); err != nil {
		return nil, err
	}
	// 拿到解压缩之后的buf
	outBytes := out.Bytes()
	// 如果解压缩之后的buf和从池里面拿的buf不一样，就把从池里面拿的buf放回去
	if unsafe.SliceData(*decodeBuf) != unsafe.SliceData(outBytes) {
		bytespool.PutBytes(decodeBuf)
	}

	r.Close()
	return &outBytes, nil
}

// 上下文-解压缩
type deCompressContextTakeover struct {
	dict historyDict
	io.ReadCloser
}

// 初始化一个对象
func newDecompressContextTakeover(bit uint8) (*deCompressContextTakeover, error) {
	size := 1 << uint(bit)
	r := flate.NewReader(nil)
	de := &deCompressContextTakeover{
		ReadCloser: r,
	}
	de.dict.InitHistoryDict(size)
	return de, nil
}

// 解压缩
func (d *deCompressContextTakeover) decompress(payload []byte) (*[]byte, error) {
	// 获取dict
	dict := d.dict.GetData()
	// 拿到接口
	frt := d.ReadCloser.(flate.Resetter)
	// 重置
	frt.Reset(io.MultiReader(bytes.NewReader(payload), bytes.NewReader(tailBytes)), dict)
	// 从池里面拿buf, 这里的2是经验值，解压缩之后是2倍的大小
	decodeBuf := bytespool.GetBytes(len(payload) * 2)
	// 包装下
	out := bytes.NewBuffer((*decodeBuf)[:0])
	// 解压缩
	if _, err := io.Copy(out, d.ReadCloser); err != nil {
		return nil, err
	}
	// 拿到解压缩之后的buf
	outBytes := out.Bytes()
	// 如果解压缩之后的buf和从池里面拿的buf不一样，就把从池里面拿的buf放回去
	if unsafe.SliceData(*decodeBuf) != unsafe.SliceData(outBytes) {
		bytespool.PutBytes(decodeBuf)
	}
	// 写入dict
	d.dict.Write(out.Bytes())
	// 返回解压缩之后的buf
	return &outBytes, nil

}

// 解压缩入口函数
func (c *Conn) decode(payload []byte) (decodePayload *[]byte, err error) {
	ct := (c.pd.clientContextTakeover && c.client || !c.client && c.pd.serverContextTakeover) && c.decompression
	// 上下文接管
	if ct {
		// 这里的读取是单go程的。所以不用加锁
		if c.deCtx == nil {

			bit := uint8(0)
			if c.client {
				bit = c.pd.clientMaxWindowBits
			} else {
				bit = c.pd.serverMaxWindowBits
			}
			c.deCtx, err = newDecompressContextTakeover(bit)
			if err != nil {
				return nil, err
			}
		}

		return c.deCtx.decompress(payload)
	}

	// 非上下文按管
	return decodeNoTontext(payload)
}
