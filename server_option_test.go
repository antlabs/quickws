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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_ServerOption(t *testing.T) {
	t.Run("2.1 Subprotocol", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := Upgrade(w, r, WithServerSubprotocols([]string{"crud", "im"}))
			if err != nil {
				t.Error(err)
			}
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		h := make(http.Header)
		con, err := DialConf(url, ClientOptionToConf(WithClientBindHTTPHeader(&h), WithClientHTTPHeader(http.Header{
			"Sec-WebSocket-Protocol": []string{"crud"},
		})))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		if h["Sec-Websocket-Protocol"][0] != "crud" {
			t.Error("header fail")
		}
	})

	t.Run("2.2 Subprotocol", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := Upgrade(w, r, WithServerSubprotocols([]string{"crud", "im"}))
			if err != nil {
				t.Error(err)
			}
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		h := make(http.Header)
		con, err := Dial(url, WithClientBindHTTPHeader(&h), WithClientHTTPHeader(http.Header{
			"Sec-WebSocket-Protocol": []string{"crud"},
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		if h["Sec-Websocket-Protocol"][0] != "crud" {
			t.Error("header fail")
		}
	})

	t.Run("2.3 Subprotocol", func(t *testing.T) {
		upgrade := NewUpgrade(WithServerSubprotocols([]string{"crud", "im"}))
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := upgrade.Upgrade(w, r)
			if err != nil {
				t.Error(err)
			}
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		h := make(http.Header)
		con, err := Dial(url, WithClientBindHTTPHeader(&h), WithClientHTTPHeader(http.Header{
			"Sec-WebSocket-Protocol": []string{"crud"},
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		if h["Sec-Websocket-Protocol"][0] != "crud" {
			t.Error("header fail")
		}
	})
}
