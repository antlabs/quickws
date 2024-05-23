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
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"unsafe"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

var mu sync.Mutex
var uuid = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")

func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// // StringToBytes 没有内存开销的转换
// func StringToBytes(s string) (b []byte) {
// 	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
// 	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
// 	bh.Data = sh.Data
// 	bh.Len = sh.Len
// 	bh.Cap = sh.Len
// 	return b
// }

// func BytesToString(b []byte) string {
// 	return *(*string)(unsafe.Pointer(&b))
// }

func secWebSocketAccept() string {
	// rfc规定是16字节
	var key [16]byte
	mu.Lock()
	rng.Read(key[:])
	mu.Unlock()
	return base64.StdEncoding.EncodeToString(key[:])
}

func secWebSocketAcceptVal(val string) string {
	s := sha1.New()
	s.Write(StringToBytes(val))
	s.Write(uuid)
	r := s.Sum(nil)
	return base64.StdEncoding.EncodeToString(r)
}

func getHttpErrMsg(statusCode int) error {
	errMsg := http.StatusText(statusCode)
	if errMsg != "" {
		return errors.New(errMsg)
	}

	return fmt.Errorf("status code:%d", statusCode)
}
