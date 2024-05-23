// Copyright 2021-2024 antlabs. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package quickws

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/antlabs/wsutil/bufio2"
	"github.com/antlabs/wsutil/bytespool"
	"github.com/antlabs/wsutil/deflate"
	"github.com/antlabs/wsutil/enum"
	"github.com/antlabs/wsutil/fixedreader"
	"github.com/antlabs/wsutil/hostname"
)

var (
	defaultTimeout = time.Minute * 30
)

type DialOption struct {
	Header               http.Header
	u                    *url.URL
	tlsConfig            *tls.Config
	dialTimeout          time.Duration
	bindClientHttpHeader *http.Header // 握手成功之后, 客户端获取http.Header,
	Config
}

func ClientOptionToConf(opts ...ClientOption) *DialOption {
	var dial DialOption
	if err := dial.defaultSetting(); err != nil {
		panic(err.Error())
	}
	for _, o := range opts {
		o(&dial)
	}
	return &dial
}

func DialConf(rawUrl string, conf *DialOption) (*Conn, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	conf.u = u
	conf.dialTimeout = defaultTimeout
	if conf.Header == nil {
		conf.Header = make(http.Header)
	}
	return conf.Dial()
}

// https://datatracker.ietf.org/doc/html/rfc6455#section-4.1
// 又是一顿if else, 咬文嚼字
func Dial(rawUrl string, opts ...ClientOption) (*Conn, error) {
	var dial DialOption
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	dial.u = u
	dial.dialTimeout = defaultTimeout
	if dial.Header == nil {
		dial.Header = make(http.Header)
	}

	if err := dial.defaultSetting(); err != nil {
		return nil, err
	}

	for _, o := range opts {
		o(&dial)
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
		return nil, "", fmt.Errorf("Unknown scheme, only supports ws:// or wss://: got %s", d.u.Scheme)
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
	if d.Decompression && d.Compression {
		// d.Header.Add("Sec-WebSocket-Extensions", genSecWebSocketExtensions(d.Pd))
		d.Header.Add("Sec-WebSocket-Extensions", deflate.GenSecWebSocketExtensions(d.PermessageDeflateConf))
	}

	if len(d.subProtocols) > 0 {
		d.Header["Sec-WebSocket-Protocol"] = []string{strings.Join(d.subProtocols, ", ")}
	}

	req.Header = d.Header
	return req, secWebSocket, nil
}

// 检查服务端响应的数据
// 4.2.2.5
func (d *DialOption) validateRsp(rsp *http.Response, secWebSocket string) error {
	if rsp.StatusCode != 101 {
		return fmt.Errorf("%w %d", ErrWrongStatusCode, rsp.StatusCode)
	}

	// 第2点
	if !strings.EqualFold(rsp.Header.Get("Upgrade"), "websocket") {
		return ErrUpgradeFieldValue
	}

	// 第3点
	if !strings.EqualFold(rsp.Header.Get("Connection"), "Upgrade") {
		return ErrConnectionFieldValue
	}

	// 第4点
	if !strings.EqualFold(rsp.Header.Get("Sec-WebSocket-Accept"), secWebSocketAcceptVal(secWebSocket)) {
		return ErrSecWebSocketAccept
	}

	// TODO 5点

	// TODO 6点
	return nil
}

// wss已经修改为https
func (d *DialOption) tlsConn(c net.Conn) net.Conn {
	if d.u.Scheme == "https" {
		cfg := d.tlsConfig
		if cfg == nil {
			cfg = &tls.Config{}
		} else {
			cfg = cfg.Clone()
		}

		if cfg.ServerName == "" {
			host := d.u.Host
			if pos := strings.Index(host, ":"); pos != -1 {
				host = host[:pos]
			}
			cfg.ServerName = host
		}
		return tls.Client(c, cfg)
	}

	return c
}

func (d *DialOption) Dial() (wsCon *Conn, err error) {
	// scheme ws -> http
	// scheme wss -> https
	req, secWebSocket, err := d.handshake()
	if err != nil {
		return nil, err
	}

	var conn net.Conn
	begin := time.Now()

	hostName := hostname.GetHostName(d.u)
	// conn, err := net.DialTimeout("tcp", d.u.Host /* TODO 加端号*/, d.dialTimeout)
	dialFunc := net.Dial
	if d.dialFunc != nil {
		dialInterface, err := d.dialFunc()
		if err != nil {
			return nil, err
		}
		dialFunc = dialInterface.Dial
	}

	if d.proxyFunc != nil {
		proxyURL, err := d.proxyFunc(req)
		if err != nil {
			return nil, err
		}
		dialFunc = newhttpProxy(proxyURL, dialFunc).Dial
	}

	conn, err = dialFunc("tcp", hostName)
	if err != nil {
		return nil, err
	}

	dialDuration := time.Since(begin)

	conn = d.tlsConn(conn)
	defer func() {
		if err != nil && conn != nil {
			conn.Close()
			conn = nil
		}
	}()

	if to := d.dialTimeout - dialDuration; to > 0 {
		if err = conn.SetDeadline(time.Now().Add(to)); err != nil {
			return
		}
	}

	defer func() {
		if err == nil {
			err = conn.SetDeadline(time.Time{})
		}
	}()

	if err = req.Write(conn); err != nil {
		return
	}

	br := bufio.NewReader(bufio.NewReader(conn))
	rsp, err := http.ReadResponse(br, req)
	if err != nil {
		return nil, err
	}

	if d.bindClientHttpHeader != nil {
		*d.bindClientHttpHeader = rsp.Header.Clone()
	}

	pd, err := deflate.GetConnPermessageDeflate(rsp.Header)
	if err != nil {
		return nil, err
	}
	if d.Decompression {
		pd.Decompression = pd.Enable && d.Decompression
	}
	if d.Compression {
		pd.Compression = pd.Enable && d.Compression
	}

	if err = d.validateRsp(rsp, secWebSocket); err != nil {
		return
	}

	// 处理下已经在bufio里面的数据，后面都是直接操作net.Conn，所以需要取出bufio里面已读取的数据
	var fr fixedreader.FixedReader
	if d.parseMode == ParseModeWindows {
		fr.Init(conn, bytespool.GetBytes(1024+enum.MaxFrameHeaderSize))
		if br.Buffered() > 0 {
			b, err := br.Peek(br.Buffered())
			if err != nil {
				return nil, err
			}

			buf := fr.BufPtr()
			if len(b) > 1024+enum.MaxFrameHeaderSize {
				bytespool.PutBytes(buf)
				buf = bytespool.GetBytes(len(b) + enum.MaxFrameHeaderSize)

				fr.Reset(buf)
			}

			copy(*buf, b)
			fr.W = len(b)
		}
		bufio2.ClearReader(br)
		br = nil
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return nil, err
	}
	wsCon = newConn(conn, true /* client is true*/, &d.Config, fr, br)
	wsCon.pd = pd
	return wsCon, nil
}
