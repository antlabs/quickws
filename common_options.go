// Copyright 2021-2023 antlabs. All rights reserved.
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

// 1. callback
// 配置客户端callback
func WithClientCallback(cb Callback) ClientOption {
	return func(o *DialOption) {
		o.Callback = cb
	}
}

// 配置服务端回调函数
func WithServerCallback(cb Callback) ServerOption {
	return func(o *ConnOption) {
		o.Callback = cb
	}
}

// 2. 设置TCP_NODELAY
// 设置客户端TCP_NODELAY
func WithClientTCPDelay() ClientOption {
	return func(o *DialOption) {
		o.tcpNoDelay = false
	}
}

// 设置TCP_NODELAY 为false, 开启nagle算法
// 设置服务端TCP_NODELAY
func WithServerTCPDelay() ServerOption {
	return func(o *ConnOption) {
		o.tcpNoDelay = false
	}
}

// 3.关闭utf8检查
func WithServerDisableUTF8Check() ServerOption {
	return func(o *ConnOption) {
		o.utf8Check = func([]byte) bool { return true }
	}
}

func WithClientDisableUTF8Check() ServerOption {
	return func(o *ConnOption) {
		o.utf8Check = func([]byte) bool { return true }
	}
}

// 4.仅仅配置OnMessae函数
// 仅仅配置OnMessae函数
func WithServerOnMessageFunc(cb OnMessageFunc) ServerOption {
	return func(o *ConnOption) {
		o.Callback = OnMessageFunc(cb)
	}
}

// 仅仅配置OnMessae函数
func WithClientOnMessageFunc(cb OnMessageFunc) ClientOption {
	return func(o *DialOption) {
		o.Callback = OnMessageFunc(cb)
	}
}

// 5.
// 配置自动回应ping frame, 当收到ping， 回一个pong
func WithServerReplyPing() ServerOption {
	return func(o *ConnOption) {
		o.replyPing = true
	}
}

// 配置自动回应ping frame, 当收到ping， 回一个pong
func WithClientReplyPing() ClientOption {
	return func(o *DialOption) {
		o.replyPing = true
	}
}

// 6.
// 设置几倍payload的缓冲区
// 只有解析方式是窗口的时候才有效
func WithServerWindowsMultipleTimesPayloadSize(mt float32) ServerOption {
	return func(o *ConnOption) {
		if mt < 1.0 {
			mt = 1.0
		}
		o.windowsMultipleTimesPayloadSize = mt
	}
}

func WithClientWindowsMultipleTimesPayloadSize(mt float32) ClientOption {
	return func(o *DialOption) {
		if mt < 1.0 {
			mt = 1.0
		}
		o.windowsMultipleTimesPayloadSize = mt
	}
}

// 7
// 默认使用窗口解析方式
func WithServerWindowsParseMode() ServerOption {
	return func(o *ConnOption) {
		o.parseMode = ParseModeWindows
	}
}

func WithClientWindowsParseMode() ClientOption {
	return func(o *DialOption) {
		o.parseMode = ParseModeWindows
	}
}

// 8 配置忽略pong消息
func WithClientIgnorePong() ClientOption {
	return func(o *DialOption) {
		o.ignorePong = true
	}
}

func WithServerIgnorePong() ServerOption {
	return func(o *ConnOption) {
		o.ignorePong = true
	}
}

//	9.
//
// 使用基于bufio的解析方式
func WithServerBufioParseMode() ServerOption {
	return func(o *ConnOption) {
		o.parseMode = ParseModeBufio
	}
}

func WithClientBufioParseMode() ClientOption {
	return func(o *DialOption) {
		o.parseMode = ParseModeBufio
	}
}

// 10 配置解压缩
func WithClientDecompression() ClientOption {
	return func(o *DialOption) {
		o.decompression = true
	}
}

func WithServerDecompression() ServerOption {
	return func(o *ConnOption) {
		o.decompression = true
	}
}

// 11 关闭bufio clear hack优化
func WithServerDisableBufioClearHack() ServerOption {
	return func(o *ConnOption) {
		o.disableBufioClearHack = true
	}
}

func WithClientDisableBufioClearHack() ClientOption {
	return func(o *DialOption) {
		o.disableBufioClearHack = true
	}
}
