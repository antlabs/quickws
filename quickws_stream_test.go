// Copyright 2021-2023 antlabs. All rights reserved.
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

// // 测试stream 场景mock服务
// func createStreamServer(t *testing.T, max int) *httptest.Server {
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		c, err := Upgrade(w, r)
// 		assert.NoError(t, err)
// 		if err != nil {
// 			return
// 		}

// 		defer func() {
// 			time.Sleep(time.Second / 10)
// 			c.Close()
// 		}()

// 		data := make(chan []byte)
// 		// 读流
// 		go func() {
// 			defer func() { close(data) }()
// 			for i := 0; i < max; i++ {
// 				all, _, err := c.ReadTimeout(3 * time.Second)
// 				if err != nil {
// 					assert.NoError(t, err)
// 					return
// 				}

// 				// 检测
// 				i2, err := strconv.Atoi(string(all))
// 				assert.NoError(t, err)
// 				if err != nil {
// 					c.Close()
// 					return
// 				}
// 				assert.Equal(t, i, i2)

// 				data <- all
// 			}
// 		}()

// 		// 写流
// 		for d := range data {
// 			err = c.WriteTimeout(Binary, d, 3*time.Second)
// 			assert.NoError(t, err)
// 		}
// 		assert.NoError(t, c.WriteTimeout(Binary, []byte("stop"), 3*time.Second))
// 		assert.NoError(t, c.WriteTimeout(Close, nil, 3*time.Second))
// 	}))

// 	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
// 	return ts
// }

// // 测试流并发读写
// func Test_Stream_RW(t *testing.T) {

// 	max := 10000
// 	ts := createStreamServer(t, max)

// 	conn, err := Dial(ts.URL)

// 	assert.NoError(t, err)
// 	if err != nil {
// 		return
// 	}

// 	// 写数据
// 	done := make(chan struct{}, 1)
// 	go func() {
// 		i := int64(0)
// 		defer func() {
// 			done <- struct{}{}
// 		}()
// 		for ; ; i++ {
// 			d := strconv.AppendInt([]byte{}, i, 10)
// 			err := conn.WriteTimeout(Binary, d, 3*time.Second)
// 			if err != nil {
// 				conn.Close()
// 				return
// 			}
// 		}

// 	}()

// 	i := 0
// 	stop := false
// 	// 读数据
// 	defer func() {
// 		<-done
// 		assert.True(t, stop)
// 		assert.Equal(t, i, max)
// 	}()
// 	for i = 0; ; i++ {
// 		d, _, err := conn.ReadTimeout(3 * time.Second)
// 		if err != nil {
// 			conn.Close()
// 			return
// 		}

// 		if bytes.Equal(d, []byte("stop")) {
// 			stop = true
// 			conn.Close()
// 			return
// 		}

// 		i2, err := strconv.Atoi(string(d))
// 		assert.NoError(t, err)
// 		if err != nil {
// 			conn.Close()
// 			return
// 		}

// 		assert.Equal(t, i, i2)
// 		//assert.Equal(t, d[0]+1, d[len(d)-1])
// 	}
// }
