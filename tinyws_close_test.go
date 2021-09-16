package tinyws

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// close fail服务
func newServerCloseFail(t *testing.T, index *int) *httptest.Server {
	errMsg := []error{
		ErrMaxControlFrameSize,
		ErrClosePayloadTooSmall,
		ErrClosePayloadTooSmall,
		ErrTextNotUTF8,
		ErrCloseValue,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := Upgrade(w, r)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		defer c.Close()

		_, op, err := c.ReadTimeout(3 * time.Second)
		assert.Error(t, err, fmt.Sprintf("index:%d", *index))
		assert.Equal(t, op, Close, fmt.Sprintf("index:%d", *index))
		assert.Equal(t, err, errMsg[*index], fmt.Sprintf("index:%d", *index))
		(*index)++
	}))

	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
	return ts
}

// 测试close错误的函数
func Test_Fail_Close(t *testing.T) {
	index := 0
	ts := newServerCloseFail(t, &index)
	defer ts.Close()

	for _, d := range [][]byte{
		bytes.Repeat([]byte("123456789123456789"), 20),
		[]byte(""),
		[]byte("1"),
		[]byte("12中")[:3],
		statusCodeToBytes(StatusCode(25555)),
	} {
		conn, err := Dial(ts.URL)
		assert.NoError(t, err)
		conn.WriteTimeout(Close, d, 3*time.Second)
		conn.Close()
	}
}

func newServerSucessClose(t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := Upgrade(w, r)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		defer c.Close()

		_, _, err = c.ReadTimeout(3 * time.Second)
		assert.Error(t, err)
		fmt.Printf("%v\n", err)
	}))
	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
	return ts
}

// 测试正常的close流程
func Test_Success_Close(t *testing.T) {
	ts := newServerSucessClose(t)
	conn, err := Dial(ts.URL)

	err = conn.WriteCloseTimeout(NormalClosure, 3*time.Second)

	assert.NoError(t, err)
}
