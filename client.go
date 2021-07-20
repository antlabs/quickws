package tinyws

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type DialOption struct {
	Header http.Header
	u      *url.URL
}

// https://datatracker.ietf.org/doc/html/rfc6455#section-4.1
// 又是一顿if else, 咬文嚼字
func Dail(rawUrl string) (*Conn, error) {

	var dial DialOption
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	dial.u = u
	if dial.Header == nil {
		dial.Header = make(http.Header)
	}
	return dial.Dial()
}

// 准备握手的数据
func (d *DialOption) handshake() (*http.Request, string, error) {
	switch {
	case d.u.Scheme == "wss":
		d.u.Scheme = "https"
	case d.u.Scheme == "ws":
		d.u.Scheme = "http"
	default:
		//TODO 返回错误
	}

	// 满足4.1
	// 第2点 GET约束http 1.1版本约束
	req, err := http.NewRequest("GET", d.u.String(), nil)
	if err != nil {
		return nil, "", err
	}
	// 第5点
	d.Header.Add("Upgrade", "websocket")
	// 第6点
	d.Header.Add("Connection", "Upgrade")
	// 第7点
	secWebSocket := secWebSocketAccept()
	d.Header.Add("Sec-WebSocket-Key", secWebSocket)
	// TODO 第8点
	// 第9点
	d.Header.Add("Sec-WebSocket-Version", "13")
	req.Header = d.Header
	return req, secWebSocket, nil

}

// 检查服务端响应的数据
// 4.2.2.5
func (d *DialOption) validateRsp(rsp *http.Response, secWebSocket string) error {
	if rsp.StatusCode != 101 {
		return errors.New("状态码不对")
	}

	// 第2点
	if !strings.EqualFold(rsp.Header.Get("Upgrade"), "websocket") {
		return errors.New("Upgrade的值不对")
	}

	// 第3点
	if !strings.EqualFold(rsp.Header.Get("Connection"), "Upgrade") {
		return errors.New("Connection的值不对")
	}

	// 第4点
	if strings.EqualFold(rsp.Header.Get("Sec-WebSocket-Accept"), secWebSocketAcceptVal(secWebSocket)) {
		return errors.New("sec websocket accept错误")
	}

	// TODO 5点

	// TODO 6点
	return nil
}

func (d *DialOption) Dial() (*Conn, error) {

	//检查响应值的合法性
	req, secWebSocket, err := d.handshake()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", d.u.Host /* TODO 加端号*/)
	if err != nil {
		return nil, err
	}

	req.Write(conn)
	brw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	rsp, err := http.ReadResponse(brw.Reader, req)
	if err != nil {
		return nil, err
	}

	if err = d.validateRsp(rsp, secWebSocket); err != nil {
		return nil, err
	}
	return newConn(conn, brw), nil
}
