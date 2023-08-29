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
## 注意⚠️
quickws默认返回read buffer的浅引用，如果生命周期超过OnMessage的，需要clone一份再使用

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

	c.StartReadLoop()
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

    c.WriteMessage(quickws.Text, []byte("hello")])
	c.ReadLoop()

}

```

## 配置函数
### 客户端配置参数
#### 配置header
```go
func main() {
	quickws.Dial("ws://127.0.0.1:12345/test", quickws.WithClientHTTPHeader(http.Header{
		"h1": "v1",
		"h2":"v2", 
	}))
}
```
#### 配置握手时的超时时间
```go
func main() {
	quickws.Dial("ws://127.0.0.1:12345/test", quickws.WithClientDialTimeout(2 * time.Second))
}
```

#### 配置自动回复ping消息
```go
func main() {
	quickws.Dial("ws://127.0.0.1:12345/test", quickws.WithClientReplyPing())
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

## 常见问题
### 1.为什么quickws不标榜zero upgrade?
第一：quickws 是基于 std 的方案实现的 websocket 协议。

第二：原因是 zero upgrade 对 websocket 的性能提升几乎没有影响(同步方式)，所以 quickws 就没有选择花时间优化 upgrade 过程， 

直接基于 net/http， websocket 的协议是整体符合大数定律，一个存活几秒的websocket协议由 upgrade(握手) frame(数据包) frame frame 。。。组成。

所以随着时间的增长, upgrade 对整体的影响接近于0，我们用数字代入下。

A: 代表 upgrade 可能会慢点，但是 frame 的过程比较快，比如基于 net/http 方案的 websocket

upgrade (100ms) frame(10ms) frame(10ms) frame(10ms) avg = 32.5ms

B: 代表主打zero upgrade的库，假如frame的过程处理慢点，

upgrade (90ms) frame(15ms) frame(15ms)  frame(15ms) avg = 33.75ms

简单代入下已经证明了，决定 websocket 差距的是 frame 的处理过程，无论是tps还是内存占用 quickws 在实战中也会证明这个点。所以没有必须也不需要在 upgrade 下功夫，常规优化就够了。


### 2.quickws tps如何
在5800h的cpu上面，tps稳定在47w/s，接近48w/s。比gorilla使用ReadMessage的38.9w/s，快了近9w/s
```
quickws.1:
1s:357999/s 2s:418860/s 3s:440650/s 4s:453360/s 5s:461108/s 6s:465898/s 7s:469211/s 8s:470780/s 9s:472923/s 10s:473821/s 11s:474525/s 12s:475463/s 13s:476021/s 14s:476410/s 15s:477593/s 16s:477943/s 17s:478038/s
gorilla-linux-ReadMessage.4.1 
1s:271126/s 2s:329367/s 3s:353468/s 4s:364842/s 5s:371908/s 6s:377633/s 7s:380870/s 8s:383271/s 9s:384646/s 10s:385986/s 11s:386448/s 12s:386554/s 13s:387573/s 14s:388263/s 15s:388701/s 16s:388867/s 17s:389383/s
gorilla-linux-UseReader.4.2:
1s:293888/s 2s:377628/s 3s:399744/s 4s:413150/s 5s:421092/s 6s:426666/s 7s:430239/s 8s:432801/s 9s:434977/s 10s:436058/s 11s:436805/s 12s:437865/s 13s:438421/s 14s:438901/s 15s:439133/s 16s:439409/s 17s:439578/s 
gobwas.6:
1s:215995/s 2s:279405/s 3s:302249/s 4s:312545/s 5s:318922/s 6s:323800/s 7s:326908/s 8s:329977/s 9s:330959/s 10s:331510/s 11s:331911/s 12s:332396/s 13s:332418/s 14s:332887/s 15s:333198/s 16s:333390/s 17s:333550/s
```
### 3.quickws 流量测试数据如何 ?
在5800h的cpu上面, 同尺寸read buffer(4k), 对比默认用法，quickws在30s处理119GB数据，gorilla处理48GB数据。
* quickws
```
quickws.windows.tcp.delay.4x:
Destination: [127.0.0.1]:9000
Interface lo address [127.0.0.1]:0
Using interface lo to connect to [127.0.0.1]:9000
Ramped up to 10000 connections.
Total data sent:     119153.9 MiB (124941915494 bytes)
Total data received: 119594.6 MiB (125404036361 bytes)
Bandwidth per channel: 6.625⇅ Mbps (828.2 kBps)
Aggregate bandwidth: 33439.980↓, 33316.752↑ Mbps
Packet rate estimate: 3174704.8↓, 2930514.7↑ (9↓, 34↑ TCP MSS/op)
Test duration: 30.001 s.
```
* gorilla 使用ReadMessage取数据
```
gorilla-linux-ReadMessage.tcp.delay:
WARNING: Dumb terminal, expect unglorified output.
Destination: [127.0.0.1]:9003
Interface lo address [127.0.0.1]:0
Using interface lo to connect to [127.0.0.1]:9003
Ramped up to 10000 connections.
Total data sent:     48678.1 MiB (51042707521 bytes)
Total data received: 50406.2 MiB (52854715802 bytes)
Bandwidth per channel: 2.771⇅ Mbps (346.3 kBps)
Aggregate bandwidth: 14094.587↓, 13611.385↑ Mbps
Packet rate estimate: 1399915.6↓, 1190593.2↑ (6↓, 45↑ TCP MSS/op)
Test duration: 30 s.
```

### 4.内存占用如何 ？
quickws的特色之一是低内存占用。

1w连接的tps测试，1k payload 回写，初始内存占用约122MB， 在240s-260s之后大约86MB，
