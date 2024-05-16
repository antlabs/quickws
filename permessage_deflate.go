package quickws

import (
	"strconv"
	"strings"

	"github.com/antlabs/wsutil/deflate"
)

var strExtensions = "permessage-deflate; server_no_context_takeover; client_no_context_takeover"

func (c *Conn) encoode(payload *[]byte) (encodePayload *[]byte, err error) {

	ct := (c.pd.ClientContextTakeover && c.client || !c.client && c.pd.ServerContextTakeover) && c.Compression
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
			c.enCtx, err = deflate.NewCompressContextTakeover(bit)
			if err != nil {
				return nil, err
			}
		}

		return c.enCtx.Compress(payload)
	}

	bit := uint8(0)
	if c.client {
		bit = c.pd.ClientMaxWindowBits
	} else {
		bit = c.pd.ServerMaxWindowBits
	}
	// 非上下文按管
	return deflate.CompressNoContextTakeover(payload, deflate.DefaultCompressionLevel, bit)
}

// 解压缩入口函数
func (c *Conn) decode(payload *[]byte) (decodePayload *[]byte, err error) {
	ct := (c.pd.ClientContextTakeover && c.client || !c.client && c.pd.ServerContextTakeover) && c.Decompression
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

		return c.deCtx.Decompress(payload, c.readMaxMessage)
	}

	// 非上下文按管
	return deflate.DecompressNoContextTakeover(payload)
}

func genSecWebSocketExtensions(pd deflate.PermessageDeflateConf) string {
	ext := make([]string, 1, 5)
	ext[0] = "permessage-deflate"
	if !pd.ClientContextTakeover {
		ext = append(ext, "client_no_context_takeover")
	}

	if !pd.ServerContextTakeover {
		ext = append(ext, "server_no_context_takeover")
	}

	if pd.ClientMaxWindowBits != 0 {
		ext = append(ext, "client_max_window_bits="+strconv.Itoa(int(pd.ClientMaxWindowBits)))
	}

	if pd.ServerMaxWindowBits != 0 {
		ext = append(ext, "server_max_window_bits="+strconv.Itoa(int(pd.ServerMaxWindowBits)))
	}
	return strings.Join(ext, "; ")
}
