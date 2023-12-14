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

import "errors"

var (
	// conn已经被关闭
	ErrClosed = errors.New("closed")

	ErrWrongStatusCode      = errors.New("Wrong status code")
	ErrUpgradeFieldValue    = errors.New("The value of the upgrade field is not 'websocket'")
	ErrConnectionFieldValue = errors.New("The value of the connection field is not 'upgrade'")
	ErrSecWebSocketAccept   = errors.New("The value of Sec-WebSocketAaccept field is invalid")

	ErrHostCannotBeEmpty   = errors.New("Host cannot be empty")
	ErrSecWebSocketKey     = errors.New("The value of SEC websocket key field is wrong")
	ErrSecWebSocketVersion = errors.New("The value of SEC websocket version field is wrong, not 13")

	ErrHTTPProtocolNotSupported = errors.New("HTTP protocol not supported")

	ErrOnlyGETSupported     = errors.New("error:Only get methods are supported")
	ErrMaxControlFrameSize  = errors.New("error:max control frame size > 125, need <= 125")
	ErrRsv123               = errors.New("error:rsv1 or rsv2 or rsv3 has a value")
	ErrOpcode               = errors.New("error:wrong opcode")
	ErrNOTBeFragmented      = errors.New("error:since control message MUST NOT be fragmented")
	ErrFrameOpcode          = errors.New("error:since all data frames after the initial data frame must have opcode 0.")
	ErrTextNotUTF8          = errors.New("error:text is not utf8 data")
	ErrClosePayloadTooSmall = errors.New("error:close payload too small")
	ErrCloseValue           = errors.New("error:close value is wrong") // close值不对
	ErrEmptyClose           = errors.New("error:close value is empty") // close的值是空的
	ErrWriteClosed          = errors.New("write close")
)
