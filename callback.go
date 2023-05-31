package tinyws

type Callback interface {
	OnOpen(*Conn)
	OnMessage(*Conn, Opcode, []byte)
	OnClose(*Conn)
}
