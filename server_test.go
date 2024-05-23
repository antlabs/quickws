// Copyright 2021-2024 antlabs. All rights reserved.
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
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// 测试服务端握手失败的情况
func Test_Server_HandshakeFail(t *testing.T) {
	// u := NewUpgrade()
	t.Run("local config:case:method fail", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := Upgrade(w, r)
			if err == nil {
				t.Error("upgrade method fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		defer ts.Close()

		url := ts.URL
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			t.Error(err)
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("global config:case:method fail", func(t *testing.T) {
		run := int32(0)
		upgrade := NewUpgrade()
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := upgrade.Upgrade(w, r)
			if err == nil {
				t.Error("upgrade method fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		defer ts.Close()

		url := ts.URL
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			t.Error(err)
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("local config:case:http proto fail", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := Upgrade(w, r)
			if err == nil {
				t.Error("upgrade http proto fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		url := strings.ReplaceAll(ts.URL, "http://", "")
		defer ts.Close()
		c, err := net.Dial("tcp", url)
		if err != nil {
			t.Error(err)
		}
		_, err = c.Write([]byte("GET / HTTP/1.0\r\nHost: localhost:8080\r\n\r\n"))
		if err != nil {
			t.Error(err)
			return
		}
		c.Close()
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}

		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:http proto fail")
		}
	})

	t.Run("global config:case:http proto fail", func(t *testing.T) {
		run := int32(0)
		upgrade := NewUpgrade()
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := upgrade.Upgrade(w, r)
			if err == nil {
				t.Error("upgrade http proto fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		url := strings.ReplaceAll(ts.URL, "http://", "")
		defer ts.Close()
		c, err := net.Dial("tcp", url)
		if err != nil {
			t.Error(err)
		}
		_, err = c.Write([]byte("GET / HTTP/1.0\r\n\r\n"))
		if err != nil {
			t.Error(err)
			return
		}
		// c.Write([]byte("GET / HTTP/1.0\r\nHost: localhost:8080\r\n\r\n"))
		c.Close()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:http proto fail")
		}
	})

	t.Run("local config:case:host empty", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := Upgrade(w, r)
			if err == nil {
				t.Error("upgrade host fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http://", "")
		c, err := net.Dial("tcp", url)
		if err != nil {
			t.Error(err)
		}
		_, err = c.Write([]byte("GET / HTTP/1.1\r\nHost: \r\n\r\n"))
		if err != nil {
			t.Error(err)
			return
		}
		defer c.Close()
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:host empty")
		}
	})

	t.Run("global config:case:upgrade fail", func(t *testing.T) {
		run := int32(0)
		upgrade := NewUpgrade()
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := upgrade.Upgrade(w, r)
			if err == nil {
				t.Error("upgrade : upgrade field fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		url := strings.ReplaceAll(ts.URL, "http://", "")
		defer ts.Close()
		c, err := net.Dial("tcp", url)
		if err != nil {
			t.Error(err)
		}
		wbuf := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: xx\r\n\r\n", url))
		_, err = c.Write(wbuf)
		if err != nil {
			t.Error(err)
			return
		}
		c.Close()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:upgrade field fail")
		}
	})

	t.Run("local config:case:upgrade fail", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := Upgrade(w, r)
			if err == nil {
				t.Error("upgrade : upgrade field fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		url := strings.ReplaceAll(ts.URL, "http://", "")
		defer ts.Close()
		c, err := net.Dial("tcp", url)
		if err != nil {
			t.Error(err)
		}
		wbuf := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: xx\r\n\r\n", url))
		_, err = c.Write(wbuf)
		if err != nil {
			t.Error(err)
			return
		}
		c.Close()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:upgrade field fail")
		}
	})

	t.Run("global config:case:Connection fail", func(t *testing.T) {
		run := int32(0)
		upgrade := NewUpgrade()
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := upgrade.Upgrade(w, r)
			if err == nil {
				t.Error("upgrade : Connection field fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		url := strings.ReplaceAll(ts.URL, "http://", "")
		defer ts.Close()
		c, err := net.Dial("tcp", url)
		if err != nil {
			t.Error(err)
		}
		wbuf := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: xx\r\n\r\n", url))
		_, err = c.Write(wbuf)
		if err != nil {
			t.Error(err)
			return
		}
		c.Close()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:Connection field fail")
		}
	})

	t.Run("local config:case:Connection fail", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := Upgrade(w, r)
			if err == nil {
				t.Error("upgrade : Connection field fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		url := strings.ReplaceAll(ts.URL, "http://", "")
		defer ts.Close()
		c, err := net.Dial("tcp", url)
		if err != nil {
			t.Error(err)
		}
		wbuf := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: xx\r\n\r\n", url))
		_, err = c.Write(wbuf)
		if err != nil {
			t.Error(err)
			return
		}
		c.Close()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:Connection field fail")
		}
	})

	t.Run("global config:case: Sec-WebSocket-Key fail", func(t *testing.T) {
		run := int32(0)
		upgrade := NewUpgrade()
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := upgrade.Upgrade(w, r)
			if err == nil {
				t.Error("upgrade : Connection field fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		url := strings.ReplaceAll(ts.URL, "http://", "")
		defer ts.Close()
		c, err := net.Dial("tcp", url)
		if err != nil {
			t.Error(err)
		}
		wbuf := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n", url))
		_, err = c.Write(wbuf)
		if err != nil {
			t.Error(err)
			return
		}
		c.Close()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:Connection field fail")
		}
	})

	t.Run("local config:case: Sec-WebSocket-Key fail", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := Upgrade(w, r)
			if err == nil {
				t.Error("upgrade : Connection field fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		url := strings.ReplaceAll(ts.URL, "http://", "")
		defer ts.Close()
		c, err := net.Dial("tcp", url)
		if err != nil {
			t.Error(err)
		}
		wbuf := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n", url))
		_, err = c.Write(wbuf)
		if err != nil {
			t.Error(err)
			return
		}
		c.Close()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:Connection field fail")
		}
	})

	t.Run("global config:case: Sec-WebSocket-Version fail", func(t *testing.T) {
		run := int32(0)
		upgrade := NewUpgrade()
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := upgrade.Upgrade(w, r)
			if err == nil {
				t.Error("upgrade : Connection field fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		url := strings.ReplaceAll(ts.URL, "http://", "")
		defer ts.Close()
		c, err := net.Dial("tcp", url)
		if err != nil {
			t.Error(err)
		}
		wbuf := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: key\r\n\r\n", url))
		_, err = c.Write(wbuf)
		if err != nil {
			t.Error(err)
			return
		}
		c.Close()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:Connection field fail")
		}
	})

	t.Run("local config:case: Sec-WebSocket-Version fail", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := Upgrade(w, r)
			if err == nil {
				t.Error("upgrade : Connection field fail")
			}
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		url := strings.ReplaceAll(ts.URL, "http://", "")
		defer ts.Close()
		c, err := net.Dial("tcp", url)
		if err != nil {
			t.Error(err)
		}
		wbuf := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: key\r\n\r\n", url))
		_, err = c.Write(wbuf)
		if err != nil {
			t.Error(err)
			return
		}
		c.Close()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:Connection field fail")
		}
	})
}
