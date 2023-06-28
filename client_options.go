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

import (
	"crypto/tls"
	"net/http"
	"time"
)

type OptionClient func(*DialOption)

// 配置callback
func WithClientCallback(cb Callback) OptionClient {
	return func(o *DialOption) {
		o.Callback = cb
	}
}

// 仅仅配置OnMessae函数
func WithClientOnMessageFunc(cb OnMessageFunc) OptionClient {
	return func(o *DialOption) {
		o.Callback = OnMessageFunc(cb)
	}
}

// 配置tls.config
func WithTLSConfig(tls *tls.Config) OptionClient {
	return func(o *DialOption) {
		o.tlsConfig = tls
	}
}

// 配置http.Header
func WithHTTPHeader(h http.Header) OptionClient {
	return func(o *DialOption) {
		o.Header = h
	}
}

// 配置握手时的timeout
func WithDialTimeout(t time.Duration) OptionClient {
	return func(o *DialOption) {
		o.dialTimeout = t
	}
}

// 配置自动回应ping frame, 当收到ping， 回一个pong
func WithReplyPing() OptionClient {
	return func(o *DialOption) {
		o.replyPing = true
	}
}

// 配置解压缩
func WithDecompression() OptionClient {
	return func(o *DialOption) {
		o.decompression = true
	}
}

// 配置压缩
func WithCompression() OptionClient {
	return func(o *DialOption) {
		o.compression = true
	}
}

// 配置压缩和解压缩
func WithDecompressAndCompress() OptionClient {
	return func(o *DialOption) {
		o.compression = true
		o.decompression = true
	}
}
