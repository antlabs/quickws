package tinyws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	globalTestData = []byte("hello world")
)

func newEchoWebSocketServer(t *testing.T, testData []byte) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := Upgrade(w, r)
		assert.NoError(t, err)
		if err != nil {
			return
		}
		all, op, err := conn.ReadMessage()
		assert.NoError(t, err)
		assert.Equal(t, string(all), string(testData))
		err = conn.WriteMessage(op, all)
		assert.NoError(t, err)
	}))
	s.URL = "ws" + strings.TrimPrefix(s.URL, "http")
	return s
}

func sendAndRecv(c *Conn, t *testing.T, testData []byte) {
	err := c.WriteMessage(Text, testData)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	all, op, err := c.ReadMessage()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, string(all), string(testData))
	assert.Equal(t, op, Text)

}

func TestDial(t *testing.T) {
	ts := newEchoWebSocketServer(t, globalTestData)
	c, err := Dial(ts.URL)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	defer c.Close()

	sendAndRecv(c, t, globalTestData)
}
