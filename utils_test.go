package tinyws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SecWebSocketAcceptVal(t *testing.T) {
	need := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	got := secWebSocketAcceptVal("dGhlIHNhbXBsZSBub25jZQ==")
	assert.Equal(t, need, got)
}
