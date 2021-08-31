// +build tinywstest

package main

import (
	"fmt"
	"net/http"

	"github.com/guonaihong/tinyws"

	//"os"

	"time"
)

// echo测试服务
func echo(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	c, err := tinyws.Upgrade(w, r, tinyws.WithServerReplyPing(), tinyws.WithServerCompression())
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	fmt.Printf("new conn:%p ", c)
	defer func() {
		c.Close()
		fmt.Printf("conn bye bye %p:%v\n", c, time.Since(now))
	}()

	for {
		now := time.Now()
		all, op, err := c.ReadTimeout(3 * time.Second)
		if err != nil {
			fmt.Printf("err = %v ", err)
			return
		}

		fmt.Printf("%p, read:%v ", c, time.Since(now))

		//os.Stdout.Write(all)
		now = time.Now()
		c.WriteTimeout(op, all, 3*time.Second)
		fmt.Printf("%p, write:%v ", c, time.Since(now))
	}
}

func main() {
	http.HandleFunc("/", echo)

	http.ListenAndServe(":9001", nil)
}
