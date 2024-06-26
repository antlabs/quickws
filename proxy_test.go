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
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type testServer struct {
	path       string
	rawQuery   string
	requestURL string
	subprotos  []string
	*testing.T
}

func newTestServer(t *testing.T) *testServer {
	return &testServer{path: "/test", rawQuery: "a=1&b=2", requestURL: "/test?a=1&b=2", T: t, subprotos: []string{"proto1", "proto2"}}
}

func (t *testServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != t.path {
		t.Errorf("path error: %s", r.URL.Path)
		return
	}

	if r.URL.RawQuery != t.rawQuery {
		t.Errorf("raw query error: %s", r.URL.RawQuery)
		return
	}

	sub := subProtocol(r.Header.Get("Sec-Websocket-Protocol"), &Config{subProtocols: t.subprotos})
	if sub != "proto1" {
		t.Errorf("sub protocol error: (%s)", sub)
		return
	}

	conn, err := Upgrade(w, r, WithServerOnMessageFunc(func(c *Conn, o Opcode, b []byte) {
		err := c.WriteMessage(o, b)
		if err != nil {
			t.Error(err)
			return
		}
	}))
	if err != nil {
		t.Error(err)
		return
	}
	_ = conn.ReadLoop()
}

func (t *testServer) clientSend(c *Conn) {
	_ = c.WriteMessage(Text, []byte("hello world"))
}

func HTTPToWS(u string) string {
	return strings.ReplaceAll(u, "http://", "ws://")
}

func WsToHTTP(u string) string {
	return strings.ReplaceAll(u, "ws://", "http://")
}

func Test_Proxy(t *testing.T) {
	t.Run("test proxy dial.1", func(t *testing.T) {
		connect := false
		s := newTestServer(t)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Logf("method: %s, url: %s", r.Method, r.URL.String())
			if r.Method == http.MethodConnect {
				connect = true
				w.WriteHeader(http.StatusOK)
				return
			}

			if !connect {
				t.Error("test proxy dial fail: not connect")
				http.Error(w, "not connect", http.StatusMethodNotAllowed)
				return
			}
			s.ServeHTTP(w, r)
		}))

		defer ts.Close()

		proxy := func(*http.Request) (*url.URL, error) {
			return url.Parse(HTTPToWS(ts.URL))
		}

		got := make(chan string, 1)
		dstURL := HTTPToWS(ts.URL + s.requestURL)
		con, err := Dial(dstURL,
			WithClientProxyFunc(proxy),
			WithClientSubprotocols(s.subprotos),
			WithClientOnMessageFunc(func(c *Conn, o Opcode, b []byte) {
				got <- string(b)
			}))
		if err != nil {
			t.Error(err)
			return
		}
		con.StartReadLoop()
		s.clientSend(con)

		defer con.Close()
		gotValue := <-got
		if gotValue != "hello world" {
			t.Errorf("got: %s, want: %s", gotValue, "hello world")
			return
		}
	})
}

func Test_httpProxy_Dial(t *testing.T) {
	type fields struct {
		proxyAddr *url.URL
		dial      func(network, addr string, timeout time.Duration) (c net.Conn, err error)
	}
	type args struct {
		network string
		addr    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantC   net.Conn
		wantErr bool
	}{
		// 0
		{
			name: "No proxy address",
			fields: fields{
				proxyAddr: nil,
				dial: func(network, addr string, timeout time.Duration) (c net.Conn, err error) {
					// Simulate successful dialing
					return &net.TCPConn{}, errors.New("fail")
				},
			},
			args: args{
				network: "tcp",
				addr:    "example.com:80",
			},
			wantC:   &net.TCPConn{},
			wantErr: true,
		},
		// 1
		{
			name: "Proxy address",
			fields: fields{
				proxyAddr: &url.URL{Host: "1.2.3:8080", User: url.UserPassword("user", "password")},
				dial: func(network, addr string, timeout time.Duration) (c net.Conn, err error) {
					// Simulate successful dialing
					return &net.TCPConn{}, errors.New("fail")
				},
			},
			args: args{
				network: "tcp",
				addr:    "a.b.c:80",
			},
			wantC:   &net.TCPConn{},
			wantErr: true,
		},
		// 2
		{
			name: "Proxy address",
			fields: fields{
				proxyAddr: &url.URL{Host: "1.2.3:8080", User: url.UserPassword("user", "password")},
				dial: func(network, addr string, timeout time.Duration) (c net.Conn, err error) {
					// Simulate successful dialing
					return &net.TCPConn{}, nil
				},
			},
			args: args{
				network: "tcp",
				addr:    "a.b.c:80",
			},
			wantC:   &net.TCPConn{},
			wantErr: false,
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &httpProxy{
				proxyAddr:   tt.fields.proxyAddr,
				dialTimeout: tt.fields.dial,
			}
			_, err := h.dialTimeout(tt.args.network, tt.args.addr, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("index:%d, httpProxy.Dial() error = %v, wantErr %v", i, err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(gotC, tt.wantC) {
			// 	t.Errorf("httpProxy.Dial() = %v, want %v", gotC, tt.wantC)
			// }
		})
	}
}
