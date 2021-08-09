// +build tinywstest

package tinyws

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	"time"
)

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := Upgrade(w, r)
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	defer c.Close()
	for {
		all, op, err := c.ReadTimeout(3 * time.Second)
		if err != nil {
			fmt.Println("err = ", err)
			return
		}
		//fmt.Printf("%#v\n", c)

		os.Stdout.Write(all)
		c.WriteTimeout(op, all, 3*time.Second)
	}
}

func Test_Server(t *testing.T) {
	http.HandleFunc("/", echo)

	assert.NoError(t, http.ListenAndServe(":9001", nil))
}
