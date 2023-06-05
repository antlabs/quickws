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
	"testing"
)

func Benchmark_GetPool1k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := getBytes(1024)
		for i := range b {
			b[i] = 1
		}
		putBytes(b)
	}
}

func Benchmark_Make1k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := make([]byte, 1024)
		for i := range b {
			b[i] = 1
		}
	}
}

func Benchmark_GetPool2k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := getBytes(1024 * 2)
		for i := range b {
			b[i] = 1
		}
		putBytes(b)
	}
}

func Benchmark_Make2k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := make([]byte, 1024*2)
		for i := range b {
			b[i] = 1
		}
	}
}

func Benchmark_GetPool3k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := getBytes(1024 * 3)
		for i := range b {
			b[i] = 1
		}
		putBytes(b)
	}
}

func Benchmark_Make3k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := make([]byte, 1024*3)
		for i := range b {
			b[i] = 1
		}
	}
}

func Benchmark_GetPool4k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := getBytes(1024 * 4)
		for i := range b {
			b[i] = 1
		}
		putBytes(b)
	}
}

func Benchmark_Make4k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := make([]byte, 1024*4)
		for i := range b {
			b[i] = 1
		}
	}
}

func Benchmark_GetPool8k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := getBytes(1024 * 8)
		for i := range b {
			b[i] = 1
		}
		putBytes(b)
	}
}

func Benchmark_Make8k(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := make([]byte, 1024*8)
		for i := range b {
			b[i] = 1
		}
	}
}
