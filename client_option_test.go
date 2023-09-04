// Copyright 2021-2023 antlabs. All rights reserved.
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
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
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
}
