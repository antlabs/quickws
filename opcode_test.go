package tinyws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Control(t *testing.T) {
	assert.False(t, Text.isControl())
	assert.False(t, Binary.isControl())
	assert.True(t, Ping.isControl())
	assert.True(t, Pong.isControl())
	assert.True(t, Close.isControl())
}
