// Copyright 2021-2024 antlabs. All rights reserved.
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
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/antlabs/wsutil/enum"
)

var ErrDialFuncAndProxyFunc = errors.New("dialFunc and proxyFunc can't be set at the same time")

type Dialer interface {
	Dial(network, addr string) (c net.Conn, err error)
}

// Config的配置，有两个种用法
// 一种是声明一个全局的配置，后面不停使用。
// 另外一种是局部声明一个配置，然后使用WithXXX函数设置配置
type Config struct {
	Callback
	permessageDeflate               // 静态配置, 从WithXXX函数中获取
	tcpNoDelay                      bool
	replyPing                       bool              // 开启自动回复
	decompression                   bool              // 开启解压缩功能
	compression                     bool              // 开启压缩功能
	ignorePong                      bool              // 忽略pong消息
	disableBufioClearHack           bool              // 关闭bufio的clear hack优化
	utf8Check                       func([]byte) bool // utf8检查
	readTimeout                     time.Duration     // 读超时时间
	windowsMultipleTimesPayloadSize float32           // 设置几倍(1024+14)的payload大小
	bufioMultipleTimesPayloadSize   float32           // 设置几倍(1024)的payload大小
	parseMode                       parseMode         // 解析模式
	maxDelayWriteNum                int32             // 最大延迟包的个数, 默认值为10
	delayWriteInitBufferSize        int32             // 延迟写入的初始缓冲区大小, 默认值是8k
	maxDelayWriteDuration           time.Duration     // 最大延迟时间, 默认值是10ms
	subProtocols                    []string          // 设置支持的子协议
	maxMessage                      int64             //最大消息大小
	dialFunc                        func() (Dialer, error)
	proxyFunc                       func(*http.Request) (*url.URL, error) //
}

func (c *Config) initPayloadSize() int {
	return int((1024.0 + float32(enum.MaxFrameHeaderSize)) * c.windowsMultipleTimesPayloadSize)
}

// 默认设置
func (c *Config) defaultSetting() error {
	c.Callback = &DefCallback{}
	c.maxDelayWriteNum = 10
	c.windowsMultipleTimesPayloadSize = 1.0
	c.delayWriteInitBufferSize = 8 * 1024
	c.maxDelayWriteDuration = 10 * time.Millisecond
	c.tcpNoDelay = true
	c.parseMode = ParseModeWindows
	// 对于text消息，默认不检查text是utf8字符
	c.utf8Check = func(b []byte) bool { return true }

	if c.dialFunc != nil && c.proxyFunc != nil {
		return ErrDialFuncAndProxyFunc
	}
	return nil
}
