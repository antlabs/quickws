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
	"sync/atomic"
	"unsafe"

	"github.com/antlabs/wsutil/deflate"
)

// 压缩的入口函数
func (c *Conn) encoode(payload *[]byte) (encodePayload *[]byte, err error) {

	ct := (c.pd.ClientContextTakeover && c.client || !c.client && c.pd.ServerContextTakeover) && c.pd.Compression
	// 上下文接管
	bit := uint8(0)
	if c.client {
		bit = c.pd.ClientMaxWindowBits
	} else {
		bit = c.pd.ServerMaxWindowBits
	}
	if ct {
		// 这里的读取是单go程的。所以不用加锁
		if atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&c.enCtx))) == nil {

			c.wmu.Lock() //加锁
			if atomic.LoadPointer((*unsafe.Pointer)((unsafe.Pointer)(&c.enCtx))) != nil {
				goto decode
			}
			enCtx, err := deflate.NewCompressContextTakeover(bit)
			if err != nil {
				c.wmu.Unlock()
				return nil, err
			}
			atomic.StorePointer((*unsafe.Pointer)((unsafe.Pointer)(&c.enCtx)), unsafe.Pointer(enCtx))
			c.wmu.Unlock()
		}

	}
	c.wmu.Lock()
decode:
	defer c.wmu.Unlock()
	// 处理上下文接管和非上下文接管两种情况
	// bit 为啥放在参数里面传递, 因为非上下文接管的时候，也需要正确处理bit
	return c.enCtx.Compress(payload, bit)
}

// 解压缩入口函数
// 解压目前只在一个go程里面按序列处理，所以不需要加锁
func (c *Conn) decode(payload *[]byte) (decodePayload *[]byte, err error) {
	ct := (c.pd.ClientContextTakeover && c.client || !c.client && c.pd.ServerContextTakeover) && c.pd.Decompression
	// 上下文接管
	if ct {
		if c.deCtx == nil {

			bit := uint8(0)

			if c.client {
				bit = c.pd.ClientMaxWindowBits
			} else {
				bit = c.pd.ServerMaxWindowBits
			}
			c.deCtx, err = deflate.NewDecompressContextTakeover(bit)
			if err != nil {
				return nil, err
			}
		}

	}

	// 上下文接管, deCtx是nil
	// 非上下文接管, deCtx是非nil

	return c.deCtx.Decompress(payload, c.readMaxMessage)
}
