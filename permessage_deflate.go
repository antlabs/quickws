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
	"net/http"
	"strconv"
)

// https://datatracker.ietf.org/doc/html/rfc7692#section-7.1
type permessageDeflate struct {
	// 是否启用
	enable bool
	// 解压缩
	decompression bool
	// 压缩
	compression bool
	// 服务端是否支持上下文接管
	// https://datatracker.ietf.org/doc/html/rfc7692#section-7.1.1.1
	// 客户端可以发送 server_no_context_takeover 参数，表示服务端不需要上下文接管
	serverContextTakeover bool
	// 客户端是否支持上下文接管
	// https://datatracker.ietf.org/doc/html/rfc7692#section-7.1.1.2
	// 客户端发关 client_no_context_takeover 参数，表示客户端不使用上下文接管
	// 即使服务端没有响应 client_no_context_takeover 参数，客户端也不会使用上下文接管
	clientContextTakeover bool

	// 客户端最大窗口位数， N=8-15, 窗口的大小2^N
	clientMaxWindowBits uint8
	// 服务端最大窗口位数， N=8-15, 窗口的大小2^N
	serverMaxWindowBits uint8
}

func parseMaxWindowBits(val string) (uint8, error) {
	if val == "" {
		return 15, nil
	}
	bits, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	if bits < 8 || bits > 15 {
		return 0, http.ErrNotSupported
	}
	return uint8(bits), nil
}

func parsePermessageDeflate(header http.Header) (pmd permessageDeflate, err error) {
	params := parseExtensions(header)
	pd := false
	for _, param := range params {
		switch param.key {
		case "permessage-deflate":
			pd = true
		case "server_no_context_takeover":
			if pd {
				pmd.serverContextTakeover = false
				pmd.enable = true
			}
		case "client_no_context_takeover":
			if pd {
				pmd.clientContextTakeover = false
				pmd.enable = true
			}
		case "client_max_window_bits":
			if pmd.clientMaxWindowBits, err = parseMaxWindowBits(param.val); err != nil {
				return
			}
			pmd.clientContextTakeover = true
		case "server_max_window_bits":
			if pmd.serverMaxWindowBits, err = parseMaxWindowBits(param.val); err != nil {
				return
			}
			pmd.serverContextTakeover = true
		default:
			err = http.ErrNotSupported
			return
		}
	}
	return
}

// 是否打开解压缩
func needDecompression(header http.Header) (pd permessageDeflate, err error) {
	return parsePermessageDeflate(header)
}

// 客户端用的函数
func maybeCompressionDecompression(header http.Header) (permessageDeflate, error) {
	return parsePermessageDeflate(header)
}
