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
	"testing"

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
		{w: &failWriter{failCount: 4}, wantErr: true, args: args{r: &http.Request{Header: http.Header{}}, cnf: &Config{decompression: true}}},
		{w: &failWriter{failCount: 5}, wantErr: true, args: args{r: &http.Request{Header: http.Header{"Sec-Websocket-Protocol": []string{"token"}}}, cnf: &Config{decompression: true}}},
		{w: &failWriter{failCount: 6}, wantErr: true, args: args{r: &http.Request{Header: http.Header{"Sec-Websocket-Protocol": []string{"token"}}}, cnf: &Config{decompression: true}}},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := prepareWriteResponse(tt.args.r, tt.w, tt.args.cnf, deflate.PermessageDeflateConf{}); (err != nil) != tt.wantErr {
				t.Errorf("index:%d, prepareWriteResponse() error = %v, wantErr %v, %d", i, err, tt.wantErr, tt.w.(*failWriter).count)
				return
			}
		})
	}
}
