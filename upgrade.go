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
	"bytes"
	"net"
	"net/http"
	"time"

	"github.com/antlabs/wsutil/bufio2"
	"github.com/antlabs/wsutil/bytespool"
	"github.com/antlabs/wsutil/deflate"
	"github.com/antlabs/wsutil/fixedreader"
)

type UpgradeServer struct {
	config Config
}

func NewUpgrade(opts ...ServerOption) *UpgradeServer {
	var conf ConnOption
	if err := conf.defaultSetting(); err != nil {
		panic(err.Error())
	}
	for _, o := range opts {
		o(&conf)
	}
	return &UpgradeServer{config: conf.Config}
}

func (u *UpgradeServer) Upgrade(w http.ResponseWriter, r *http.Request) (c *Conn, err error) {
	return upgradeInner(w, r, &u.config, nil)
}

func (u *UpgradeServer) UpgradeV2(w http.ResponseWriter, r *http.Request, cb Callback) (c *Conn, err error) {
	return upgradeInner(w, r, &u.config, cb)
}

func Upgrade(w http.ResponseWriter, r *http.Request, opts ...ServerOption) (c *Conn, err error) {
	var conf ConnOption
	if err := conf.defaultSetting(); err != nil {
		return nil, err
	}
	for _, o := range opts {
		o(&conf)
	}
	return upgradeInner(w, r, &conf.Config, nil)
}

func upgradeInner(w http.ResponseWriter, r *http.Request, conf *Config, cb Callback) (wsCon *Conn, err error) {
	if ecode, err := checkRequest(r); err != nil {
		http.Error(w, err.Error(), ecode)
		return nil, err
	}

	hi, ok := w.(http.Hijacker)
	if !ok {
		return nil, ErrNotFoundHijacker
	}

	var br *bufio.Reader
	var conn net.Conn
	var rw *bufio.ReadWriter
	if conf.parseMode == ParseModeWindows {
		// 这里不需要rw，直接使用conn
		conn, rw, err = hi.Hijack()
		if !conf.disableBufioClearHack {
			bufio2.ClearReadWriter(rw)
		}
		// TODO
		// rsp.ClearRsp(w)
		// rw = nil
	} else {
		var rw *bufio.ReadWriter
		conn, rw, err = hi.Hijack()
		br = rw.Reader
		rw = nil
	}
	if err != nil {
		return nil, err
	}

	// 是否打开解压缩
	// 外层接收压缩, 并且客户端发送扩展过来
	var pd deflate.PermessageDeflateConf
	if conf.Decompression {
		pd, err = deflate.GetConnPermessageDeflate(r.Header)
		if err != nil {
			return nil, err
		}
	}

	buf := bytespool.GetUpgradeRespBytes()

	tmpWriter := bytes.NewBuffer((*buf)[:0])
	defer func() {
		bytespool.PutUpgradeRespBytes(buf)
		tmpWriter = nil
	}()

	resetPermessageDeflate(&pd, conf)
	if err = prepareWriteResponse(r, tmpWriter, conf, pd); err != nil {
		return
	}

	if _, err := conn.Write(tmpWriter.Bytes()); err != nil {
		return nil, err
	}

	var fr fixedreader.FixedReader
	if conf.parseMode == ParseModeWindows {
		fr.Init(conn, bytespool.GetBytes(conf.initPayloadSize()))
	}

	if err := conn.SetDeadline(time.Time{}); err != nil {
		return nil, err
	}
	if wsCon, err = newConn(conn, false, conf, fr, br); err != nil {
		return nil, err
	}

	wsCon.pd = pd
	wsCon.Callback = cb
	if cb == nil {
		wsCon.Callback = conf.cb
	}
	return wsCon, nil
}

func resetPermessageDeflate(pd *deflate.PermessageDeflateConf, conf *Config) {
	pd.Decompression = pd.Enable && conf.Decompression
	pd.Compression = pd.Enable && conf.Compression
	pd.ServerContextTakeover = pd.Enable && pd.ServerContextTakeover && conf.ServerContextTakeover
	pd.ClientContextTakeover = pd.Enable && pd.ClientContextTakeover && conf.ClientContextTakeover
}
