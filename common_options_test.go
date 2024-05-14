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
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var badUTF8 = []byte{128, 129, 130, 131}

type testServerOptionReadTimeout struct {
	err chan error
	run int32
}

func (defcallback *testServerOptionReadTimeout) OnOpen(_ *Conn) {
}

func (defcallback *testServerOptionReadTimeout) OnMessage(_ *Conn, _ Opcode, _ []byte) {
}

func (defcallback *testServerOptionReadTimeout) OnClose(c *Conn, err error) {
	atomic.AddInt32(&defcallback.run, int32(1))
	defcallback.err <- err
}

// 测试客户端和服务端都有的配置项
func Test_CommonOption(t *testing.T) {
	t.Run("0.server.local: Without setting WithClientCallbackFunc", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerTCPDelay())
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientCallbackFunc(func(c *Conn) {
		}, func(c *Conn, mt Opcode, payload []byte) {
		}, func(c *Conn, err error) {
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
	})

	t.Run("0.server.global: Without setting WithClientCallbackFunc", func(t *testing.T) {
		upgrade := NewUpgrade()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientCallbackFunc(func(c *Conn) {
		}, func(c *Conn, mt Opcode, payload []byte) {
		}, func(c *Conn, err error) {
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
	})

	t.Run("0.server.global: Without setting WithClientCallbackFunc", func(t *testing.T) {
		upgrade := NewUpgrade()
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url)
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
	})
	t.Run("0.client: WithClientCallbackFunc", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerTCPDelay(), WithServerOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
				c.WriteMessage(mt, payload)
				atomic.AddInt32(&run, int32(1))
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		messageDone := make(chan bool, 1)
		url := strings.ReplaceAll(ts.URL, "http", "ws")
		clientRun := int32(0)
		con, err := Dial(url, WithClientCallbackFunc(func(c *Conn) {
			atomic.AddInt32(&clientRun, 10)
		}, func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&clientRun, 100)
			messageDone <- true
		}, func(c *Conn, err error) {
			atomic.AddInt32(&clientRun, 1000)
			done <- true
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		for i := 0; i < 2; i++ {
			select {
			case <-messageDone:
				con.Close()
			case <-done:
			case <-time.After(100 * time.Millisecond):
			}
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}

		if atomic.LoadInt32(&clientRun) != 1110 {
			t.Errorf("not run client:method fail:%d, need:1110\n", atomic.LoadInt32(&clientRun))
		}
	})

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

		err = con.WriteMessage(Text, badUTF8)
		if err == nil {
			t.Error("写入非法utf8数据，没有报错")
		}

		con.WriteMessage(Binary, badUTF8)
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

		err = con.WriteMessage(Text, badUTF8)
		if err == nil {
			t.Error("写入非法utf8数据，没有报错")
		}
		_ = con.WriteMessage(Binary, badUTF8)

		select {
		case <-done:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("3.client: WithClientDisableUTF8Check", func(t *testing.T) {
		// 客户端不检查utf8， 服务端检查utf8
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
		con, err := Dial(url, WithClientCallback(&testDefaultCallback{}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		err = con.WriteMessage(Text, badUTF8)
		if err != nil {
			t.Error("关闭utf8检查, 写入非法utf8数据，不报错")
		}

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 0 {
			t.Error("not run server:method fail")
		}
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

	t.Run("10.client: WithClientDecompression", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerDecompressAndCompress(),
				WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
					c.WriteMessage(op, payload)
				}))
			if err != nil {
				t.Error(err)
				return
			}
			c.ReadLoop()
		}))

		defer ts.Close()

		fmt.Printf(">>> WithClientDecompression.%s\n", ts.URL)
		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientBufioParseMode(), WithClientCompression(), WithClientDecompression(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
			return
		}
		defer con.Close()

		err = con.WriteMessage(Binary, []byte("hello"))
		if err != nil {
			t.Error(err)
			return
		}

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

	t.Run("12.server.local: WithServerBufioMultipleTimesPayloadSize", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r,
				WithServerBufioParseMode(),
				WithServerBufioMultipleTimesPayloadSize(6), /*这里写-1只是为了代码覆盖度测试*/
				WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
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

	t.Run("12.server.global: WithServerBufioMultipleTimesPayloadSize", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerWindowsParseMode(), WithServerBufioMultipleTimesPayloadSize(-1), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
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

	t.Run("12.client: WithServerBufioMultipleTimesPayloadSize", func(t *testing.T) {
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
		con, err := Dial(url, WithClientWindowsParseMode(), WithClientBufioMultipleTimesPayloadSize(-1), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
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

	t.Run("13-15.server.local.1: WriteMessageDelay", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerMaxDelayWriteDuration(time.Millisecond*20), WithServerMaxDelayWriteNum(3), WithServerDelayWriteInitBufferSize(4096), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessageDelay(op, payload)
				c.WriteMessageDelay(op, payload)
				c.WriteMessageDelay(op, payload)
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
			if run == 3 {
			}
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		con.StartReadLoop()
		for i := 0; i < 3; i++ {
			select {
			case d := <-data:
				if d != "hello" {
					t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
				}
			case <-time.After(1000 * time.Millisecond):
			}
		}
		if atomic.LoadInt32(&run) != 3 {
			t.Errorf("not run server:method fail:%d\n", atomic.LoadInt32(&run))
		}
	})

	t.Run("13-15.server.global.2: WriteMessageDelay", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerMaxDelayWriteDuration(time.Millisecond*20),
			WithServerMaxDelayWriteNum(3),
			WithServerDelayWriteInitBufferSize(4096),
			WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				err := c.WriteMessageDelay(op, payload)
				if err != nil {
					t.Errorf("write message fail:%v\n", err)
				}
				err = c.WriteMessageDelay(op, payload)
				if err != nil {
					t.Errorf("write message fail:%v\n", err)
				}
				err = c.WriteMessageDelay(op, payload)
				if err != nil {
					t.Errorf("write message fail:%v\n", err)
				}
			}))

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
				return
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
		if atomic.LoadInt32(&run) != 3 {
			t.Errorf("not run server:method fail: run:%d,need:3\n", run)
		}
	})

	t.Run("13-15.client: WriteMessageDelay", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r,
				WithServerDecompressAndCompress(),
				WithServerBufioParseMode(),
				WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
					c.WriteMessage(op, payload)
					atomic.AddInt32(&run, int32(1))
					data <- string(payload)
				}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientDecompressAndCompress(),
			WithClientDecompression(),
			WithClientMaxDelayWriteDuration(30*time.Millisecond),
			WithClientMaxDelayWriteNum(3),
			WithClientWindowsParseMode(),
			WithClientDelayWriteInitBufferSize(4096),
			WithClientOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessageDelay(op, []byte("hello"))
				c.WriteMessageDelay(op, []byte("hello"))
				c.WriteMessageDelay(op, []byte("hello"))
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

	t.Run("13-15.client: WriteMessageDelay-Compress", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r,
				WithServerDecompressAndCompress(),
				// WithServerBufioParseMode(),
				WithServerCallbackFunc(nil, func(c *Conn, op Opcode, payload []byte) {
					if op != Binary {
						t.Error("opcode error")
					}
					c.WriteMessage(op, payload)
				}, func(c *Conn, err error) {
					// t.Errorf("%T\n", err)
				},
				))
			if err != nil {
				t.Error(err)
			}

			if !c.Compression {
				t.Error("compression fail")
			}

			if !c.pd.Decompression {
				t.Error("compression fail")
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url,
			WithClientDecompressAndCompress(),
			WithClientMaxDelayWriteDuration(30*time.Millisecond),
			WithClientMaxDelayWriteNum(3),
			WithClientWindowsParseMode(),
			WithClientDelayWriteInitBufferSize(4096),
			WithClientOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				if op != Binary {
					t.Error("opcode error")
				}
				err := c.WriteMessageDelay(op, []byte("hello"))
				if err != nil {
					t.Error(err)
				}
				err = c.WriteMessageDelay(op, []byte("hello"))
				if err != nil {
					t.Error(err)
				}
				err = c.WriteMessageDelay(op, []byte("hello"))
				if err != nil {
					t.Error(err)
				}
				data <- "hello"
				atomic.AddInt32(&run, int32(1))
			}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		if !con.Compression {
			t.Error("not compression:method fail")
		}
		err = con.WriteMessage(Binary, []byte("hello"))
		if err != nil {
			t.Error(err)
		}

		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
			t.Errorf("write message timeout\n")
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("13-15.client: WriteMessageDelay-Compress-checkUTF8", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r,
				WithServerBufioParseMode(),
				WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url,
			WithClientEnableUTF8Check(),
			WithClientOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
			}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		err = con.WriteMessageDelay(Text, []byte{128, 129, 130, 131})
		if err == nil {
			t.Error("not check utf8:method fail")
		}
		con.StartReadLoop()
	})

	t.Run("13-15.client: WriteMessageDelay-timeout-send", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r,
				WithServerDecompressAndCompress(),
				WithServerBufioParseMode(),
				WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
					c.WriteMessage(op, payload)
					atomic.AddInt32(&run, int32(1))
					data <- string(payload)
				}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientDecompressAndCompress(),
			WithClientDecompression(),
			WithClientMaxDelayWriteDuration(30*time.Millisecond),
			WithClientMaxDelayWriteNum(3),
			WithClientWindowsParseMode(),
			WithClientDelayWriteInitBufferSize(4096),
			WithClientOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
			}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		err = con.WriteMessageDelay(Text, []byte("hello"))
		if err != nil {
			t.Error(err)
		}
		con.StartReadLoop()
		select {
		case d := <-data:
			if d != "hello" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
			t.Errorf("13-15.client: WriteMessageDelay-timeout-send: timeout \n")
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("13-15.client: WriteMessageDelay-Compress-Close", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r,
				WithServerBufioParseMode(),
				WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url,
			WithClientEnableUTF8Check(),
			WithClientOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
			}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.Close()
		err = con.WriteMessageDelay(Text, []byte{128, 129, 130, 131})
		if err != ErrClosed {
			t.Error("not Close:method fail")
		}
		con.StartReadLoop()
	})

	t.Run("16.1.WithServerReadTimeout:local-Upgrade", func(t *testing.T) {
		var tsort testServerOptionReadTimeout

		tsort.err = make(chan error, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerCallback(&tsort), WithServerReadTimeout(time.Millisecond*60))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientBufioParseMode(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Text, []byte("hello"))
		select {
		case d := <-tsort.err:
			if d == nil {
				t.Errorf("got:nil, need:error\n")
			}
		case <-time.After(1000 * time.Millisecond):
			t.Errorf(" Test_ServerOption:WithServerReadTimeout timeout\n")
		}
		if atomic.LoadInt32(&tsort.run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("16.2.WithServerReadTimeout", func(t *testing.T) {
		var tsort testServerOptionReadTimeout

		upgrade := NewUpgrade(WithServerCallback(&tsort), WithServerReadTimeout(time.Millisecond*60))
		tsort.err = make(chan error, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientBufioParseMode(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Text, []byte("hello"))
		select {
		case d := <-tsort.err:
			if d == nil {
				t.Errorf("got:nil, need:error\n")
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf(" Test_ServerOption:WithServerReadTimeout timeout\n")
		}
		if atomic.LoadInt32(&tsort.run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("16.3.WithClientReadTimeout", func(t *testing.T) {
		var tsort testServerOptionReadTimeout

		upgrade := NewUpgrade(WithServerCallback(&tsort), WithServerReadTimeout(time.Millisecond*60))
		tsort.err = make(chan error, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientBufioParseMode(), WithClientReadTimeout(2*time.Second), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Text, []byte("hello"))
		select {
		case d := <-tsort.err:
			if d == nil {
				t.Errorf("got:nil, need:error\n")
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf(" Test_ServerOption:WithServerReadTimeout timeout\n")
		}
		if atomic.LoadInt32(&tsort.run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("17.client.WithClientOnCloseFunc", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerBufioParseMode(), WithServerEnableUTF8Check(), WithServerOnCloseFunc(func(c *Conn, err error) {
			c.Close()
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
		con, err := Dial(url, WithClientDisableBufioClearHack(),
			WithClientEnableUTF8Check(), WithClientOnCloseFunc(func(c *Conn, err error) {
				atomic.AddInt32(&run, 1)
				data <- err.Error()
			}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()
		// 这里必须要报错
		err = con.WriteMessage(Text, []byte("hello"))
		if err != nil {
			t.Error("not error")
		}
		con.StartReadLoop()
		select {
		case _ = <-data:
		case <-time.After(500 * time.Millisecond):
		}

		if atomic.LoadInt32(&run) != 0 {
			t.Error("not run server:method fail")
		}
	})
}
