package tinyws

import "bufio"

type Conn struct {
	r  *bufio.Reader
	w  *bufio.Writer
	c  net.Conn
	rw sync.Mutex
}

func (c *Conn) nextFrame() (h headFrame, err error) {
	h, err = readHeader(c.r)
	if err != nil {
		return
	}

}

func (c *Conn) ReadMessage() (all []byte, op Opcode, err error) {
}

func (c *Conn) ReadTimeout(t time.Time) (all []byte, op Opcode, err error) {
}

func (c *Conn) WriteMessage(op Opcode, data []byte) (err error) {
}

func (c *Conn) WriteTimeout(t time.Time) (err error) {
}
