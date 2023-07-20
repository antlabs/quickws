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
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	//"os"
)

type echoHandler struct{}

func (e *echoHandler) OnOpen(c *Conn) {
}

func (e *echoHandler) OnMessage(c *Conn, op Opcode, msg []byte) {
	// if err := c.WriteTimeout(op, msg, 3*time.Second); err != nil {
	// 	fmt.Println("write fail:", err)
	// }
	if err := c.WriteMessage(op, msg); err != nil {
		fmt.Println("write fail:", err)
	}
}

func (e *echoHandler) OnClose(c *Conn, err error) {
}

type echoClientHandler struct {
	DefCallback
	Count int
}

func (e *echoClientHandler) OnMessage(c *Conn, op Opcode, msg []byte) {
	e.Count++
	if e.Count == 100 {
		c.Close()
	}
	c.WriteMessage(op, msg)
}

func newProfileServrEcho(t *testing.T, data []byte) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := Upgrade(w, r, WithServerCallback(&echoHandler{}))
		if err != nil {
			t.Error(err)
			return
		}

		c.ReadLoop()
	}))

	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
	return ts
}

// 跑profile的echo服务
func Test_ServerProfile(t *testing.T) {
	payload := make([]byte, 1024)
	for i := 0; i < len(payload); i++ {
		payload[i] = byte(i)
	}

	maxGo := 10
	ts := newProfileServrEcho(t, payload)
	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(maxGo)
	for i := 0; i < maxGo; i++ {
		go func() {
			defer wg.Done()
			c, err := Dial(ts.URL, WithClientCallback(&echoClientHandler{}))
			if err != nil {
				t.Error(err)
				return
			}
			c.WriteMessage(Binary, payload)
			c.ReadLoop()
		}()
	}
}

func Test_Upgrade(t *testing.T) {
	r, err := http.NewRequest("GET", "http://test.com", nil)
	if err != nil {
		t.Error(err)
		return
	}

	var out bytes.Buffer
	prepareWriteResponse(r, &out, &Config{})
	fmt.Printf("%s\n %d", out.Bytes(), out.Len())
}
