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

// 管理dict的数据结构
type historyDict struct {
	data      []byte
	ringthPos int
}

func (w *historyDict) InitHistoryDict(ringthPos int) {
	w.data = make([]byte, ringthPos)
	w.ringthPos = ringthPos
}

func (w *historyDict) Write(data []byte) {

	// 第1种情况。
	// 新的数据大于dict的长度, 直接覆盖
	if len(data) >= len(w.data) {
		copy(w.data, data[len(data)-len(w.data):])
		w.ringthPos = len(w.data)
		return
	}

	// 第2种情况。
	// 剩余空间已经不够放旧的数据了，这时候需要将旧的数据往前移动
	// w.data=[A B C D E F G H I J 0 0]
	// w.ringthPos=9
	// data       [3 4 5 6 7 8 9]
	// w.data [I J 3 4 5 6 7 8 9 0]
	if len(data) > len(w.data)-w.ringthPos {
		oldPos := len(w.data) - len(data)
		moveSize := w.ringthPos - oldPos
		copy(w.data, w.data[moveSize:w.ringthPos])
		w.ringthPos = oldPos
		// 第二种情况， 旧数据从尾巴移到头部就变成第三种情况，所以这里不会return
	}

	// 第3情况，空间是够的, 直接放到后面
	copy(w.data[w.ringthPos:], data)
	w.ringthPos += len(data)
}

func (w *historyDict) GetData() []byte {
	return w.data[:w.ringthPos]
}
