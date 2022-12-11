// Copyright 2021-2022 antlabs. All rights reserved.
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

type Opcode uint8

const (
	Continuation Opcode = iota
	Text
	Binary
	// 3 - 7保留
	_ //3
	_
	_ //5
	_
	_ //7
	Close
	Ping
	Pong
)

var ErrClose = "websocket"

func (c Opcode) String() string {
	switch {
	case c >= 3 && c <= 7:
		return "control"
	case c == Text:
		return "text"
	case c == Binary:
		return "binary"
	case c == Close:
		return "close"
	case c == Ping:
		return "ping"
	case c == Pong:
		return "pong"
	case c == Continuation:
		return "continuation"
	default:
		return "unknown"
	}
}

func (c Opcode) isControl() bool {
	return (c & (1 << 3)) > 0
}
