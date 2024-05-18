package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/antlabs/quickws"
)

// https://github.com/snapview/tokio-tungstenite/blob/master/examples/autobahn-client.rs

const (
	// host = "ws://192.168.128.44:9003"
	host  = "ws://127.0.0.1:9003"
	agent = "quickws"
)

type echoHandler struct {
	done chan struct{}
}

func (e *echoHandler) OnOpen(c *quickws.Conn) {
	fmt.Printf("OnOpen::%p\n", c)
}

func (e *echoHandler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	fmt.Printf("OnMessage: opcode:%s, msg.size:%d\n", op, len(msg))
	if op == quickws.Text || op == quickws.Binary {
		// os.WriteFile("./debug.dat", msg, 0o644)
		// if err := c.WriteMessage(op, msg); err != nil {
		// 	fmt.Println("write fail:", err)
		// }
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

func runTest(caseNo int) {
	done := make(chan struct{})
	c, err := quickws.Dial(fmt.Sprintf("%s/runCase?case=%d&agent=%s", host, caseNo, agent),
		quickws.WithClientReplyPing(),
		quickws.WithClientEnableUTF8Check(),
		quickws.WithClientDecompressAndCompress(),
		quickws.WithClientContextTakeover(),
		quickws.WithClientMaxWindowsBits(10),
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

// 1.先通过接口获取case的总个数
// 2.运行测试客户端client
func main() {
	total := getCaseCount()
	fmt.Println("total case:", total)
	for i := 1; i <= total; i++ {
		runTest(i)
	}
	updateReports()
}
