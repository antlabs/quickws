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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SecWebSocketAcceptVal(t *testing.T) {
	need := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	got := secWebSocketAcceptVal("dGhlIHNhbXBsZSBub25jZQ==")
	assert.Equal(t, need, got)
}
