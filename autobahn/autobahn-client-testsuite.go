package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/antlabs/quickws"
)

// https://github.com/snapview/tokio-tungstenite/blob/master/examples/autobahn-client.rs

type echoHandler struct{}

func (e *echoHandler) OnOpen(c *quickws.Conn) {
	fmt.Println("OnOpen:", c)
}

func (e *echoHandler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	// fmt.Println("OnMessage:", c, msg, op)
	if err := c.WriteTimeout(op, msg, 3*time.Second); err != nil {
		fmt.Println("write fail:", err)
	}
}

func (e *echoHandler) OnClose(c *quickws.Conn, err error) {
	fmt.Println("OnClose:", c, err)
}

// echo测试服务
func echo(w http.ResponseWriter, r *http.Request) {
	c, err := quickws.Upgrade(w, r, quickws.WithServerReplyPing(),
		quickws.WithServerDecompression(),
		quickws.WithServerIgnorePong(),
		quickws.WithServerCallback(&echoHandler{}),
		quickws.WithServerReadTimeout(5*time.Second),
	)
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	c.ReadLoop()
}

func main() {
	c, err := quickws.Dial("ws://127.0.0.1:9003", quickws.WithClientCallback(&echoHandler{}))
	if err != nil {
		fmt.Println("Dial fail:", err)
		return
	}

	c.ReadLoop()
}
