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

package quickws

import (
	"encoding/binary"
	"strconv"
	"strings"
)

// https://datatracker.ietf.org/doc/html/rfc6455#section-7.4.1
// 这里记录了各种状态码的含义
type StatusCode int16

const (
	// NormalClosure 正常关闭
	NormalClosure StatusCode = 1000
	// EndpointGoingAway 对端正在消失
	EndpointGoingAway StatusCode = 1001
	// ProtocolError 表示对端由于协议错误正在终止连接
	ProtocolError StatusCode = 1002
	// DataCannotAccept 收到一个不能接受的数据类型
	DataCannotAccept StatusCode = 1003
	// NotConsistentMessageType 表示对端正在终止连接, 消息类型不一致
	NotConsistentMessageType StatusCode = 1007
	// TerminatingConnection 表示对端正在终止连接, 没有好用的错误, 可以用这个错误码表示
	TerminatingConnection StatusCode = 1008
	// TooBigMessage  消息太大, 不能处理, 关闭连接
	TooBigMessage StatusCode = 1009
	// NoExtensions 只用于客户端, 服务端返回扩展消息
	NoExtensions StatusCode = 1010
	// ServerTerminating 服务端遇到意外情况, 中止请求
	ServerTerminating StatusCode = 1011
)

func (s StatusCode) String() string {
	switch s {
	case NormalClosure:
		return "NormalClosure"
	case EndpointGoingAway:
		return "EndpointGoingAway"
	case ProtocolError:
		return "ProtocolError"
	case DataCannotAccept:
		return "DataCannotAccept"
	case NotConsistentMessageType:
		return "NotConsistentMessageType"
	case TerminatingConnection:
		return "TerminatingConnection"
	case TooBigMessage:
		return "TooBigMessage"
	case NoExtensions:
		return "NoExtensions"
	case ServerTerminating:
		return "ServerTerminating"
	}

	return "unkown"
}

type CloseErrMsg struct {
	Code StatusCode
	Msg  string
}

func (c CloseErrMsg) Error() string {
	var out strings.Builder

	out.WriteString("<quickws close: code:")

	out.WriteString(strconv.Itoa(int(c.Code)))

	out.WriteString(" msg:")

	out.WriteString(c.Code.String())

	if len(c.Msg) > 0 {
		out.WriteString(c.Msg)
	}

	out.WriteString(">")
	return out.String()
}

func bytesToCloseErrMsg(payload []byte) *CloseErrMsg {
	var ce CloseErrMsg
	if len(payload) >= 2 {
		ce.Code = StatusCode(binary.BigEndian.Uint16(payload))
	}

	if len(payload) >= 3 {
		ce.Msg = string(payload[3:])
	}
	return &ce
}

func statusCodeToBytes(code StatusCode) (rv []byte) {
	rv = make([]byte, 2+len(code.String()))
	binary.BigEndian.PutUint16(rv, uint16(code))
	copy(rv[2:], code.String())
	return
}

func validCode(code uint16) bool {
	switch code {
	case 1004, 1005, 1006, 1015:
		return false
	}

	if code >= 1000 && code <= 1015 {
		return true
	}

	if code >= 3000 && code <= 4999 {
		return true
	}

	return false
}
