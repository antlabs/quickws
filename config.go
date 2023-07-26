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
	"time"
	"unicode/utf8"

	"github.com/antlabs/wsutil/enum"
)

type Config struct {
	Callback
	tcpNoDelay                      bool
	replyPing                       bool              // 开启自动回复
	decompression                   bool              // 开启解压缩功能
	compression                     bool              // 开启压缩功能
	ignorePong                      bool              // 忽略pong消息
	disableBufioClearHack           bool              // 关闭bufio的clear hack优化
	utf8Check                       func([]byte) bool // utf8检查
	readTimeout                     time.Duration
	windowsMultipleTimesPayloadSize float32   // 设置几倍的payload大小
	parseMode                       parseMode // 解析模式
}

func (c *Config) initPayloadSize() int {
	return int(1024.0 + float32(enum.MaxFrameHeaderSize)*c.windowsMultipleTimesPayloadSize)
}

// 默认设置
func (c *Config) defaultSetting() {
	c.windowsMultipleTimesPayloadSize = 1.0
	c.tcpNoDelay = true
	c.parseMode = ParseModeWindows
	c.utf8Check = utf8.Valid
}
