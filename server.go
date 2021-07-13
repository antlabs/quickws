package tinyws

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrNotFoundHijacker = errors.New("not found Hijacker")
var strHeaderUpgrade = "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n"

type Upgrader struct {
}

func Upgrade(r *http.Request, w http.ResponseWriter) (c *Conn, err error) {
	if ecode, err := checkRequest(r); err != nil {
		http.Error(w, err.Error(), ecode)
		return nil, err
	}

	hi, ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrNotFoundHijacker
	}

	conn, rw, err := hi.Hijack()
	if err != nil {
		return nil, err
	}

	writeResponse(r, rw.Writer)
	return newConn(conn, rw), nil
}

func writeHeaderKey(w *bufio.Writer, key string) {
	w.WriteString(key)
	w.WriteString(": ")
}

func writeHeaderVal(w *bufio.Writer, val string) {
	w.WriteString(val)
	w.WriteString("\r\n")
}

// https://datatracker.ietf.org/doc/html/rfc6455#section-4.2.2
// 第5小点
func writeResponse(r *http.Request, w *bufio.Writer) {
	w.WriteString(strHeaderUpgrade)
	//写入Sec-WebSocket-Accept key
	writeHeaderKey(w, "Sec-WebSocket-Accept")
	//写入Sec-WebSocket-Accept vla
	writeHeaderVal(w, r.Header.Get("Sec-WebSocket-Key"))
	// TODO 5小点, 处理子协议
}

// https://datatracker.ietf.org/doc/html/rfc6455#section-4.2.1
// 按rfc标准, 先来一顿if else判断, 检查发的request是否满足标准
func checkRequest(r *http.Request) (ecode int, err error) {
	// 不是get方法的
	if r.Method != http.MethodGet {
		//TODO错误消息
		return http.StatusMethodNotAllowed, fmt.Errorf(":got:%s", r.Method)
	}
	// http版本低于1.1
	if r.ProtoAtLeast(1, 1) {
		//TODO错误消息
		return http.StatusHTTPVersionNotSupported, fmt.Errorf("HTTP Version Not Supported")
	}

	// 没有host字段的
	if r.Header.Get("Host") == "" {
		return http.StatusBadRequest, fmt.Errorf("host 不能为空")
	}

	// Upgrade值不等于websocket的
	if upgrade := r.Header.Get("Upgrade"); !strings.EqualFold(upgrade, "websocket") {
		return http.StatusBadRequest, fmt.Errorf("host 不能为空")
	}

	// Connection值不是Upgrader
	if conn := r.Header.Get("Connection"); !strings.EqualFold(conn, "Upgrader") {
		return http.StatusBadRequest, fmt.Errorf("Connection的值不是Upgrader")
	}

	// Sec-WebSocket-Key长度不是16字节的
	if len(r.Header.Get("Sec-WebSocket-Key")) != 16 {
		return http.StatusBadRequest, fmt.Errorf("Sec-WebSocket-Key长度不是16")
	}

	// Sec-WebSocket-Version的版本不是13的
	if r.Header.Get("Sec-WebSocket-Version") != "13" {
		return http.StatusUpgradeRequired, fmt.Errorf("Sec-WebSocket-Version内容不是13")
	}

	// TODO Sec-WebSocket-Protocol
	// TODO Sec-WebSocket-Extensions
	return 0, nil
}
