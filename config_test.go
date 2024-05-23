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
	"net/http"
	"net/url"
	"testing"
	"time"
)

func Test_InitPayloadSize(t *testing.T) {
	t.Run("InitPayload", func(t *testing.T) {
		var c Config
		for i := 1; i < 32; i++ {
			c.windowsMultipleTimesPayloadSize = float32(i)
			if c.initPayloadSize() != i*(1024+14) {
				t.Errorf("initPayloadSize() = %d, want %d", c.initPayloadSize(), i*(1024+14))
			}
		}
	})
}

func TestConfig_defaultSetting(t *testing.T) {
	type fields struct {
		Callback   Callback
		tcpNoDelay bool
		replyPing  bool
		// decompression                   bool
		// compression                     bool
		ignorePong                      bool
		disableBufioClearHack           bool
		utf8Check                       func([]byte) bool
		readTimeout                     time.Duration
		windowsMultipleTimesPayloadSize float32
		bufioMultipleTimesPayloadSize   float32
		parseMode                       parseMode
		maxDelayWriteNum                int32
		delayWriteInitBufferSize        int32
		maxDelayWriteDuration           time.Duration
		subProtocols                    []string
		dialFunc                        func() (Dialer, error)
		proxyFunc                       func(*http.Request) (*url.URL, error)
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "fail", fields: fields{
			dialFunc:  func() (Dialer, error) { return nil, nil },
			proxyFunc: func(*http.Request) (*url.URL, error) { return nil, nil },
		}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Callback:                        tt.fields.Callback,
				tcpNoDelay:                      tt.fields.tcpNoDelay,
				replyPing:                       tt.fields.replyPing,
				ignorePong:                      tt.fields.ignorePong,
				disableBufioClearHack:           tt.fields.disableBufioClearHack,
				utf8Check:                       tt.fields.utf8Check,
				readTimeout:                     tt.fields.readTimeout,
				windowsMultipleTimesPayloadSize: tt.fields.windowsMultipleTimesPayloadSize,
				bufioMultipleTimesPayloadSize:   tt.fields.bufioMultipleTimesPayloadSize,
				parseMode:                       tt.fields.parseMode,
				maxDelayWriteNum:                tt.fields.maxDelayWriteNum,
				delayWriteInitBufferSize:        tt.fields.delayWriteInitBufferSize,
				maxDelayWriteDuration:           tt.fields.maxDelayWriteDuration,
				subProtocols:                    tt.fields.subProtocols,
				dialFunc:                        tt.fields.dialFunc,
				proxyFunc:                       tt.fields.proxyFunc,
			}
			if err := c.defaultSetting(); (err != nil) != tt.wantErr {
				t.Errorf("Config.defaultSetting() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
