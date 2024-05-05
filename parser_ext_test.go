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
	"reflect"
	"testing"
)

func Test_parseExtensions(t *testing.T) {
	type args struct {
		header http.Header
	}
	tests := []struct {
		name string
		args args
		want []pair
	}{
		{
			name: "Test_parseExtensions",
			args: args{header: http.Header{"Sec-Websocket-Extensions": {"permessage-deflate; client_max_window_bits=15; server_max_window_bits=15; client_no_context_takeover; server_no_context_takeover"}}},
			want: []pair{{key: "permessage-deflate", val: ""}, {key: "client_max_window_bits", val: "15"}, {key: "server_max_window_bits", val: "15"}, {key: "client_no_context_takeover", val: ""}, {key: "server_no_context_takeover", val: ""}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseExtensions(tt.args.header); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseExtensions() = %v, want %v", got, tt.want)
			}
		})
	}
}
