package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/antlabs/quickws"
)

// https://github.com/snapview/tokio-tungstenite/blob/master/examples/autobahn-client.rs

const (
	host = "ws://192.168.128.44:9003"
	// host  = "ws://127.0.0.1:9003"
	agent = "quickws"
)

type echoHandler struct {
	done chan struct{}
}

func (e *echoHandler) OnOpen(c *quickws.Conn) {
	fmt.Println("OnOpen:", c)
}

func (e *echoHandler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	fmt.Println("OnMessage:", c, op, msg)
	if op == quickws.Text || op == quickws.Binary {
		// os.WriteFile("./debug.dat", msg, 0o644)
		if err := c.WriteTimeout(op, msg, 1*time.Minute); err != nil {
			fmt.Println("write fail:", err)
		}
	}
}

func (e *echoHandler) OnClose(c *quickws.Conn, err error) {
	fmt.Println("OnClose:", c, err)
	close(e.done)
}

func getCaseCount() int {
	var count int
	c, err := quickws.Dial(fmt.Sprintf("%s/getCaseCount", host), quickws.WithClientOnMessageFunc(func() quickws.OnMessageFunc {
		return func(c *quickws.Conn, op quickws.Opcode, msg []byte) {
			var err error
			fmt.Printf("msg(%s)\n", msg)
			count, err = strconv.Atoi(string(msg))
			if err != nil {
				panic(err)
			}
			c.Close()
		}
	}()))
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = c.ReadLoop()
	fmt.Printf("readloop rv:%s\n", err)
	return count
}

// echo测试服务
// func echo(w http.ResponseWriter, r *http.Request) {
// 	c, err := quickws.Upgrade(w, r, quickws.WithServerReplyPing(),
// 		quickws.WithServerDecompression(),
// 		quickws.WithServerIgnorePong(),
// 		quickws.WithServerCallback(&echoHandler{}),
// 		quickws.WithServerReadTimeout(5*time.Second),
// 	)
// 	if err != nil {
// 		fmt.Println("Upgrade fail:", err)
// 		return
// 	}

// 	go c.ReadLoop()
// }

func runTest(caseNo int) {
	done := make(chan struct{})
	c, err := quickws.Dial(fmt.Sprintf("%s/runCase?case=%d&agent=%s", host, caseNo, agent),
		quickws.WithClientReplyPing(),
		// quickws.WithClientCompression(),
		quickws.WithClientDecompressAndCompress(),
		quickws.WithClientCallback(&echoHandler{done: done}),
	)
	if err != nil {
		fmt.Println("Dial fail:", err)
		return
	}

	go c.ReadLoop()
	<-done
}

func updateReports() {
	c, err := quickws.Dial(fmt.Sprintf("%s/updateReports?agent=%s", host, agent))
	if err != nil {
		fmt.Println("Dial fail:", err)
		return
	}

	c.Close()
}

func main() {
	total := getCaseCount()
	fmt.Println("total case:", total)
	for i := 1; i <= total; i++ {
		runTest(i)
	}
	updateReports()
}
