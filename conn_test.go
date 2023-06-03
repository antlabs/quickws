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

// Copyright 2021-2022 antlabs. All rights reserved.
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

type echoCtrl struct {
	readNeedSuccess  bool
	writeNeedSuccess bool
}

// 测试本文件内函数专用echo服务
// func createConnTestEcho(t *testing.T, data []byte, ctrl echoCtrl) *httptest.Server {
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		c, err := Upgrade(w, r)
// 		assert.NoError(t, err)
// 		if err != nil {
// 			return
// 		}

// 		defer c.Close()

// 		all, op, err := c.ReadTimeout(3 * time.Second)
// 		if ctrl.readNeedSuccess {
// 			assert.NoError(t, err)
// 		} else {
// 			assert.Error(t, err)
// 		}

// 		if err != nil {
// 			return
// 		}

// 		assert.Equal(t, len(all), 0)

// 		err = c.WriteTimeout(op, all, 3*time.Second)
// 		if ctrl.writeNeedSuccess {
// 			assert.NoError(t, err)
// 		} else {
// 			assert.Error(t, err)
// 		}
// 	}))

// 	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
// 	return ts
// }

// 测试rsv1 rsv2 rsv3
// func Test_Rsv123_Fail(t *testing.T) {
// 	// 1.创建测试服务
// 	ts := createConnTestEcho(t, nil, echoCtrl{readNeedSuccess: false})

// 	for _, f := range []frame{
// 		{frameHeader: frameHeader{rsv1: true, mask: true}},
// 		{frameHeader: frameHeader{rsv2: true, mask: true}},
// 		{frameHeader: frameHeader{rsv3: true, mask: true}},
// 	} {

// 		// 2.连接
// 		c, err := Dial(ts.URL)
// 		assert.NoError(t, err)
// 		if err != nil {
// 			return
// 		}

// 		if f.mask {
// 			newMask(f.maskValue[:])
// 		}

// 		if err := writeFrame(c.w, f); err != nil {
// 			assert.NoError(t, err)
// 			return
// 		}

// 		err = c.w.Flush()
// 		assert.NoError(t, err)
// 		c.Close()
// 	}
// }
