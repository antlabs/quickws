//go:build tinywstest
// +build tinywstest

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/antlabs/quickws"
	//"os"
)

type echoHandler struct{}

func (e *echoHandler) OnOpen(c *quickws.Conn) {
	fmt.Println("OnOpen:", c)
}

func (e *echoHandler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	fmt.Println("OnMessage:", c, msg, op)
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
	http.HandleFunc("/", echo)

	http.ListenAndServe(":9001", nil)
}
