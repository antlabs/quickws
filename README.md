## 简介
quickws是一个极简的websocket库

![Go](https://github.com/antlabs/quickws/workflows/Go/badge.svg)
[![codecov](https://codecov.io/gh/antlabs/quickws/branch/main/graph/badge.svg)](https://codecov.io/gh/antlabs/quickws)
[![Go Report Card](https://goreportcard.com/badge/github.com/antlabs/quickws)](https://goreportcard.com/report/github.com/antlabs/quickws)

## 特性
* 3倍的简单
* 实现rfc6455
* 实现rfc7692

## 内容
* [安装](#Installation)
* [例子](#example)
	* [服务端](#服务端)
	* [客户端](#客户端)
* [配置函数](#配置函数)
	* [客户端配置参数](#客户端配置)
		* [配置header](#配置header)
		* [配置握手时的超时时间](#配置握手时的超时时间)
		* [配置自动回复ping消息](#配置自动回复ping消息)
	* [服务配置参数](#服务端配置)
		* [配置服务自动回复ping消息](#配置服务自动回复ping消息)
## Installation
```console
go get github.com/antlabs/quickws
```

## example
### 服务端
```go

package main

import (
	"fmt"
	"net/http"
	"time"
	"github.com/antlabs/quickws"
)

type echoHandler struct{}

func (e *echoHandler) OnOpen(c *quickws.Conn) {
	fmt.Println("OnOpen:", c)
}

func (e *echoHandler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	fmt.Println("OnMessage:", c, msg, op)
	if err := c.WriteTimeout(op, msg, 3*time.Second); err != nil {
		fmt.Println("write fail:", err)
	}
}

func (e *echoHandler) OnClose(c *quickws.Conn, err error) {
	fmt.Println("OnClose:", c, err)
}

// echo测试服务
func echo(w http.ResponseWriter, r *http.Request) {
	c, err := quickws.Upgrade(w, r, quickws.WithServerReplyPing(),
		// quickws.WithServerDecompression(),
		// quickws.WithServerIgnorePong(),
		quickws.WithServerCallback(&echoHandler{}),
		quickws.WithServerReadTimeout(5*time.Second),
	)
	if err != nil {
		fmt.Println("Upgrade fail:", err)
		return
	}

	c.ReadLoop()
}

func main() {
	http.HandleFunc("/", echo)

	http.ListenAndServe(":9001", nil)
}

```
### 客户端
```go
package main

import (
	"fmt"

	"github.com/antlabs/quickws"
)

type echoHandler struct{}

func (e *echoHandler) OnOpen(c *quickws.Conn) {
	fmt.Println("OnOpen:", c)
}

func (e *echoHandler) OnMessage(c *quickws.Conn, op quickws.Opcode, msg []byte) {
	fmt.Println("OnMessage:", c, msg, op)
	if err := c.WriteTimeout(op, msg, 3*time.Second); err != nil {
		fmt.Println("write fail:", err)
	}
}

func (e *echoHandler) OnClose(c *quickws.Conn, err error) {
	fmt.Println("OnClose:", c, err)
}

func main() {
	c, err := quickws.Dial("ws://127.0.0.1:12345/test")
	if err != nil {
		fmt.Printf("err = %v\n", err)
		return
	}

	c.ReadLoop()

}

```

## 配置函数
### 客户端配置参数
#### 配置header
```go
func main() {
	quickws.Dial("ws://127.0.0.1:12345/test", quickws.WithHTTPHeader(http.Header{
		"h1": "v1",
		"h2":"v2", 
	}))
}
```
#### 配置握手时的超时时间
```go
func main() {
	quickws.Dial("ws://127.0.0.1:12345/test", quickws.WithDialTimeout(2 * time.Second))
}
```

#### 配置自动回复ping消息
```go
func main() {
	quickws.Dial("ws://127.0.0.1:12345/test", quickws.WithReplyPing())
}
```
### 服务端配置参数
#### 配置服务自动回复ping消息
```go
func main() {
	c, err := quickws.Upgrade(w, r, quickws.WithServerReplyPing())
        if err != nil {
                fmt.Println("Upgrade fail:", err)
                return
        }   
}
```
