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

import "testing"

func Test_InitPayloadSize(t *testing.T) {
	t.Run("InitPayload", func(t *testing.T) {
		var c Config
		for i := 1; i < 32; i++ {
			c.windowsMultipleTimesPayloadSize = float32(i)
			if c.initPayloadSize() != i*(1024+14) {
				t.Errorf("initPayloadSize() = %d, want %d", c.initPayloadSize(), i*(1024+14))
			}
		}
	})
}
