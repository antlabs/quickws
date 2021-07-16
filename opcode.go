package tinyws

type Opcode uint8

const (
	Continuation Opcode = iota
	Text
	Binary
	// 3 - 7保留
	_ //3
	_
	_ //5
	_
	_ //7
	Close
	Ping
	Pong
)

func (c Opcode) String() string {
	switch {
	case c >= 3 && c <= 7:
		return "control"
	case c == Text:
		return "text"
	case c == Binary:
		return "binary"
	case c == Close:
		return "close"
	case c == Ping:
		return "ping"
	case c == Pong:
		return "pong"
	case c == Continuation:
		return "continuation"
	default:
		return "unknown"
	}
}

func (c Opcode) isControl() bool {
	return (c & (1 << 3)) > 0
}
