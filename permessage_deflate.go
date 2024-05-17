package quickws

import (
	"strconv"
	"strings"

	"github.com/antlabs/wsutil/deflate"
)

var strExtensions = "permessage-deflate; server_no_context_takeover; client_no_context_takeover"

// 压缩的入口函数
func (c *Conn) encoode(payload *[]byte) (encodePayload *[]byte, err error) {

	ct := (c.pd.ClientContextTakeover && c.client || !c.client && c.pd.ServerContextTakeover) && c.Compression
	// 上下文接管
	bit := uint8(0)
	if c.client {
		bit = c.pd.ClientMaxWindowBits
	} else {
		bit = c.pd.ServerMaxWindowBits
	}
	if ct {
		// 这里的读取是单go程的。所以不用加锁
		if c.enCtx == nil {

			c.enCtx, err = deflate.NewCompressContextTakeover(bit)
			if err != nil {
				return nil, err
			}
		}

	}
	// 处理上下文接管和非上下文接管两种情况
	// bit 为啥放在参数里面传递, 因为非上下文接管的时候，也需要正确处理bit
	return c.enCtx.Compress(payload, bit)
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

	}

	// 上下文接管, deCtx是nil
	// 非上下文接管, deCtx是非nil
	return c.deCtx.Decompress(payload, c.readMaxMessage)
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
