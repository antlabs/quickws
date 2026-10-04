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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/antlabs/wsutil/deflate"
)

type failWriter struct {
	count     int
	failCount int
}

func (f *failWriter) Write(p []byte) (n int, err error) {
	f.count++
	if f.count == f.failCount {
		return 0, errors.New("fail")
	}
	return len(p), nil
}
func Test_writeHeaderVal(t *testing.T) {
	type args struct {
		val []byte
	}
	tests := []struct {
		w       io.Writer
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{w: &failWriter{failCount: 1}, wantErr: true},
		{w: &failWriter{failCount: 2}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeHeaderVal(tt.w, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("writeHeaderVal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_HttpGet(t *testing.T) {
	h := make(http.Header)
	h.Set(strGetSecWebSocketProtocolKey, "token")
	fmt.Printf("%#v\n", h)
	if h.Get(strGetSecWebSocketProtocolKey) == "" {
		panic("error")
	}
}
func Test_prepareWriteResponse(t *testing.T) {
	type args struct {
		r   *http.Request
		cnf *Config
	}
	tests := []struct {
		w       io.Writer
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{w: &failWriter{failCount: 1}, wantErr: true, args: args{r: &http.Request{Header: http.Header{}}, cnf: &Config{}}},
		{w: &failWriter{failCount: 2}, wantErr: true, args: args{r: &http.Request{Header: http.Header{}}, cnf: &Config{}}},
		{w: &failWriter{failCount: 3}, wantErr: true, args: args{r: &http.Request{Header: http.Header{}}, cnf: &Config{}}},
		{w: &failWriter{failCount: 4}, wantErr: true, args: args{r: &http.Request{Header: http.Header{}}, cnf: &Config{PermessageDeflateConf: deflate.PermessageDeflateConf{Decompression: true, Compression: true}}}},
		// {w: &failWriter{failCount: 5}, wantErr: true, args: args{r: &http.Request{Header: http.Header{}}, cnf: &Config{PermessageDeflateConf: deflate.PermessageDeflateConf{Decompression: true, Compression: true}}}},
		// {w: &failWriter{failCount: 6}, wantErr: true, args: args{r: &http.Request{Header: http.Header{}}, cnf: &Config{PermessageDeflateConf: deflate.PermessageDeflateConf{Decompression: true, Compression: true}}}},
		// {w: &failWriter{failCount: 7}, wantErr: true, args: args{r: &http.Request{Header: http.Header{"Sec-Websocket-Protocol": []string{"token"}}}, cnf: &Config{PermessageDeflateConf: deflate.PermessageDeflateConf{Decompression: true, Compression: true}}}},
		// {w: &failWriter{failCount: 8}, wantErr: true, args: args{r: &http.Request{Header: http.Header{"Sec-Websocket-Protocol": []string{"token"}}}, cnf: &Config{PermessageDeflateConf: deflate.PermessageDeflateConf{Decompression: true, Compression: true}}}},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := prepareWriteResponse(tt.args.r, tt.w, tt.args.cnf, deflate.PermessageDeflateConf{}); (err != nil) != tt.wantErr {
				t.Errorf("index:%d, prepareWriteResponse() error = %v, wantErr %v, count= %d", i, err, tt.wantErr, tt.w.(*failWriter).count)
				return
			}
		})
	}
}

func Test_Handshake_ConnectionTokens(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		valid  bool
	}{
		{name: "single", values: []string{"Upgrade"}, valid: true},
		{name: "lowercase", values: []string{"upgrade"}, valid: true},
		{name: "last token", values: []string{"keep-alive, Upgrade"}, valid: true},
		{name: "first token", values: []string{"Upgrade, keep-alive"}, valid: true},
		{name: "whitespace", values: []string{"keep-alive ,\t UpGrAdE\t"}, valid: true},
		{name: "last field", values: []string{"keep-alive", "Upgrade"}, valid: true},
		{name: "first field", values: []string{"Upgrade", "keep-alive"}, valid: true},
		{name: "empty list member", values: []string{", Upgrade, "}, valid: true},
		{name: "missing"},
		{name: "empty", values: []string{""}},
		{name: "other token", values: []string{"keep-alive"}},
		{name: "suffix", values: []string{"keep-alive, Upgraded"}},
		{name: "prefix", values: []string{"xUpgrade"}},
		{name: "quoted", values: []string{"\"Upgrade\""}},
		{name: "missing comma", values: []string{"keep-alive Upgrade"}},
		{name: "non HTTP whitespace", values: []string{"\u00a0Upgrade\u00a0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := http.Header{
				"Upgrade":               []string{"websocket"},
				"Connection":            tt.values,
				"Sec-Websocket-Key":     []string{"dGhlIHNhbXBsZSBub25jZQ=="},
				"Sec-Websocket-Version": []string{"13"},
				"Sec-Websocket-Accept":  []string{"s3pPLMBiTxaQ9kYGzzhZRbK+xOo="},
			}
			t.Run("server", func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
				req.Header = header
				_, err := checkRequest(req)
				if tt.valid && err != nil {
					t.Fatalf("valid Connection header rejected: %v", err)
				}
				if !tt.valid && !errors.Is(err, ErrConnectionFieldValue) {
					t.Fatalf("got %v, want ErrConnectionFieldValue", err)
				}
			})
			t.Run("client", func(t *testing.T) {
				rsp := &http.Response{StatusCode: http.StatusSwitchingProtocols, Header: header}
				err := (&DialOption{}).validateRsp(rsp, header.Get("Sec-WebSocket-Key"))
				if tt.valid && err != nil {
					t.Fatalf("valid Connection header rejected: %v", err)
				}
				if !tt.valid && !errors.Is(err, ErrConnectionFieldValue) {
					t.Fatalf("got %v, want ErrConnectionFieldValue", err)
				}
			})
		})
	}
}

func Test_Client_ConnectionTokens(t *testing.T) {
	for _, values := range [][]string{{"keep-alive, Upgrade"}, {"keep-alive", "UpGrAdE"}} {
		for _, useConf := range []bool{false, true} {
			t.Run(fmt.Sprintf("%v/DialConf=%v", values, useConf), func(t *testing.T) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Upgrade", "websocket")
					for _, value := range values {
						w.Header().Add("Connection", value)
					}
					w.Header().Set("Sec-WebSocket-Accept", secWebSocketAcceptVal(r.Header.Get("Sec-WebSocket-Key")))
					w.WriteHeader(http.StatusSwitchingProtocols)
				}))
				defer ts.Close()
				rawURL := strings.Replace(ts.URL, "http://", "ws://", 1)
				var c *Conn
				var err error
				if useConf {
					c, err = DialConf(rawURL, ClientOptionToConf())
				} else {
					c, err = Dial(rawURL)
				}
				if err != nil {
					t.Fatal(err)
				}
				c.Close()
			})
		}
	}
}

func Test_Server_ConnectionTokens(t *testing.T) {
	for _, values := range [][]string{{"keep-alive, Upgrade"}, {"keep-alive", "UpGrAdE"}} {
		for _, useConf := range []bool{false, true} {
			t.Run(fmt.Sprintf("%v/NewUpgrade=%v", values, useConf), func(t *testing.T) {
				serverErr := make(chan error, 1)
				upgrade := NewUpgrade()
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var c *Conn
					var err error
					if useConf {
						c, err = upgrade.Upgrade(w, r)
					} else {
						c, err = Upgrade(w, r)
					}
					serverErr <- err
					if c != nil {
						c.Close()
					}
				}))
				defer ts.Close()
				req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Upgrade", "websocket")
				req.Header["Connection"] = values
				req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
				req.Header.Set("Sec-WebSocket-Version", "13")
				client := &http.Client{Timeout: time.Second}
				rsp, err := client.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer rsp.Body.Close()
				if rsp.StatusCode != http.StatusSwitchingProtocols {
					t.Fatalf("got HTTP %d, want HTTP 101", rsp.StatusCode)
				}
				if err := <-serverErr; err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
