// Copyright 2021-2023 antlabs. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package quickws

import "time"

type ServerOption func(*ConnOption)

// 设置TCP_NODELAY
func WithServerTCPDelay() ServerOption {
	return func(o *ConnOption) {
		var b bool
		o.tcpNoDelay = &b
	}
}

// 关闭utf8检查
func WithServerDisableUTF8Check() ServerOption {
	return func(o *ConnOption) {
		o.utf8Check = func([]byte) bool { return true }
	}
}

func WithServerReadTimeout(t time.Duration) ServerOption {
	return func(o *ConnOption) {
		o.readTimeout = t
	}
}

func WithServerCallback(cb Callback) ServerOption {
	return func(o *ConnOption) {
		o.Callback = cb
	}
}

// 仅仅配置OnMessae函数
func WithServerOnMessageFunc(cb func(*Conn, Opcode, []byte)) ClientOption {
	return func(o *DialOption) {
		o.Callback = OnMessageFunc(cb)
	}
}

// 配置自动回应ping frame, 当收到ping， 回一个pong
func WithServerReplyPing() ServerOption {
	return func(o *ConnOption) {
		o.replyPing = true
	}
}

// 配置解压缩
func WithServerDecompression() ServerOption {
	return func(o *ConnOption) {
		o.decompression = true
	}
}

// 配置压缩和解压缩
func WithServerDecompressAndCompress() ServerOption {
	return func(o *ConnOption) {
		o.compression = true
		o.decompression = true
	}
}

// 配置忽略pong消息
func WithServerIgnorePong() ServerOption {
	return func(o *ConnOption) {
		o.ignorePong = true
	}
}

// 设置几倍payload的缓冲区
// 只有解析方式是窗口的时候才有效
func WithWindowsMultipleTimesPayloadSize(mt float32) ServerOption {
	return func(o *ConnOption) {
		// 如果mt < 1.0, 直接panic
		if mt < 1.0 {
			panic("multipleTimesPayloadSize must >= 1.0")
		}
		o.windowsMultipleTimesPayloadSize = mt
	}
}

// 默认使用窗口解析方式
func WithWindowsParseMode() ServerOption {
	return func(o *ConnOption) {
		o.parseMode = ParseModeWindows
	}
}

// 使用基于bufio的解析方式
func WithBufioParseMode() ServerOption {
	return func(o *ConnOption) {
		o.parseMode = ParseModeBufio
	}
}
