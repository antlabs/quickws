package quickws

type Callback interface {
	OnOpen(*Conn)
	OnMessage(*Conn, Opcode, []byte)
	OnClose(*Conn, error)
}
