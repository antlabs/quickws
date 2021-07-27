## 简介
tinyws是一个极简的websocket库, 总代码量控制在1k行以下.

## 特性
* 简单, 简单, 简单, 总代码量不足1k行, 能有什么复杂性.

## 例子
### 客户端连服务端
```go
func main() {
    c, err := Dial("ws://127.0.0.1:8080/test")
	if err != nil {
		fmt.Printf("err = %v\n", err)
		return
	}
    defer c.Close()
	
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
```
### 服务端接受客户端请求
```go
func main() {
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
```