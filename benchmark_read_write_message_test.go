// Copyright 2021-2024 antlabs. All rights reserved.
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

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/antlabs/wsutil/enum"
	"github.com/antlabs/wsutil/frame"
	"github.com/antlabs/wsutil/opcode"
)

var noMaskData = []byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (testconn *testConn) Read(b []byte) (n int, err error) {
	return testconn.buf.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (testconn *testConn) Write(b []byte) (n int, err error) {
	return testconn.buf.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (testconn *testConn) Close() error {
	return nil
}

// LocalAddr returns the local network address, if known.
func (testconn *testConn) LocalAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

// RemoteAddr returns the remote network address, if known.
func (testconn *testConn) RemoteAddr() net.Addr {
	panic("not implemented") // TODO: Implement
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (testconn *testConn) SetDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (testconn *testConn) SetReadDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (testconn *testConn) SetWriteDeadline(t time.Time) error {
	panic("not implemented") // TODO: Implement
}

type testConn struct {
	buf *bytes.Buffer
}

func Benchmark_WriteMessage(b *testing.B) {
	b.Run("1.case", func(b *testing.B) {
		var c Conn
		buf2 := bytes.NewBuffer(make([]byte, 0, 1024))
		c.c = &testConn{buf: buf2}
		buf := make([]byte, 1024)
		for i := range buf {
			buf[i] = 1
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = c.WriteMessage(opcode.Binary, buf)
			buf2.Reset()
		}
	})
}

func Benchmark_ReadMessage(b *testing.B) {
	b.Run("bufio-TODO", func(b *testing.B) {
	})

	b.Run("windows", func(b *testing.B) {
		var c Conn
		buf2 := bytes.NewBuffer(make([]byte, 0, 1024+enum.MaxFrameHeaderSize))

		c.c = &testConn{buf: buf2}

		windows := make([]byte, 0, 1024)

		c.fr.Init(c.c, &windows)
		c.Callback = &DefCallback{}

		wbuf := make([]byte, 1024)
		for i := range wbuf {
			wbuf[i] = 1
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = c.WriteMessage(opcode.Binary, wbuf)
			_ = c.ReadLoop()
			buf2.Reset()
		}
	})
}

func Benchmark_ReadFrame(b *testing.B) {
	r := bytes.NewReader(noMaskData)
	var headArray [enum.MaxFrameHeaderSize]byte
	for i := 0; i < b.N; i++ {

		r.Reset(noMaskData)
		_, _, err := frame.ReadHeader(r, &headArray)
		if err != nil {
			b.Fatal(err)
		}
	}
}
