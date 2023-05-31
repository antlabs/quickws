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

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrNotFoundHijacker = errors.New("not found Hijacker")
	strHeaderUpgrade    = "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n"
	strHeaderExtensions = "Sec-WebSocket-Extensions: permessage-deflate; server_no_context_takeover; client_no_context_takeover\r\n"
	strCRLF             = "\r\n"
)

type ConnOption struct {
	config
}

func Upgrade(w http.ResponseWriter, r *http.Request, opts ...OptionServer) (c *Conn, err error) {
	var conf ConnOption
	for _, o := range opts {
		o(&conf)
	}

	if ecode, err := checkRequest(r); err != nil {
		http.Error(w, err.Error(), ecode)
		return nil, err
	}

	hi, ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrNotFoundHijacker
	}

	conn, rw, err := hi.Hijack()
	if err != nil {
		return nil, err
	}

	// 是否打开解压缩
	// 外层接收压缩, 并且客户端发送扩展过来
	if conf.decompression {
		conf.decompression = needDecompression(r.Header)
	}

	if err = writeResponse(r, rw.Writer, conf.config); err != nil {
		return
	}

	return newConn(conn, rw, false, conf.config), nil
}

func writeHeaderKey(w *bufio.Writer, key string) (err error) {
	if _, err = w.WriteString(key); err != nil {
		return
	}
	if _, err = w.WriteString(": "); err != nil {
		return
	}
	return
}

func writeHeaderVal(w *bufio.Writer, val string) (err error) {
	if _, err = w.WriteString(val); err != nil {
		return
	}

	if _, err = w.WriteString("\r\n"); err != nil {
		return
	}
	return
}

// https://datatracker.ietf.org/doc/html/rfc6455#section-4.2.2
// 第5小点
func writeResponse(r *http.Request, w *bufio.Writer, cnf config) (err error) {
	if _, err = w.WriteString(strHeaderUpgrade); err != nil {
		return
	}

	// 写入Sec-WebSocket-Accept key
	if err = writeHeaderKey(w, "Sec-WebSocket-Accept"); err != nil {
		return
	}
	// 写入Sec-WebSocket-Accept vla
	if err = writeHeaderVal(w, secWebSocketAcceptVal(r.Header.Get("Sec-WebSocket-Key"))); err != nil {
		return
	}
	// 给客户端回个信, 表示支持解压缩模式
	if cnf.decompression {
		if _, err = w.WriteString(strHeaderExtensions); err != nil {
			return
		}
	}

	if _, err = w.WriteString(strCRLF); err != nil {
		return
	}
	// TODO 5小点, 处理子协议
	if err = w.Flush(); err != nil {
		return
	}
	return
}

// https://datatracker.ietf.org/doc/html/rfc6455#section-4.2.1
// 按rfc标准, 先来一顿if else判断, 检查发的request是否满足标准
func checkRequest(r *http.Request) (ecode int, err error) {
	// 不是get方法的
	if r.Method != http.MethodGet {
		// TODO错误消息
		return http.StatusMethodNotAllowed, fmt.Errorf("%w :%s", ErrOnlyGETSupported, r.Method)
	}
	// http版本低于1.1
	if !r.ProtoAtLeast(1, 1) {
		// TODO错误消息
		return http.StatusHTTPVersionNotSupported, ErrHTTPProtocolNotSupported
	}

	// 没有host字段的
	if r.Host == "" {
		return http.StatusBadRequest, ErrHostCannotBeEmpty
	}

	// Upgrade值不等于websocket的
	if upgrade := r.Header.Get("Upgrade"); !strings.EqualFold(upgrade, "websocket") {
		return http.StatusBadRequest, ErrUpgradeFieldValue
	}

	// Connection值不是Upgrade
	if conn := r.Header.Get("Connection"); !strings.EqualFold(conn, "Upgrade") {
		return http.StatusBadRequest, ErrConnectionFieldValue
	}

	// Sec-WebSocket-Key解码之后是16字节长度
	// TODO后续优化
	if len(r.Header.Get("Sec-WebSocket-Key")) == 0 {
		return http.StatusBadRequest, ErrSecWebSocketKey
	}

	// Sec-WebSocket-Version的版本不是13的
	if r.Header.Get("Sec-WebSocket-Version") != "13" {
		return http.StatusUpgradeRequired, ErrSecWebSocketVersion
	}

	// TODO Sec-WebSocket-Protocol
	// TODO Sec-WebSocket-Extensions
	return 0, nil
}
