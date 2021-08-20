// +build tinywstest

package tinyws

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	//"os"
	"testing"
	"time"
)

// echo测试服务
func echo(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	c, err := Upgrade(w, r, WithServerReplyPing())
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

func Test_Server(t *testing.T) {
	http.HandleFunc("/", echo)

	assert.NoError(t, http.ListenAndServe(":9001", nil))
}
