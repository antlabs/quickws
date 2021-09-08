package tinyws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newServerDecompressAndCompression(t *testing.T, data []byte) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := Upgrade(w, r, WithServerDecompressAndCompress())
		assert.NoError(t, err)
		if err != nil {
			return
		}

		defer c.Close()

		all, op, err := c.ReadTimeout(3 * time.Second)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		assert.Equal(t, all, data)

		err = c.WriteTimeout(op, all, 3*time.Second)
		assert.NoError(t, err)
	}))

	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
	return ts
}

func Test_Server_Compression(t *testing.T) {
	data := []byte("test data")
	ts := newServerDecompressAndCompression(t, data)
	c, err := Dial(ts.URL, WithDecompressAndCompress())
	assert.NoError(t, err)
	if err != nil {
		return
	}
	defer c.Close()

	err = c.WriteTimeout(Text, data, 3*time.Second)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	all, _, err := c.ReadTimeout(3 * time.Second)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, data, all)
}
