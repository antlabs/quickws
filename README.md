## 简介
tinyws是一个极简的websocket库, 总代码量控制在3k行以下.

![Go](https://github.com/guonaihong/tinyws/workflows/Go/badge.svg)
[![codecov](https://codecov.io/gh/guonaihong/tinyws/branch/main/graph/badge.svg)](https://codecov.io/gh/guonaihong/tinyws)
[![Go Report Card](https://goreportcard.com/badge/github.com/guonaihong/tinyws)](https://goreportcard.com/report/github.com/guonaihong/tinyws)

## 特性
* 3倍的简单
* 实现rfc6455
* 实现rfc7692

## 内容
* [安装](#Installation)
* [例子](#example)
	* [客户端连服务端](#客户端连服务端)
	* [服务端接受客户端请求](#服务端接受客户端请求)
* [配置函数](#配置函数)
	* [客户端配置参数](#客户端配置)
		* [配置header](#配置header)
		* [配置握手时的超时时间](#配置握手时的超时时间)
		* [配置自动回复ping消息](#配置自动回复ping消息)
	* [服务配置参数](#服务端配置)
		* [配置服务自动回复ping消息](#配置服务自动回复ping消息)
## Installation
```console
go get github.com/guonaihong/tinyws
```

## example
### 客户端连服务端
```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/guonaihong/tinyws"
)

func main() {
	h1 := func(w http.ResponseWriter, r *http.Request) {
		c, err := tinyws.Upgrade(w, r)
		if err != nil {
			fmt.Println("Upgrade fail:", err)
			return
		}
		defer c.Close()

		for {
			all, op, err := c.ReadTimeout(3 * time.Second)
			if err != nil {
				if err != io.EOF {
					fmt.Println("err = ", err)
				}
				return
			}

			os.Stdout.Write(all)
			c.WriteTimeout(op, all, 3*time.Second)
		}
	}

	http.HandleFunc("/", h1)

	http.ListenAndServe(":12345", nil)
}

```
### 服务端接受客户端请求
```go
package main

import (
	"fmt"

	"github.com/guonaihong/tinyws"
)

func main() {
	c, err := tinyws.Dial("ws://127.0.0.1:12345/test")
	if err != nil {
		fmt.Printf("err = %v\n", err)
		return
	}

	defer c.Close()

	err = c.WriteMessage(tinyws.Text, []byte("hello"))
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

```

## 配置函数
### 客户端配置参数
#### 配置header
```go
func main() {
	tinyws.Dial("ws://127.0.0.1:12345/test", tinyws.WithHTTPHeader(http.Header{
		"h1": "v1",
		"h2":"v2", 
	}))
}
```
#### 配置握手时的超时时间
```go
func main() {
	tinyws.Dial("ws://127.0.0.1:12345/test", tinyws.WithDialTimeout(2 * time.Second))
}
```

#### 配置自动回复ping消息
```go
func main() {
	tinyws.Dial("ws://127.0.0.1:12345/test", tinyws.WithReplyPing())
}
```
### 服务端配置参数
#### 配置服务自动回复ping消息
```go
func main() {
	  c, err := tinyws.Upgrade(w, r, tinyws.WithServerReplyPing())
        if err != nil {
                fmt.Println("Upgrade fail:", err)
                return
        }   
}
```