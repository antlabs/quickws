package main

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
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

// 1.测试不接管上下文，只解压
func echoNoContextDecompression(w http.ResponseWriter, r *http.Request) {
	c, err := quickws.Upgrade(w, r,
		quickws.WithServerReplyPing(),
		quickws.WithServerDecompression(),
		quickws.WithServerIgnorePong(),
		quickws.WithServerCallback(&echoHandler{}),
		quickws.WithServerEnableUTF8Check(),
		// quickws.WithServerReadTimeout(5*time.Second),
	)
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	_ = c.ReadLoop()
}

// 2.测试不接管上下文，压缩和解压
func echoNoContextDecompressionAndCompression(w http.ResponseWriter, r *http.Request) {
	c, err := quickws.Upgrade(w, r,
		quickws.WithServerReplyPing(),
		quickws.WithServerDecompressAndCompress(),
		quickws.WithServerIgnorePong(),
		quickws.WithServerCallback(&echoHandler{}),
		quickws.WithServerEnableUTF8Check(),
	)
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	_ = c.ReadLoop()
}

// 3.测试接管上下文，解压
func echoContextTakeoverDecompression(w http.ResponseWriter, r *http.Request) {
	c, err := quickws.Upgrade(w, r,
		quickws.WithServerReplyPing(),
		quickws.WithServerDecompression(),
		quickws.WithServerIgnorePong(),
		quickws.WithServerContextTakeover(),
		quickws.WithServerCallback(&echoHandler{}),
		quickws.WithServerEnableUTF8Check(),
	)
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	_ = c.ReadLoop()
}

// 4.测试接管上下文，压缩/解压缩
func echoContextTakeoverDecompressionAndCompression(w http.ResponseWriter, r *http.Request) {
	c, err := quickws.Upgrade(w, r,
		quickws.WithServerReplyPing(),
		quickws.WithServerDecompressAndCompress(),
		quickws.WithServerIgnorePong(),
		quickws.WithServerContextTakeover(),
		quickws.WithServerCallback(&echoHandler{}),
		quickws.WithServerEnableUTF8Check(),
	)
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	_ = c.ReadLoop()
}
func echoReadTime(w http.ResponseWriter, r *http.Request) {
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

	_ = c.ReadLoop()
}

var upgrade = quickws.NewUpgrade(
	quickws.WithServerReplyPing(),
	quickws.WithServerDecompression(),
	quickws.WithServerIgnorePong(),
	quickws.WithServerEnableUTF8Check(),
	quickws.WithServerReadTimeout(5*time.Second),
)

func global(w http.ResponseWriter, r *http.Request) {
	c, err := upgrade.UpgradeV2(w, r, &echoHandler{openWriteTimeout: true})
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	_ = c.ReadLoop()
}

func startTLSServer(mux *http.ServeMux) {

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

func startServer(mux *http.ServeMux) {

	rawTCP, err := net.Listen("tcp", ":9001")
	if err != nil {
		fmt.Println("Listen fail:", err)
		return
	}

	log.Println("non-tls server exit:", http.Serve(rawTCP, mux))
}

func main() {
	mux := &http.ServeMux{}
	mux.HandleFunc("/timeout", echoReadTime)
	mux.HandleFunc("/global", global)
	mux.HandleFunc("/no-context-takeover-decompression", echoNoContextDecompression)
	mux.HandleFunc("/no-context-takeover-decompression-and-compression", echoNoContextDecompressionAndCompression)
	mux.HandleFunc("/context-takeover-decompression", echoContextTakeoverDecompression)
	mux.HandleFunc("/context-takeover-decompression-and-compression", echoContextTakeoverDecompressionAndCompression)

	var wg sync.WaitGroup
	wg.Add(2)

	defer wg.Wait()

	go func() {
		defer wg.Done()
		startServer(mux)
	}()

	go func() {
		startTLSServer(mux)
	}()

}
