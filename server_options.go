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

type OptionServer func(*ConnOption)

func WithServerReadTimeout(t time.Duration) OptionServer {
	return func(o *ConnOption) {
		o.readTimeout = t
	}
}

func WithServerCallback(cb Callback) OptionServer {
	return func(o *ConnOption) {
		o.Callback = cb
	}
}

// 配置自动回应ping frame, 当收到ping， 回一个pong
func WithServerReplyPing() OptionServer {
	return func(o *ConnOption) {
		o.replyPing = true
	}
}

// 配置解压缩
func WithServerDecompression() OptionServer {
	return func(o *ConnOption) {
		o.decompression = true
	}
}

// 配置压缩和解压缩
func WithServerDecompressAndCompress() OptionServer {
	return func(o *ConnOption) {
		o.compression = true
		o.decompression = true
	}
}

// 配置忽略pong消息
func WithServerIgnorePong() OptionServer {
	return func(o *ConnOption) {
		o.ignorePong = true
	}
}
