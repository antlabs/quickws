// Copyright 2021 guonaihong. All rights reserved.
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

// +build tinywstest

package tinyws

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	//"os"
	"testing"
	"time"
)

// echo测试服务
func echo(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	c, err := Upgrade(w, r, WithServerReplyPing(), WithCompression())
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	fmt.Printf("new conn:%p ", c)
	defer func() {
		c.Close()
		fmt.Printf("conn bye bye %p:%v\n", c, time.Since(now))
	}()

	for {
		now := time.Now()
		all, op, err := c.ReadTimeout(2 * time.Second)
		if err != nil {
			fmt.Printf("err = %v ", err)
			return
		}

		fmt.Printf("%p, read:%v ", c, time.Since(now))

		//os.Stdout.Write(all)
		now = time.Now()
		c.WriteTimeout(op, all, 3*time.Second)
		fmt.Printf("%p, write:%v ", c, time.Since(now))
	}
}

func Test_Server(t *testing.T) {
	http.HandleFunc("/", echo)

	assert.NoError(t, http.ListenAndServe(":9001", nil))
}
