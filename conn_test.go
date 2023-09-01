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
	"crypto/md5"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/antlabs/wsutil/fixedwriter"
	"github.com/antlabs/wsutil/frame"
	"github.com/antlabs/wsutil/opcode"
)

var (
	testBinaryMessage64kb = bytes.Repeat([]byte("1"), 65535)
	testTextMessage64kb   = bytes.Repeat([]byte("中"), 65535/len("中"))
	testBinaryMessage10   = bytes.Repeat([]byte("1"), 10)
)

type testMessageHandler struct {
	DefCallback
	t           *testing.T
	need        []byte
	callbed     int32
	callbedChan chan bool
	server      bool
	count       int
	got         chan []byte
	done        chan struct{}
	output      bool
}

func (t *testMessageHandler) OnMessage(c *Conn, op opcode.Opcode, msg []byte) {
	need := append([]byte(nil), t.need...)
	atomic.StoreInt32(&t.callbed, 1)
	t.callbedChan <- true
	if t.count == 0 {
		return
	}
	t.count--

	message := "#client"
	if t.server {
		message = "#server"
	}
	if t.output {
		// fmt.Printf(">>>>>%p %s, %#v\n", &msg, message, msg)
	}
	if len(msg) < 30 {
		if !bytes.Equal(msg, need) {
			t.t.Errorf(">>>>>%p %s, %#v\n", &msg, message, msg)
		}
	} else {

		md51 := md5.Sum(need)
		md52 := md5.Sum(msg)
		if !bytes.Equal(md51[:], md52[:]) {
			t.t.Errorf("md51 %x, md52 %x\n", md51, md52)
		}
	}
	err := c.WriteMessage(op, msg)
	if err != nil {
		t.t.Error(err)
	}
	if !t.server {
		// c.Close()
	}
}

func (t *testMessageHandler) OnClose(c *Conn, err error) {
	message := "#client.OnClose"
	if t.server {
		message = "#server.OnClose"
	}

	fmt.Printf("OnClose: %s:%s\n", message, err)
	if t.done != nil {
		close(t.done)
	}
}

func newServrEcho(t *testing.T, data []byte, output bool) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := Upgrade(w, r,
			WithServerCallback(&testMessageHandler{t: t, need: data, server: true, count: -1, output: true, callbedChan: make(chan bool, 1)}),
		)
		if err != nil {
			t.Error(err)
			return
		}

		c.ReadLoop()
	}))

	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
	return ts
}

// 测试read message
func Test_ReadMessage(t *testing.T) {
	t.Run("ReadMessage10", func(t *testing.T) {
		ts := newServrEcho(t, testBinaryMessage10, true)
		client := &testMessageHandler{t: t, need: append([]byte(nil), testBinaryMessage10...), count: 1, done: make(chan struct{}), output: true}
		client.callbedChan = make(chan bool, 1)
		c, err := Dial(ts.URL, WithClientCallback(client))
		if err != nil {
			t.Error(err)
			return
		}
		go c.ReadLoop()

		tmp := append([]byte(nil), testBinaryMessage10...)

		err = c.WriteMessage(Binary, tmp)
		if err != nil {
			t.Error(err)
			return
		}
		// <-client.done
		select {
		case <-client.callbedChan:
		case <-time.After(time.Second / 3):
		}
		if atomic.LoadInt32(&client.callbed) != 1 {
			t.Error("not callbed")
		}
	})

	t.Run("ReadMessage64K", func(t *testing.T) {
		ts := newServrEcho(t, testBinaryMessage64kb, false)
		client := &testMessageHandler{t: t, need: append([]byte(nil), testBinaryMessage64kb...), count: 1}
		client.callbedChan = make(chan bool, 1)
		c, err := Dial(ts.URL, WithClientCallback(client))
		if err != nil {
			t.Error(err)
			return
		}
		go c.ReadLoop()

		tmp := append([]byte(nil), testBinaryMessage64kb...)
		err = c.WriteMessage(Binary, tmp)
		select {
		case <-client.callbedChan:
		case <-time.After(time.Second / 3):
		}
		if err != nil {
			t.Error(err)
			return
		}

		if atomic.LoadInt32(&client.callbed) != 1 {
			t.Error("not callbed")
		}
	})

	t.Run("ReadMessage64K_Text", func(t *testing.T) {
		ts := newServrEcho(t, testTextMessage64kb, false)
		client := &testMessageHandler{t: t, need: append([]byte(nil), testTextMessage64kb...), count: 1}
		client.callbedChan = make(chan bool, 1)
		c, err := Dial(ts.URL, WithClientCallback(client))
		if err != nil {
			t.Error(err)
			return
		}

		go c.ReadLoop()

		tmp := append([]byte(nil), testTextMessage64kb...)
		err = c.WriteMessage(Text, tmp)
		if err != nil {
			t.Error(err)
			return
		}
		select {
		case <-client.callbedChan:
		case <-time.After(time.Second / 3):
		}
		if atomic.LoadInt32(&client.callbed) != 1 {
			t.Errorf("not callbed:%d\n", client.callbed)
		}
	})

	t.Run("ReadMessage_Fail_Rsv.1", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
				c.WriteMessage(op, payload)
			}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientBufioParseMode(), WithClientCompression(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		// err = con.WriteMessage(Binary, []byte("hello"))
		maskValue := rand.Uint32()
		var fw fixedwriter.FixedWriter
		err = frame.WriteFrame(&fw, con.c, []byte("hello"), true, true, con.client, Binary, maskValue)
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
		}
		if atomic.LoadInt32(&run) > 0 {
			// 需要不运行server
			t.Errorf("need not run server")
		}
	})

	t.Run("ReadMessage_Fail_Rsv.2", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r,
				// WithServerDecompression(),
				WithServerDecompressAndCompress(),
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
		con, err := Dial(url,
			WithClientDecompressAndCompress(),
			WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
				atomic.AddInt32(&run, int32(1))
				// data <- string(payload)
			}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		// err = con.WriteMessage(Binary, []byte("hello"))
		maskValue := rand.Uint32()
		var fw fixedwriter.FixedWriter
		err = frame.WriteFrame(&fw, con.c, []byte("hello"), true, true, con.client, Ping, maskValue)
		if err != nil {
			t.Error(err)
		}

		select {
		case _ = <-data:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) > 0 {
			// 需要不运行server
			t.Errorf("need not run server")
		}
	})
}

// 测试分段frame
func TestFragmentFrame(t *testing.T) {
	t.Run("FragmentFrame10", func(t *testing.T) {
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

		con.writeFragment(Binary, []byte("hello"), 1)
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

	t.Run("Ping-FragmentFrame-Fail", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerBufioParseMode(), WithServerOnCloseFunc(func(c *Conn, err error) {
			atomic.AddInt32(&run, int32(1))
			data <- err.Error()
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
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.writeFragment(Ping, []byte("hello"), 1)
		con.StartReadLoop()
		select {
		case <-data:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("Text-FragmentFrame-Fail", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerBufioParseMode(), WithServerOnCloseFunc(func(c *Conn, err error) {
			atomic.AddInt32(&run, int32(1))
			data <- err.Error()
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
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()
		// con.writeFragment(Ping, []byte("hello"), 1)

		maskValue := rand.Uint32()
		var fw fixedwriter.FixedWriter
		err = frame.WriteFrame(&fw, con.c, []byte("h"), false, false, con.client, Text, maskValue)
		if err != nil {
			t.Error(err)
		}
		maskValue = rand.Uint32()
		err = frame.WriteFrame(&fw, con.c, []byte{}, true, false, con.client, Text, maskValue)
		if err != nil {
			t.Error(err)
		}
		con.StartReadLoop()
		select {
		case <-data:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	// 分段传递，并且压缩
	t.Run("FragmentFrame-Compression", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerBufioParseMode(), WithServerDecompression(), WithServerOnMessageFunc(func(c *Conn, op Opcode, payload []byte) {
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
		con, err := Dial(url, WithClientDisableBufioClearHack(), WithClientDecompressAndCompress(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			atomic.AddInt32(&run, int32(1))
			data <- string(payload)
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.writeFragment(Binary, []byte("hello"), 1)
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

	t.Run("FragmentFrame-Small-Buffer", func(t *testing.T) {
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
		con, err := Dial(url, WithClientDisableBufioClearHack(), WithClientEnableUTF8Check(),
			WithClientDecompressAndCompress(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
				atomic.AddInt32(&run, int32(1))
				data <- string(payload)
			}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		sendData := []byte("hell")
		// 这里必须要报错
		err = con.writeFragment(Text, sendData, 5)
		if err != nil {
			t.Errorf("error:%v", err)
		}

		con.StartReadLoop()

		select {
		case d := <-data:
			if d != string(sendData) {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("FragmentFrame-Client-Not-UTF8", func(t *testing.T) {
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
		con, err := Dial(url, WithClientDisableBufioClearHack(), WithClientEnableUTF8Check(),
			WithClientDecompressAndCompress(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		// 这里必须要报错
		err = con.writeFragment(Text, []byte{128, 129, 130, 131}, 1)
		if err == nil {
			t.Error("not error")
		}
	})

	t.Run("FragmentFrame-Server-Not-UTF8", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerBufioParseMode(), WithServerEnableUTF8Check(), WithServerOnCloseFunc(func(c *Conn, err error) {
			data <- err.Error()
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
			WithClientDecompressAndCompress(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
			}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()
		// 这里必须要报错
		err = con.writeFragment(Text, []byte{128, 129, 130, 131}, 1)
		if err != nil {
			t.Error("error")
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

type testPingPongCloseHandler struct {
	DefCallback
	run  int32
	data chan string
}

func (t *testPingPongCloseHandler) OnClose(c *Conn, err error) {
	fmt.Printf("%s\n", err.Error())
	atomic.AddInt32(&t.run, 1)
	t.data <- "eof"
}

func Test_WriteControl(t *testing.T) {
	t.Run("WriteControl > maxControlFrameSize.message.fail", func(t *testing.T) {
		var shandler testPingPongCloseHandler
		shandler.data = make(chan string, 1)
		upgrade := NewUpgrade(WithServerBufioParseMode(), WithServerCallback(&shandler))
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
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		err = con.WriteControl(Close, bytes.Repeat([]byte{1}, 126))
		// 这里必须要报错
		if err == nil {
			t.Error("not error")
		}
	})
}

// 测试ping pong close control信息
func TestPingPongClose(t *testing.T) {
	// 写一个超过maxControlFrameSize的消息
	t.Run("1.>maxControlFrameSize.fail", func(t *testing.T) {
		run := int32(0)
		var shandler testPingPongCloseHandler
		shandler.data = make(chan string, 1)
		upgrade := NewUpgrade(WithServerBufioParseMode(), WithServerCallback(&shandler))
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
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Close, bytes.Repeat([]byte("a"), maxControlFrameSize+3))
		con.StartReadLoop()
		select {
		case d := <-shandler.data:
			if d != "eof" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&shandler.run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("3.WriteCloseEmpty.fail", func(t *testing.T) {
		run := int32(0)
		var shandler testPingPongCloseHandler
		shandler.data = make(chan string, 1)
		upgrade := NewUpgrade(WithServerBufioParseMode(), WithServerCallback(&shandler))
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
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Close, nil)
		con.StartReadLoop()
		select {
		case d := <-shandler.data:
			if d != "eof" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&shandler.run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("4.WriteClose Payload < 2.fail", func(t *testing.T) {
		run := int32(0)
		var shandler testPingPongCloseHandler
		shandler.data = make(chan string, 1)
		upgrade := NewUpgrade(WithServerBufioParseMode(), WithServerCallback(&shandler))
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
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Close, []byte{1})
		con.StartReadLoop()
		select {
		case d := <-shandler.data:
			if d != "eof" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&shandler.run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("5.WriteClose Payload > 2, utf8.check.fail", func(t *testing.T) {
		run := int32(0)
		var shandler testPingPongCloseHandler
		shandler.data = make(chan string, 1)
		upgrade := NewUpgrade(WithServerEnableUTF8Check(), WithServerBufioParseMode(), WithServerCallback(&shandler))
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
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Close, []byte{128, 129, 130, 131})
		con.StartReadLoop()
		select {
		case d := <-shandler.data:
			if d != "eof" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&shandler.run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("6.WriteClose status code.fail", func(t *testing.T) {
		run := int32(0)
		var shandler testPingPongCloseHandler
		shandler.data = make(chan string, 1)
		upgrade := NewUpgrade(WithServerEnableUTF8Check(), WithServerBufioParseMode(), WithServerCallback(&shandler))
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
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		var badSc StatusCode = 3
		con.WriteCloseTimeout(badSc, 10*time.Second)
		con.StartReadLoop()
		select {
		case d := <-shandler.data:
			if d != "eof" {
				t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&shandler.run) != 1 {
			t.Error("not run server:method fail")
		}
	})
	t.Run("7.WriteClose", func(t *testing.T) {
		run := int32(0)
		statusCodes := []StatusCode{
			NormalClosure,
			EndpointGoingAway,
			ProtocolError,
			DataCannotAccept,
			NotConsistentMessageType,
			TerminatingConnection,
			TooBigMessage,
			NoExtensions,
			ServerTerminating,
			1004,
			3000,
		}

		for _, st := range statusCodes {
			var shandler testPingPongCloseHandler
			shandler.data = make(chan string, 1)
			upgrade := NewUpgrade(WithServerEnableUTF8Check(), WithServerBufioParseMode(), WithServerCallback(&shandler))
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
			}))
			if err != nil {
				t.Error(err)
			}
			defer con.Close()

			con.WriteCloseTimeout(st, 10*time.Second)
			con.StartReadLoop()
			select {
			case d := <-shandler.data:
				if d != "eof" {
					t.Errorf("write message or read message fail:got:%s, need:hello\n", d)
				}
			case <-time.After(1000 * time.Millisecond):
			}
			if atomic.LoadInt32(&shandler.run) != 1 {
				t.Error("not run server:method fail")
			}
		}
	})

	t.Run("8.fail-control", func(t *testing.T) {
		run := int32(0)
		data := make(chan string, 1)
		upgrade := NewUpgrade(WithServerOnCloseFunc(func(c *Conn, err error) {
			atomic.AddInt32(&run, 1)
			data <- err.Error()
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
			}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()
		// 这里必须要报错
		err = con.WriteMessage(4 /*这是rfc保留的frame*/, []byte("hello"))
		if err != nil {
			t.Error("not error")
		}
		con.StartReadLoop()
		select {
		case _ = <-data:
		case <-time.After(500 * time.Millisecond):
		}

		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})
}
