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
	"encoding/json"
	"time"
)

// 写入json至websocket连接
func (c *Conn) WriteJsonTimeout(op Opcode, v interface{}, t time.Duration) (err error) {
	all, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return c.WriteTimeout(op, all, t)
}

// 从websocket读取json
func (c *Conn) ReadJsonTimeout(t time.Duration, v interface{}) (op Opcode, err error) {
	all, op, err := c.ReadTimeout(t)
	if err != nil {
		return op, err
	}

	if err := json.Unmarshal(all, v); err != nil {
		return op, err
	}
	return op, nil
}
