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
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/net/proxy"
)

func Test_ClientOption(t *testing.T) {
	t.Run("ClientOption.WithClientHTTPHeader", func(t *testing.T) {
		done := make(chan string, 1)
		run := int32(0)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := r.Header.Get("A")
			done <- v
			con, err := Upgrade(w, r)
			if err != nil {
				t.Error(err)
				return
			}

			defer con.Close()
			atomic.AddInt32(&run, 1)
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientHTTPHeader(http.Header{
			"A": []string{"A"},
		}), WithClientCallback(&testDefaultCallback{}))
		if err != nil {
			t.Error(err)
			return
		}
		defer con.Close()

		select {
		case v := <-done:
			if v != "A" {
				t.Error("header fail")
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("ClientOption.WithClientTLSConfig", func(t *testing.T) {
		done := make(chan string, 1)
		run := int32(0)
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := r.Header.Get("A")
			atomic.AddInt32(&run, 1)
			done <- v
			con, err := Upgrade(w, r)
			if err != nil {
				t.Error(err)
				return
			}

			defer con.Close()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url,
			WithClientTLSConfig(&tls.Config{InsecureSkipVerify: true}),
			WithClientHTTPHeader(http.Header{
				"A": []string{"A"},
			}), WithClientCallback(&testDefaultCallback{}))
		if err != nil {
			t.Error(err)
			return
		}
		defer con.Close()

		select {
		case v := <-done:
			if v != "A" {
				t.Error("header fail")
			}
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("6.1 Dial: WithClientBindHTTPHeader and echo Sec-Websocket-Protocol", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		h := make(http.Header)
		con, err := Dial(url, WithClientBindHTTPHeader(&h), WithClientHTTPHeader(http.Header{
			"Sec-WebSocket-Protocol": []string{"token"},
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		if h["Sec-Websocket-Protocol"][0] != "token" {
			t.Error("header fail")
		}
	})

	t.Run("6.2 DialConf: WithClientBindHTTPHeader and echo Sec-Websocket-Protocol", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		h := make(http.Header)
		con, err := DialConf(url, ClientOptionToConf(WithClientBindHTTPHeader(&h), WithClientHTTPHeader(http.Header{
			"Sec-WebSocket-Protocol": []string{"token"},
		})))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		if h["Sec-Websocket-Protocol"][0] != "token" {
			t.Error("header fail")
		}
	})

	t.Run("18 Dial: WithClientDialFunc.1", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := Upgrade(w, r, WithServerOnMessageFunc(func(c *Conn, o Opcode, b []byte) {
				c.WriteMessage(o, b)
				c.Close()
			}))
			if err != nil {
				t.Error(err)
			}

			conn.StartReadLoop()
		}))

		proxyAddr, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Error(err)
		}
		defer ts.Close()

		go func() {
			newConn, err := proxyAddr.Accept()
			if err != nil {
				t.Error(err)
			}

			newConn.SetDeadline(time.Now().Add(30 * time.Second))

			buf := make([]byte, 128)
			if _, err := io.ReadFull(newConn, buf[:3]); err != nil {
				t.Errorf("read failed: %v", err)
				return
			}

			// socks version 5, 1 authentication method, no auth
			if want := []byte{5, 1, 0}; !bytes.Equal(want, buf[:len(want)]) {
				t.Errorf("read %x, want %x", buf[:len(want)], want)
			}

			// socks version 5, connect command, reserved, ipv4 address, port 80
			if _, err := newConn.Write([]byte{5, 0}); err != nil {
				t.Errorf("write failed: %v", err)
				return
			}

			// ver cmd rsv atyp dst.addr dst.port
			if _, err := io.ReadFull(newConn, buf[:10]); err != nil {
				t.Errorf("read failed: %v", err)
				return
			}
			if want := []byte{5, 1, 0, 1}; !bytes.Equal(want, buf[:len(want)]) {
				t.Errorf("read %x, want %x", buf[:len(want)], want)
				return
			}
			buf[1] = 0
			if _, err := newConn.Write(buf[:10]); err != nil {
				t.Errorf("write failed: %v", err)
				return
			}

			// 提取ip
			ip := net.IP(buf[4:8])
			port := binary.BigEndian.Uint16(buf[8:10])

			c2, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: ip, Port: int(port)})
			if err != nil {
				t.Errorf("dial failed; %v", err)
				return
			}
			defer c2.Close()
			done := make(chan struct{})
			go func() {
				io.Copy(newConn, c2)
				close(done)
			}()
			io.Copy(c2, newConn)
			<-done
		}()

		got := make([]byte, 0, 128)
		url := strings.ReplaceAll(ts.URL, "http", "ws")
		c, err := Dial(url, WithClientDialFunc(func() (Dialer, error) {
			return proxy.SOCKS5("tcp", proxyAddr.Addr().String(), nil, nil)
		}), WithClientOnMessageFunc(func(c *Conn, o Opcode, b []byte) {
			got = append(got, b...)
			c.Close()
		}))
		if err != nil {
			t.Error(err)
		}

		data := []byte("hello world")
		c.WriteMessage(Binary, data)
		c.ReadLoop()

		t.Log("got", string(got), "want", string(data))
		if !bytes.Equal(got, data) {
			t.Errorf("got %s, want %s", got, data)
		}
	})

	t.Run("18 Dial: WithClientDialFunc.2", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := Upgrade(w, r, WithServerOnMessageFunc(func(c *Conn, o Opcode, b []byte) {
				c.WriteMessage(o, b)
				c.Close()
			}))
			if err != nil {
				t.Error(err)
			}

			conn.StartReadLoop()
		}))

		proxyAddr, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Error(err)
		}
		defer ts.Close()

		go func() {
			newConn, err := proxyAddr.Accept()
			if err != nil {
				t.Error(err)
			}

			newConn.SetDeadline(time.Now().Add(30 * time.Second))

			buf := make([]byte, 128)
			if _, err := io.ReadFull(newConn, buf[:3]); err != nil {
				t.Errorf("read failed: %v", err)
				return
			}

			// socks version 5, 1 authentication method, no auth
			if want := []byte{5, 1, 0}; !bytes.Equal(want, buf[:len(want)]) {
				t.Errorf("read %x, want %x", buf[:len(want)], want)
			}

			// socks version 5, connect command, reserved, ipv4 address, port 80
			if _, err := newConn.Write([]byte{5, 0}); err != nil {
				t.Errorf("write failed: %v", err)
				return
			}

			// ver cmd rsv atyp dst.addr dst.port
			if _, err := io.ReadFull(newConn, buf[:10]); err != nil {
				t.Errorf("read failed: %v", err)
				return
			}
			if want := []byte{5, 1, 0, 1}; !bytes.Equal(want, buf[:len(want)]) {
				t.Errorf("read %x, want %x", buf[:len(want)], want)
				return
			}
			buf[1] = 0
			if _, err := newConn.Write(buf[:10]); err != nil {
				t.Errorf("write failed: %v", err)
				return
			}

			// 提取ip
			ip := net.IP(buf[4:8])
			port := binary.BigEndian.Uint16(buf[8:10])

			c2, err := net.DialTCP("tcp", nil, &net.TCPAddr{IP: ip, Port: int(port)})
			if err != nil {
				t.Errorf("dial failed; %v", err)
				return
			}
			defer c2.Close()
			done := make(chan struct{})
			go func() {
				io.Copy(newConn, c2)
				close(done)
			}()
			io.Copy(c2, newConn)
			<-done
		}()

		got := make([]byte, 0, 128)
		url := strings.ReplaceAll(ts.URL, "http", "ws")
		c, err := DialConf(url, ClientOptionToConf(WithClientDialFunc(func() (Dialer, error) {
			return proxy.SOCKS5("tcp", proxyAddr.Addr().String(), nil, nil)
		}), WithClientOnMessageFunc(func(c *Conn, o Opcode, b []byte) {
			got = append(got, b...)
			c.Close()
		})))
		if err != nil {
			t.Error(err)
		}

		data := []byte("hello world")
		c.WriteMessage(Binary, data)
		c.ReadLoop()

		t.Log("got", string(got), "want", string(data))
		if !bytes.Equal(got, data) {
			t.Errorf("got %s, want %s", got, data)
		}
	})
}
