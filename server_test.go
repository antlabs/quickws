package tinyws

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

/*
func Test_Server(t *testing.T) {
	h1 := func(w http.ResponseWriter, r *http.Request) {
		c, err := Upgrade(w, r)
		if err != nil {
			fmt.Println("Upgrade fail:", err)
			return
		}

		for {
			all, _, err := c.ReadTimeout(3 * time.Second)
			if err != nil {
				fmt.Println("err = ", err)
				return
			}

			os.Stdout.Write(all)
		}
	}

	http.HandleFunc("/", h1)

	http.ListenAndServe(":12345", nil)
}
*/
