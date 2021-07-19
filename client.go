package tinyws

import (
	"net/http"
	"net/url"
)

type DialOption struct {
	Header http.Header
	u      *url.URL
	c      *http.Client
}

// https://datatracker.ietf.org/doc/html/rfc6455#section-4.1
// 又是一顿if else, 咬文嚼字
func Dail(rawUrl string) (*Conn, error) {

	var dial DialOption
	if u, err := url.Parse(rawUrl); err != nil {
		return nil, err
	}

	dial.u = u
	if dial.c == nil {
		dial.c = &http.DefaultClient
	}
	if dial.Header == nil {
		dial.Header = make(http.Header)
	}
	return dial.Dial()
}

func (d *DialOption) Dial() (*Conn, error) {
	switch {
	case d.u.Scheme == "wss":
		d.u.Scheme = "https"
	case d.u.Scheme == "ws":
		d.u.Scheme = "http"
	default:
		//TODO 返回错误
	}

	// 满足4.1
	// GET约束
	// http 1.1版本约束
	req := http.NewRequest("GET", d.u.String(), nil)
	resp, err := d.c.Do(req)
	if err != nil {
		if resp != nil {
			resp.Close()
		}
		return nil, err
	}

	// 第5点
	d.Header.Add("Upgrade", "websocket")
	// 第6点
	d.Header.Add("Connection", "Upgrade")
	// 第7点
	d.Header.Add("Sec-WebSocket-Key", secWebSocketAccept())
	// TODO 第8点
	// 第9点
	d.Header.Add("Sec-WebSocket-Version", "13")
	req.Header = d.Header
	//设置需要的值
	//检查响应值的合法性
}
