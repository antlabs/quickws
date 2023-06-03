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

// func newServerDecompressAndCompression(t *testing.T, data []byte) *httptest.Server {
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		c, err := Upgrade(w, r, WithServerDecompressAndCompress())
// 		assert.NoError(t, err)
// 		if err != nil {
// 			return
// 		}

// 		defer c.Close()

// 		all, op, err := c.ReadTimeout(3 * time.Second)
// 		assert.NoError(t, err)
// 		if err != nil {
// 			return
// 		}

// 		assert.Equal(t, all, data)

// 		err = c.WriteTimeout(op, all, 3*time.Second)
// 		assert.NoError(t, err)
// 	}))

// 	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
// 	return ts
// }

// func Test_Server_Compression(t *testing.T) {
// 	data := []byte("test data")
// 	ts := newServerDecompressAndCompression(t, data)
// 	c, err := Dial(ts.URL, WithDecompressAndCompress())
// 	assert.NoError(t, err)
// 	if err != nil {
// 		return
// 	}
// 	defer c.Close()

// 	err = c.WriteTimeout(Text, data, 3*time.Second)
// 	assert.NoError(t, err)
// 	if err != nil {
// 		return
// 	}
// 	all, _, err := c.ReadTimeout(3 * time.Second)
// 	assert.NoError(t, err)
// 	if err != nil {
// 		return
// 	}
// 	assert.Equal(t, data, all)
// }
