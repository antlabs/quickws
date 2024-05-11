package quickws

import "github.com/antlabs/wsutil/deflate"

func (c *Conn) encoode(payload []byte) (encodePayload *[]byte, err error) {

	ct := (c.pd.ClientContextTakeover && c.client || !c.client && c.pd.ServerContextTakeover) && c.compression
	// 上下文接管
	if ct {
		// 这里的读取是单go程的。所以不用加锁
		if c.enCtx == nil {

			bit := uint8(0)
			if c.client {
				bit = c.pd.ClientMaxWindowBits
			} else {
				bit = c.pd.ServerMaxWindowBits
			}
			c.enCtx, err = deflate.NewEncompressContextTakeover(bit)
			if err != nil {
				return nil, err
			}
		}

		return c.enCtx.Compress(payload)
	}

	// 非上下文按管
	return deflate.CompressNoContextTakeover(payload, deflate.DefaultCompressionLevel)
}

// 解压缩入口函数
func (c *Conn) decode(payload []byte) (decodePayload *[]byte, err error) {
	ct := (c.pd.ClientContextTakeover && c.client || !c.client && c.pd.ServerContextTakeover) && c.Decompression
	// 上下文接管
	if ct {
		// 这里的读取是单go程的。所以不用加锁
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

		return c.deCtx.Decompress(payload, c.readMaxMessage)
	}

	// 非上下文按管
	return deflate.DecompressNoContextTakeover(payload)
}
