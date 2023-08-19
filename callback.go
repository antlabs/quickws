// Copyright 2021-2023 antlabs. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package quickws

type (
	Callback interface {
		OnOpen(*Conn)
		OnMessage(*Conn, Opcode, []byte)
		OnClose(*Conn, error)
	}
)

type (
	OnOpenFunc func(*Conn)
)

type DefCallback struct{}

func (defcallback *DefCallback) OnOpen(_ *Conn) {
}

func (defcallback *DefCallback) OnMessage(_ *Conn, _ Opcode, _ []byte) {
}

func (defcallback *DefCallback) OnClose(_ *Conn, _ error) {
}

// 只设置OnMessage, 和OnClose互斥
type OnMessageFunc func(*Conn, Opcode, []byte)

func (o OnMessageFunc) OnOpen(_ *Conn) {
}

func (o OnMessageFunc) OnMessage(c *Conn, op Opcode, data []byte) {
	o(c, op, data)
}

func (o OnMessageFunc) OnClose(_ *Conn, _ error) {
}

// 只设置OnClose, 和OnMessage互斥
type OnCloseFunc func(*Conn, error)

func (o OnCloseFunc) OnOpen(_ *Conn) {
}

func (o OnCloseFunc) OnMessage(_ *Conn, _ Opcode, _ []byte) {
}

func (o OnCloseFunc) OnClose(c *Conn, err error) {
	o(c, err)
}

type funcToCallback struct {
	onOpen    func(*Conn)
	onMessage func(*Conn, Opcode, []byte)
	onClose   func(*Conn, error)
}

func (f *funcToCallback) OnOpen(c *Conn) {
	if f.onOpen != nil {
		f.onOpen(c)
	}
}

func (f *funcToCallback) OnMessage(c *Conn, op Opcode, data []byte) {
	if f.onMessage != nil {
		f.onMessage(c, op, data)
	}
}

func (f *funcToCallback) OnClose(c *Conn, err error) {
	if f.onClose != nil {
		f.onClose(c, err)
	}
}
