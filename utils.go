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
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"reflect"
	"unsafe"
)

var uuid = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

// StringToBytes 没有内存开销的转换
func StringToBytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

func newMask(mask []byte) {
	rand.Read(mask)
}

func secWebSocketAccept() string {
	// rfc规定是16字节
	key := make([]byte, 16)
	rand.Read(key)
	return base64.StdEncoding.EncodeToString(key)
}

func secWebSocketAcceptVal(val string) string {
	s := sha1.New()
	s.Write(StringToBytes(val))
	s.Write(uuid)
	r := s.Sum(nil)
	return base64.StdEncoding.EncodeToString(r)
}

// 获取没有绑定服务的端口
func GetNoPortExists() string {
	startPort := 1000 //1000以下的端口很多时间需要root权限才能使用
	for port := startPort; port < 65535; port++ {
		l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			continue
		}
		l.Close()
		return fmt.Sprintf("%d", port)
	}

	return "0"
}

// 是否打开压缩
func needCompression(header http.Header) bool {
	for _, ext := range parseExtensions(header) {
		if ext[""] != "permessage-deflate" {
			continue
		}
		return true
	}

	return false
}
