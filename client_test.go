package tinyws

import (
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 通用的echo服务
func newServerEcho(t *testing.T, data []byte) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := Upgrade(w, r)
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

		assert.Equal(t, len(all), 0)

		err = c.WriteTimeout(op, all, 3*time.Second)
		assert.NoError(t, err)
	}))

	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
	return ts
}

// 新建压缩的mock服务
func newServerCompression(t *testing.T, data []byte) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := Upgrade(w, r, WithServerDecompression())
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

// ping 1.新建测试ping的服务
func newServerPing(t *testing.T, data []byte) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := Upgrade(w, r)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		defer c.Close()

		// 读取Ping
		_, op, err := c.ReadTimeout(3 * time.Second)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		assert.Equal(t, op, Ping)

		// 回复Pong
		err = c.WriteTimeout(Pong, nil, 3*time.Second)
		assert.NoError(t, err)
	}))

	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
	return ts
}

// ping 2.新建测试ping的服务-自动回复pong消息
func newServerPingReply(t *testing.T, data []byte) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := Upgrade(w, r, WithServerReplyPing())
		assert.NoError(t, err)
		if err != nil {
			return
		}

		defer c.Close()

		_, op, err := c.ReadTimeout(time.Second / 5)
		assert.Equal(t, op, Close)
		assert.Error(t, err)
	}))

	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
	return ts
}

// pong 1.新建测试pong echo的服务
func newServerPong(t *testing.T) *httptest.Server {
	return newServerEcho(t, []byte{})
}

// 测试压缩和解压缩 1.
func Test_Client_Compression(t *testing.T) {
	data := []byte("test data")
	ts := newServerCompression(t, data)
	defer ts.Close()
	c, err := Dial(ts.URL, WithDecompression(), WithCompression())
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

// 测试压缩和解压缩 2.
func Test_Client_Compression2(t *testing.T) {
	data := []byte("test data")
	ts := newServerCompression(t, data)
	defer ts.Close()

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

// 测试ping 自动回复
func Test_Client_Ping(t *testing.T) {
	tsSlice := []*httptest.Server{newServerPing(t, nil), newServerPingReply(t, nil)}
	defer func() {
		for _, ts := range tsSlice {
			ts.Close()
		}
		time.Sleep(time.Second / 100)
	}()

	for _, ts := range tsSlice {
		c, err := Dial(ts.URL)
		assert.NoError(t, err)
		if err != nil {
			return
		}
		defer c.Close()

		// 写ping
		err = c.WriteTimeout(Ping, nil, 3*time.Second)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		// 应该读到Pong
		_, op, err := c.ReadTimeout(3 * time.Second)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		data := make([]byte, 5)
		binary.BigEndian.PutUint16(data, 1000)
		copy(data[2:], "bye")

		c.WriteMessage(Close, data)
		// 检查下是否是Pong
		assert.Equal(t, op, Pong)
	}
}

// echo pong消息
func Test_Client_Pong(t *testing.T) {
	ts := newServerPong(t)
	defer ts.Close()

	c, err := Dial(ts.URL)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	defer c.Close()

	err = c.WriteTimeout(Pong, nil, 3*time.Second)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	data, op, err := c.ReadTimeout(3 * time.Second)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, op, Pong)
	assert.Equal(t, len(data), 0)
}
