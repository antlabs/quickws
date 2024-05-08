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
	"net/url"
	"time"
	"unicode/utf8"
)

// 0. CallbackFunc
func WithClientCallbackFunc(open OnOpenFunc, m OnMessageFunc, c OnCloseFunc) ClientOption {
	return func(o *DialOption) {
		o.Callback = &funcToCallback{
			onOpen:    open,
			onMessage: m,
			onClose:   c,
		}
	}
}

// 配置服务端回调函数
func WithServerCallbackFunc(open OnOpenFunc, m OnMessageFunc, c OnCloseFunc) ServerOption {
	return func(o *ConnOption) {
		o.Callback = &funcToCallback{
			onOpen:    open,
			onMessage: m,
			onClose:   c,
		}
	}
}

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
func WithServerEnableUTF8Check() ServerOption {
	return func(o *ConnOption) {
		o.utf8Check = utf8.Valid
	}
}

func WithClientEnableUTF8Check() ClientOption {
	return func(o *DialOption) {
		o.utf8Check = utf8.Valid
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

// 6 配置忽略pong消息
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

// 7.
// 设置几倍payload的缓冲区
// 只有解析方式是窗口的时候才有效
// 如果为1.0就是1024 + 14， 如果是2.0就是2048 + 14
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

// 8 配置windows解析方式
// 默认使用窗口解析方式, 以后以后默认解析方式改变过，才有必要使用这个选项
func WithServerWindowsParseMode() ServerOption {
	return func(o *ConnOption) {
		o.parseMode = ParseModeWindows
	}
}

// 默认使用窗口解析方式, 以后以后默认解析方式改变过，才有必要使用这个选项
func WithClientWindowsParseMode() ClientOption {
	return func(o *DialOption) {
		o.parseMode = ParseModeWindows
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

// 12 配置多倍payload缓冲区, 1.是1024 2。是2048
// 为何不让用户自己配置呢，可以和底层的buffer池结合起来，/1024就知道命中哪个缓冲区了, 不需要维护index命中的哪个sync.Pool
// 如果用户传些奇奇怪怪的数字，就不好办了
func WithServerBufioMultipleTimesPayloadSize(mt float32) ServerOption {
	return func(o *ConnOption) {
		if mt <= 0 {
			mt = 1.0
		}
		o.bufioMultipleTimesPayloadSize = mt
	}
}

func WithClientBufioMultipleTimesPayloadSize(mt float32) ClientOption {
	return func(o *DialOption) {
		if mt <= 0 {
			mt = 1.0
		}
		o.bufioMultipleTimesPayloadSize = mt
	}
}

// 13. 配置延迟发送
// 配置延迟最大发送时间
func WithServerMaxDelayWriteDuration(d time.Duration) ServerOption {
	return func(o *ConnOption) {
		o.maxDelayWriteDuration = d
	}
}

// 13. 配置延迟发送
// 配置延迟最大发送时间
func WithClientMaxDelayWriteDuration(d time.Duration) ClientOption {
	return func(o *DialOption) {
		o.maxDelayWriteDuration = d
	}
}

// 14.1 配置最大延迟个数.server
func WithServerMaxDelayWriteNum(n int32) ServerOption {
	return func(o *ConnOption) {
		o.maxDelayWriteNum = n
	}
}

// 14.2 配置最大延迟个数.client
func WithClientMaxDelayWriteNum(n int32) ClientOption {
	return func(o *DialOption) {
		o.maxDelayWriteNum = n
	}
}

// 15.1 配置延迟包的初始化buffer大小
func WithServerDelayWriteInitBufferSize(n int32) ServerOption {
	return func(o *ConnOption) {
		o.delayWriteInitBufferSize = n
	}
}

// 15.2 配置延迟包的初始化buffer大小
func WithClientDelayWriteInitBufferSize(n int32) ClientOption {
	return func(o *DialOption) {
		o.delayWriteInitBufferSize = n
	}
}

// 16. 配置读超时时间
//
// 16.1 .设置服务端读超时时间
func WithServerReadTimeout(t time.Duration) ServerOption {
	return func(o *ConnOption) {
		o.readTimeout = t
	}
}

// 16.2 .设置客户端读超时时间
func WithClientReadTimeout(t time.Duration) ClientOption {
	return func(o *DialOption) {
		o.readTimeout = t
	}
}

// 17。 只配置OnClose
// 17.1 配置服务端OnClose
func WithServerOnCloseFunc(onClose func(c *Conn, err error)) ServerOption {
	return func(o *ConnOption) {
		o.Callback = OnCloseFunc(onClose)
	}
}

// 17.2 配置客户端OnClose
func WithClientOnCloseFunc(onClose func(c *Conn, err error)) ClientOption {
	return func(o *DialOption) {
		o.Callback = OnCloseFunc(onClose)
	}
}

// 18. 配置新的dial函数, 这里可以配置socks5代理地址
func WithClientDialFunc(dialFunc func() (Dialer, error)) ClientOption {
	return func(o *DialOption) {
		o.dialFunc = dialFunc
	}
}

// 19. 配置proxy地址
func WithClientProxyFunc(proxyFunc func(*http.Request) (*url.URL, error)) ClientOption {
	return func(o *DialOption) {
		o.proxyFunc = proxyFunc
	}
}

// 20. 设置支持的子协议
// 20.1 设置客户端支持的子协议
func WithClientSubprotocols(subprotocols []string) ClientOption {
	return func(o *DialOption) {
		o.subProtocols = subprotocols
	}
}

// 20.2 设置服务端支持的子协议
func WithServerSubprotocols(subprotocols []string) ServerOption {
	return func(o *ConnOption) {
		o.subProtocols = subprotocols
	}
}

// 21.1 设置客户端支持上下文接管, 默认不支持上下文接管
func WithClientContextTakeover() ServerOption {
	return func(o *ConnOption) {
		o.clientContextTakeover = false
	}
}

// 21.2 设置服务端支持上下文接管, 默认不支持上下文接管
func WithServerContextTakeover() ServerOption {
	return func(o *ConnOption) {
		o.serverContextTakeover = false
	}
}

// 21.1 设置客户端最大窗口位数，使用上下文接管时，这个参数才有效
func WithClientMaxWindowsBits(bits uint8) ClientOption {
	return func(o *DialOption) {
		o.clientMaxWindowBits = bits
	}
}

// 22.2 设置服务端最大窗口位数, 使用上下文接管时，这个参数才有效
func WithServerMaxWindowBits(bits uint8) ServerOption {
	return func(o *ConnOption) {
		o.serverMaxWindowBits = bits
	}
}

// 22.1 设置客户端最大可以处理的message的大小, 默认没有限制
func WithClientMaxMessage(size int64) ClientOption {
	return func(o *DialOption) {
		o.maxMessage = size
	}
}

// 22.2 设置服务端最大可以处理的message的大小，默认没有限制
func WithServerMaxMessage(size int64) ServerOption {
	return func(o *ConnOption) {
		o.maxMessage = size
	}
}
