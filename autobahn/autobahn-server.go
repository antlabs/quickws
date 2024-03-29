//go:build quickwstest
// +build quickwstest

package main

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/antlabs/quickws"
	//"os"
)

//go:embed public.crt
var certPEMBlock []byte

//go:embed privatekey.pem
var keyPEMBlock []byte

type echoHandler struct {
	openWriteTimeout bool
}

func (e *echoHandler) OnOpen(c *quickws.Conn) {
	fmt.Printf("OnOpen: %p\n", c)
}

func (e *echoHandler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	// fmt.Println("OnMessage:", c, msg, op)
	// if err := c.WriteTimeout(op, msg, 3*time.Second); err != nil {
	// 	fmt.Println("write fail:", err)
	// }

	if e.openWriteTimeout {
		if err := c.WriteTimeout(op, msg, 50*time.Second); err != nil {
			fmt.Println("write fail:", err)
		}
		return
	}
	if err := c.WriteMessage(op, msg); err != nil {
		fmt.Println("write fail:", err)
	}
}

func (e *echoHandler) OnClose(c *quickws.Conn, err error) {
	fmt.Printf("OnClose:%p, %v\n", c, err)
}

// echo测试服务
func echo(w http.ResponseWriter, r *http.Request) {
	c, err := quickws.Upgrade(w, r,
		quickws.WithServerReplyPing(),
		quickws.WithServerDecompression(),
		quickws.WithServerIgnorePong(),
		quickws.WithServerCallback(&echoHandler{}),
		quickws.WithServerEnableUTF8Check(),
		quickws.WithServerReadTimeout(5*time.Second),
	)
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	c.ReadLoop()
}

func echo2(w http.ResponseWriter, r *http.Request) {
	c, err := quickws.Upgrade(w, r,
		quickws.WithServerReplyPing(),
		quickws.WithServerDecompression(),
		quickws.WithServerIgnorePong(),
		quickws.WithServerCallback(&echoHandler{openWriteTimeout: true}),
		quickws.WithServerEnableUTF8Check(),
		quickws.WithServerReadTimeout(5*time.Second),
	)
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	c.ReadLoop()
}

func main() {
	mux := &http.ServeMux{}
	mux.HandleFunc("/autobahn", echo)
	mux.HandleFunc("/autobahn-timeout", echo)

	rawTCP, err := net.Listen("tcp", ":9001")
	if err != nil {
		fmt.Println("Listen fail:", err)
		return
	}

	go func() {
		log.Println("non-tls server exit:", http.Serve(rawTCP, mux))
	}()

	cert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		log.Fatalf("tls.X509KeyPair failed: %v", err)
	}
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	lnTLS, err := tls.Listen("tcp", "localhost:9002", tlsConfig)
	if err != nil {
		panic(err)
	}
	log.Println("tls server exit:", http.Serve(lnTLS, mux))
}
