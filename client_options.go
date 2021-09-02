// Copyright 2021 guonaihong. All rights reserved.
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

package tinyws

import (
	"crypto/tls"
	"net/http"
	"time"
)

type Option interface {
	apply(*DialOption)
}

type tlsConfig tls.Config

func (t *tlsConfig) apply(o *DialOption) {
	o.tlsConfig = (*tls.Config)(t)
}

// 配置tls.config
func WithTLSConfig(tls *tls.Config) Option {
	return (*tlsConfig)(tls)
}

type httpHeader http.Header

func (h httpHeader) apply(o *DialOption) {
	o.Header = (http.Header)(h)
}

// 配置http.Header
func WithHTTPHeader(h http.Header) Option {
	return (httpHeader)(h)
}

type timeDuration time.Duration

func (t timeDuration) apply(o *DialOption) {
	o.dialTimeout = (time.Duration)(t)
}

// 配置握手时的timeout
func WithDialTimeout(t time.Duration) Option {
	return (timeDuration)(t)
}

type replyPing bool

func (r replyPing) apply(o *DialOption) {
	o.replyPing = bool(r)
}

// 配置自动回应ping frame, 当收到ping， 回一个pong
func WithReplyPing() Option {
	return replyPing(true)
}

type decompression bool

func (c2 decompression) apply(o *DialOption) {
	o.decompression = bool(c2)
}

// 配置解压缩
func WithCompression() Option {
	return decompression(true)
}
