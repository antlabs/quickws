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
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
)

type (
	dialFunc  func(network, addr string) (c net.Conn, err error)
	httpProxy struct {
		proxyAddr *url.URL
		dial      func(network, addr string) (c net.Conn, err error)
	}
)

var _ Dialer = (*httpProxy)(nil)

func newhttpProxy(u *url.URL, dial dialFunc) *httpProxy {
	return &httpProxy{proxyAddr: u, dial: dial}
}

func (h *httpProxy) Dial(network, addr string) (c net.Conn, err error) {
	if h.proxyAddr == nil {
		return h.dial(network, addr)
	}

	hostName := getHostName(h.proxyAddr)
	c, err = h.dial(network, hostName)
	if err != nil {
		return nil, err
	}

	header := make(http.Header)

	if u := h.proxyAddr.User; u != nil {
		user := u.Username()
		if pass, ok := u.Password(); ok {
			credential := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
			header.Set("Proxy-Authorization", "Basic "+credential)
		}
	}

	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: hostName},
		Host:   hostName,
		Header: header,
	}

	if err := req.Write(c); err != nil {
		c.Close()
		return nil, err
	}

	br := bufio.NewReader(c)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		c.Close()
		return nil, err
	}

	if resp.StatusCode != 200 {
		c.Close()
		return nil, getHttpErrMsg(resp.StatusCode)
	}
	return c, nil
}
