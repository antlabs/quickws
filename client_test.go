// Copyright 2021-2023 antlabs. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package quickws

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// 测试客户端Dial, 返回的http.Header
func Test_Client_Dial_Check_Header(t *testing.T) {
	t.Run("Dial: valid resp: status code fail", func(t *testing.T) {
		done := make(chan bool, 1)
		run := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		defer ts.Close()

		rawURL := strings.ReplaceAll(ts.URL, "http", "ws")
		_, err := Dial(rawURL)
		if err == nil {
			t.Fatal("should be error")
		}
	})

	t.Run("DialConf: valid resp : status code fail", func(t *testing.T) {
		done := make(chan bool, 1)
		run := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		defer ts.Close()

		cnf := ClientOptionToConf()
		rawURL := strings.ReplaceAll(ts.URL, "http", "ws")
		_, err := DialConf(rawURL, cnf)
		if err == nil {
			t.Fatal("should be error")
		}
	})

	t.Run("Dial: valid resp: Upgrade field fail", func(t *testing.T) {
		done := make(chan bool, 1)
		run := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&run, int32(1))
			w.WriteHeader(101)
			w.Header().Set("Upgrade", "xx")
			done <- true
		}))

		defer ts.Close()

		rawURL := strings.ReplaceAll(ts.URL, "http", "ws")
		_, err := Dial(rawURL)
		if err == nil {
			t.Fatal("should be error")
		}
	})

	t.Run("DialConf: valid resp: Upgrade field fail", func(t *testing.T) {
		done := make(chan bool, 1)
		run := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&run, int32(1))
			w.WriteHeader(101)
			w.Header().Set("Upgrade", "xx")
			done <- true
		}))

		defer ts.Close()

		cnf := ClientOptionToConf()
		rawURL := strings.ReplaceAll(ts.URL, "http", "ws")
		_, err := DialConf(rawURL, cnf)
		if err == nil {
			t.Fatal("should be error")
		}
	})

	t.Run("Dial: valid resp: Connection fail", func(t *testing.T) {
		done := make(chan bool, 1)
		run := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&run, int32(1))
			w.Header().Set("Upgrade", "websocket")
			w.Header().Set("Connection", "xx")
			w.WriteHeader(101)
			done <- true
		}))

		defer ts.Close()

		rawURL := strings.ReplaceAll(ts.URL, "http", "ws")
		_, err := Dial(rawURL)
		if err == nil {
			t.Fatal("should be error")
		}
	})

	t.Run("DialConf: valid resp: Connection fail", func(t *testing.T) {
		done := make(chan bool, 1)
		run := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&run, int32(1))
			w.Header().Set("Upgrade", "websocket")
			w.Header().Set("Connection", "xx")
			w.WriteHeader(101)
			done <- true
		}))

		defer ts.Close()

		cnf := ClientOptionToConf()
		rawURL := strings.ReplaceAll(ts.URL, "http", "ws")
		_, err := DialConf(rawURL, cnf)
		if err == nil {
			t.Fatal("should be error")
		} else {
			fmt.Printf("err: %v\n", err)
		}
	})

	t.Run("Dial: valid resp: Sec-WebSocket-Accept fail", func(t *testing.T) {
		done := make(chan bool, 1)
		run := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&run, int32(1))
			w.Header().Set("Upgrade", "websocket")
			w.Header().Set("Connection", "Upgrade")
			w.WriteHeader(101)
			done <- true
		}))

		defer ts.Close()

		rawURL := strings.ReplaceAll(ts.URL, "http", "ws")
		_, err := Dial(rawURL)
		if err == nil {
			t.Fatal("should be error")
		}
	})

	t.Run("DialConf: valid resp: Sec-WebSocket-Accept fail", func(t *testing.T) {
		done := make(chan bool, 1)
		run := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&run, int32(1))
			w.Header().Set("Upgrade", "websocket")
			w.Header().Set("Connection", "Upgrade")
			w.WriteHeader(101)
			done <- true
		}))

		defer ts.Close()

		cnf := ClientOptionToConf()
		rawURL := strings.ReplaceAll(ts.URL, "http", "ws")
		_, err := DialConf(rawURL, cnf)
		if err == nil {
			t.Fatal("should be error")
		} else {
			fmt.Printf("err: %v\n", err)
		}
	})
}
