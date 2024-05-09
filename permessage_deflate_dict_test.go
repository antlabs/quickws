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
	"reflect"
	"testing"
)

func Test_historyDict_Write(t *testing.T) {
	type fields struct {
		data      []byte
		ringthPos int
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name:   "case1, data is greater than dict",
			fields: fields{data: make([]byte, 3), ringthPos: 3},
			args:   args{data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
			want:   []byte{8, 9, 10},
		},
		{
			name:   "case1, data is greater than dict",
			fields: fields{data: make([]byte, 10), ringthPos: 9},
			args:   args{data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
			want:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name:   "case2, data is less than dict",
			fields: fields{data: []byte{1, 2, 3, 4, 0, 0}, ringthPos: 4},
			args:   args{data: []byte{7, 8, 9}},
			want:   []byte{2, 3, 4, 7, 8, 9},
		},
		{
			name:   "case2, data is less than dict",
			fields: fields{data: []byte{0, 0, 0, 0, 0, 0}, ringthPos: 0},
			args:   args{data: []byte{7}},
			want:   []byte{7},
		},
		{
			name:   "case2, data is less than dict",
			fields: fields{data: []byte{0, 0, 0, 0, 0, 0}, ringthPos: 6},
			args:   args{data: []byte{7}},
			want:   []byte{0, 0, 0, 0, 0, 7},
		},
		{
			name:   "case2, data is less than dict",
			fields: fields{data: []byte{1, 0, 0, 0, 0}, ringthPos: 1},
			args:   args{data: []byte{2, 3, 4, 5}},
			want:   []byte{1, 2, 3, 4, 5},
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &historyDict{
				data:      tt.fields.data,
				ringthPos: tt.fields.ringthPos,
			}
			w.Write(tt.args.data)
			if !reflect.DeepEqual(w.GetData(), tt.want) {
				t.Errorf("index:%d, historyDict.Write() = %v, want %v", i, w.GetData(), tt.want)
			}
		})
	}
}
