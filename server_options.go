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

type ServerOption func(*ConnOption)

// 1.配置压缩和解压缩
func WithServerDecompressAndCompress() ServerOption {
	return func(o *ConnOption) {
		o.compression = true
		o.decompression = true
	}
}

// 2. 设置服务端支持的子协议
func WithServerSubprotocols(subprotocols []string) ServerOption {
	return func(o *ConnOption) {
		o.subProtocols = subprotocols
	}
}
