package tinyws

import (
	"fmt"
	"testing"
)

func Test_Client(t *testing.T) {
	c, err := Dial("ws://127.0.0.1:8080/test")
	if err != nil {
		fmt.Printf("err = %v\n", err)
		return
	}

	for i := 0; i < 100; i++ {
		err = c.WriteMessage(Text, []byte(fmt.Sprintf("test%d", i)))
		if err != nil {
			fmt.Printf("err = %v\n", err)
			return
		}

		all, _, err := c.ReadMessage()
		if err != nil {
			fmt.Printf("err = %v\n", err)
			return
		}
		fmt.Printf("write :%s\n", string(all))
	}
}
