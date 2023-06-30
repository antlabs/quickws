package quickws

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/antlabs/wsutil/opcode"
)

var (
	testBinaryMessage64kb = bytes.Repeat([]byte("1"), 65535)
	testTextMessage64kb   = bytes.Repeat([]byte("中"), 65535/len("中"))
	testBinaryMessage10   = bytes.Repeat([]byte("1"), 10)
)

type testMessageHandler struct {
	DefCallback
	t       *testing.T
	need    []byte
	callbed int32
	server  bool
	count   int
	got     chan []byte
	done    chan struct{}
	output  bool
}

func (t *testMessageHandler) OnMessage(c *Conn, op opcode.Opcode, msg []byte) {
	need := append([]byte(nil), t.need...)
	atomic.StoreInt32(&t.callbed, 1)
	if t.count == 0 {
		return
	}
	t.count--

	message := "#client"
	if t.server {
		message = "#server"
	}
	if t.output {
		// fmt.Printf(">>>>>%p %s, %#v\n", &msg, message, msg)
	}
	if len(msg) < 30 {
		if bytes.Equal(msg, need) {
			t.t.Logf(">>>>>%p %s, %#v\n", &msg, message, msg)
		}
	} else {

		md51 := md5.Sum(need)
		md52 := md5.Sum(msg)
		if bytes.Equal(md51[:], md52[:]) {
			t.t.Logf("md51 %x, md52 %x\n", md51, md52)
		}
	}
	err := c.WriteMessage(op, msg)
	if err != nil {
		t.t.Error(err)
	}
	if !t.server {
		// c.Close()
	}
}

func (t *testMessageHandler) OnClose(c *Conn, err error) {
	message := "#client.OnClose"
	if t.server {
		message = "#server.OnClose"
	}

	fmt.Printf("OnClose: %s:%s\n", message, err)
	if t.done != nil {
		close(t.done)
	}
}

func newServrEcho(t *testing.T, data []byte, output bool) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := Upgrade(w, r,
			WithServerCallback(&testMessageHandler{t: t, need: data, server: true, count: -1, output: true}),
		)
		if err != nil {
			t.Error(err)
			return
		}

		c.ReadLoop()
	}))

	ts.URL = "ws" + strings.TrimPrefix(ts.URL, "http")
	return ts
}

func Test_ReadMessage10(t *testing.T) {
	ts := newServrEcho(t, testBinaryMessage10, true)
	client := &testMessageHandler{t: t, need: append([]byte(nil), testBinaryMessage10...), count: 1, done: make(chan struct{}), output: true}
	c, err := Dial(ts.URL, WithClientCallback(client))
	if err != nil {
		t.Error(err)
		return
	}
	go c.ReadLoop()

	tmp := append([]byte(nil), testBinaryMessage10...)

	err = c.WriteMessage(Binary, tmp)
	if err != nil {
		t.Error(err)
		return
	}
	// <-client.done
	time.Sleep(time.Second / 2)
	if atomic.LoadInt32(&client.callbed) != 1 {
		t.Error("not callbed")
	}
}

func Test_ReadMessage64k(t *testing.T) {
	ts := newServrEcho(t, testBinaryMessage64kb, false)
	client := &testMessageHandler{t: t, need: append([]byte(nil), testBinaryMessage64kb...), count: 1}
	c, err := Dial(ts.URL, WithClientCallback(client))
	if err != nil {
		t.Error(err)
		return
	}
	go c.ReadLoop()

	tmp := append([]byte(nil), testBinaryMessage64kb...)
	err = c.WriteMessage(Binary, tmp)
	time.Sleep(time.Second / 2)
	if err != nil {
		t.Error(err)
		return
	}

	if atomic.LoadInt32(&client.callbed) != 1 {
		t.Error("not callbed")
	}
}

func Test_ReadMessage64k_Text(t *testing.T) {
	ts := newServrEcho(t, testTextMessage64kb, false)
	client := &testMessageHandler{t: t, need: append([]byte(nil), testTextMessage64kb...), count: 1}
	c, err := Dial(ts.URL, WithClientCallback(client))
	if err != nil {
		t.Error(err)
		return
	}

	go c.ReadLoop()

	tmp := append([]byte(nil), testTextMessage64kb...)
	err = c.WriteMessage(Text, tmp)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(time.Second / 2)

	if atomic.LoadInt32(&client.callbed) != 1 {
		t.Error("not callbed")
	}
}
