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
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// 测试客户端和服务端都有的配置项
func Test_CommonOption(t *testing.T) {
	t.Run("2.server.local: WithServerTCPDelay", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerTCPDelay(), WithServerOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
				atomic.AddInt32(&run, int32(1))
				done <- true
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientCallback(&testDefaultCallback{}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		select {
		case <-done:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("2.server.global: WithServerTCPDelay", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		upgrade := NewUpgrade(WithServerTCPDelay(), WithServerOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientCallback(&testDefaultCallback{}))
		if err != nil {
			t.Error(err)
		}

		con.WriteMessage(Binary, []byte("hello"))
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("2.client: WithClientTCPDelay ", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
				c.WriteMessage(mt, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientTCPDelay(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))
		if err != nil {
			t.Error(err)
		}

		con.StartReadLoop()
		con.WriteMessage(Binary, []byte("hello"))
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run client callback:method fail")
		}
	})

	t.Run("3.server.local: WithServerDisableUTF8Check", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerEnableUTF8Check(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessage(op, payload)
				atomic.AddInt32(&run, int32(1))
				done <- true
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientEnableUTF8Check(), WithClientCallback(&testDefaultCallback{}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		err = con.WriteMessage(Text, []byte{1, 2, 3, 4})
		if err != nil {
			t.Error(err)
		}
		select {
		case <-done:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("3.server.global: WithServerDisableUTF8Check", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		upgrade := NewUpgrade(WithServerEnableUTF8Check(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
			c.WriteMessage(op, payload)
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientEnableUTF8Check(), WithClientCallback(&testDefaultCallback{}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		err = con.WriteMessage(Text, []byte{1, 2, 3, 4})
		if err != nil {
			t.Error(err)
		}
		select {
		case <-done:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("3.client: WithClientDisableUTF8Check", func(t *testing.T) {
		// 3.server.local: WithServerDisableUTF8Check 已经测试过客户端，所以这里留空
	})

	t.Run("4.server.local: WithServerOnMessageFunc", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
				atomic.AddInt32(&run, int32(1))
				done <- true
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientCallback(&testDefaultCallback{}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		select {
		case <-done:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("4.server.global: WithServerOnMessageFunc", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		upgrade := NewUpgrade(WithServerOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientCallback(&testDefaultCallback{}))
		if err != nil {
			t.Error(err)
		}

		con.WriteMessage(Binary, []byte("hello"))
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("4.client: WithClientOnMessageFunc", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
				c.WriteMessage(mt, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))
		if err != nil {
			t.Error(err)
		}

		con.StartReadLoop()
		con.WriteMessage(Binary, []byte("hello"))
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run client callback:method fail")
		}
	})

	t.Run("5.server.local: WithServerReplyPing", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerReplyPing(), WithServerOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientOnMessageFunc(func(c *Conn, o Opcode, b []byte) {
			atomic.AddInt32(&run, int32(1))
			if o != Pong || bytes.Equal(b, []byte("hello")) {
				t.Error("need Pong")
			}
			done <- true
		}))
		if err != nil {
			t.Error(err)
		}

		con.StartReadLoop()
		con.WritePing([]byte("hello"))
		select {
		case <-done:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("5.server.global: WithServerReplyPing", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		upgrade := NewUpgrade(WithServerReplyPing(), WithServerOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
		}))
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientOnMessageFunc(func(c *Conn, o Opcode, b []byte) {
			atomic.AddInt32(&run, int32(1))
			if o != Pong || bytes.Equal(b, []byte("hello")) {
				t.Error("need Pong")
			}
			done <- true
		}))
		if err != nil {
			t.Error(err)
		}

		con.StartReadLoop()
		err = con.WritePing([]byte("hello"))
		if err != nil {
			t.Errorf("WritePing:%v", err)
		}
		select {
		case <-done:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("5.client: WithClientReplyPing", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerReplyPing(), WithServerOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
				if mt == Text {
					c.WritePing([]byte("hello"))
				} else if mt == Pong {
				} else {
					t.Error("unknown opcode")
				}

				atomic.AddInt32(&run, int32(1))
				if atomic.LoadInt32(&run) >= 2 {
					done <- true
				}
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientReplyPing(), WithClientOnMessageFunc(func(c *Conn, o Opcode, b []byte) {
		}))
		if err != nil {
			t.Error(err)
		}

		con.StartReadLoop()
		con.WriteMessage(Text, []byte("hello"))
		select {
		case <-done:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 2 {
			t.Errorf("run:%d\n", run)
		}
	})

	t.Run("6.server.local: WithServerIgnorePong", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerIgnorePong(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessage(op, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			fmt.Printf("opcode:%v\n", mt)
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WritePong([]byte("hello"))
		con.WriteMessage(Text, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("6.server.global: WithServerIgnorePong", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerIgnorePong(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
			c.WriteMessage(op, payload)
		}))
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			fmt.Printf("opcode:%v\n", mt)
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WritePong([]byte("hello"))
		con.WriteMessage(Text, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("6.client: WithClientIgnorePong", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerIgnorePong(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WritePong([]byte("hello"))
				c.WriteMessage(op, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientIgnorePong(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Text, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("7.server.local: WithServerWindowsMultipleTimesPayloadSize", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerWindowsParseMode(), WithServerWindowsMultipleTimesPayloadSize(-1), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessage(op, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("7.server.global: WithServerWindowsMultipleTimesPayloadSize", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerWindowsParseMode(), WithServerWindowsMultipleTimesPayloadSize(-1), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
			c.WriteMessage(op, payload)
		}))
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("7.client: WithServerWindowsMultipleTimesPayloadSize", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerBufioParseMode(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessage(op, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientWindowsParseMode(), WithClientWindowsMultipleTimesPayloadSize(-1), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})
	t.Run("8.server.local: WithServerWindowsParseMode", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerWindowsParseMode(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessage(op, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("8.server.global: WithServerWindowsParseMode", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerWindowsParseMode(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
			c.WriteMessage(op, payload)
		}))
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("8.client: WithClientWindowsParseMode", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerBufioParseMode(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessage(op, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientWindowsParseMode(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("9.server.local: WithServerBufioParseMode", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerBufioParseMode(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessage(op, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("9.server.global: WithServerBufioParseMode", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerBufioParseMode(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
			c.WriteMessage(op, payload)
		}))
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("9.client: WithClientBufioParseMode", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerBufioParseMode(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessage(op, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientBufioParseMode(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("11.server.local: WithServerDisableBufioClearHack", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerDisableBufioClearHack(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessage(op, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientDisableBufioClearHack(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("11.server.global: WithServerDisableBufioClearHack", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerBufioParseMode(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
			c.WriteMessage(op, payload)
		}))
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientDisableBufioClearHack(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("11.client: WithClientBufioParseMode", func(t *testing.T) {
		// 11.server.local 已经测试过客户端，所以这里留空
	})
}
