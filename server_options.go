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

type ServerOption interface {
	apply(*ConnOption)
}

type serverReplyPing bool

func (r serverReplyPing) apply(o *ConnOption) {
	o.replyPing = bool(r)
}

// 配置自动回应ping frame, 当收到ping， 回一个pong
func WithServerReplyPing() ServerOption {
	return serverReplyPing(true)
}

type compression bool

func (c2 compression) apply(c *ConnOption) {
	c.compression = bool(c2)
}

// 配置压缩
func WithCompression() ServerOption {
	return compression(true)
}
