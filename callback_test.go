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
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type testDefaultCallback struct {
	DefCallback
}

func Test_DefaultCallback(t *testing.T) {
	t.Run("local: default callback", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerCallback(&testDefaultCallback{}))
			if err != nil {
				t.Error(err)
				return
			}
			if c == nil {
				t.Error("conn is nil")
				return
			}
			defer c.Close()
			c.StartReadLoop()

			err = c.WriteMessage(Binary, []byte("hello"))
			if err != nil {
				t.Error(err)
				return
			}

			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientCallback(&testDefaultCallback{}))
		if err != nil {
			t.Error(err)
			return
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		select {
		case <-done:
		case <-time.After(1000 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})

	t.Run("global: default callback", func(t *testing.T) {
		run := int32(0)
		done := make(chan bool, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerCallback(&testDefaultCallback{}))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
			err = c.WriteMessage(Binary, []byte("hello"))
			atomic.AddInt32(&run, int32(1))
			done <- true
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientCallback(&testDefaultCallback{}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		con.WriteMessage(Binary, []byte("hello"))
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
		}
		if atomic.LoadInt32(&run) != 1 {
			t.Error("not run server:method fail")
		}
	})
}
