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

type ClientOption func(*DialOption)

// 配置tls.config
func WithClientTLSConfig(tls *tls.Config) ClientOption {
	return func(o *DialOption) {
		o.tlsConfig = tls
	}
}

// 配置http.Header
func WithClientHTTPHeader(h http.Header) ClientOption {
	return func(o *DialOption) {
		o.Header = h
	}
}

// 配置握手时的timeout
func WithClientDialTimeout(t time.Duration) ClientOption {
	return func(o *DialOption) {
		o.dialTimeout = t
	}
}

// 配置压缩
func WithClientCompression() ClientOption {
	return func(o *DialOption) {
		o.compression = true
	}
}

// 配置压缩和解压缩
func WithClientDecompressAndCompress() ClientOption {
	return func(o *DialOption) {
		o.compression = true
		o.decompression = true
	}
}
