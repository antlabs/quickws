package quickws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type testServerOptionReadTimeout struct {
	err chan error
	run int32
}

func (defcallback *testServerOptionReadTimeout) OnOpen(_ *Conn) {
}

func (defcallback *testServerOptionReadTimeout) OnMessage(_ *Conn, _ Opcode, _ []byte) {
}

func (defcallback *testServerOptionReadTimeout) OnClose(c *Conn, err error) {
	atomic.AddInt32(&defcallback.run, int32(1))
	defcallback.err <- err
}

func Test_ServerOption(t *testing.T) {
	t.Run("1.ServerOption: WithServerReadTimeout", func(t *testing.T) {
		var tsort testServerOptionReadTimeout

		tsort.err = make(chan error, 1)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := Upgrade(w, r, WithServerCallback(&tsort), WithServerReadTimeout(time.Millisecond*30))
			if err != nil {
				t.Error(err)
			}
			c.StartReadLoop()
		}))

		defer ts.Close()

		url := strings.ReplaceAll(ts.URL, "http", "ws")
		con, err := Dial(url, WithClientBufioParseMode(), WithClientOnMessageFunc(func(c *Conn, mt Opcode, payload []byte) {
		}))
		if err != nil {
			t.Error(err)
		}
		defer con.Close()

		select {
		case d := <-tsort.err:
			if d == nil {
				t.Errorf("got:nil, need:error\n")
			}
		case <-time.After(1000 * time.Millisecond):
			t.Errorf(" Test_ServerOption:WithServerReadTimeout timeout\n")
		}
		if atomic.LoadInt32(&tsort.run) != 1 {
			t.Error("not run server:method fail")
		}
	})
}
