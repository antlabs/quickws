package tinyws

// https://datatracker.ietf.org/doc/html/rfc6455#section-7.4.1
// 这里记录了各种状态码的含义
type StatusCode int32

const (
	// NormalClosure 正常关闭
	NormalClosure StatusCode = 1000
	// EndpointGoingAway 对端正在消失
	EndpointGoingAway StatusCode = 1001
	// ProtocolError 表示对端由于协议错误正在终止连接
	ProtocolError StatusCode = 1002
	// DataCannotAccept 收到一个不能接受的数据类型
	DataCannotAccept StatusCode = 1003
	// NotConsistentMessageType 表示对端正在终止连接, 消息类型不一致
	NotConsistentMessageType StatusCode = 1007
	// TerminatingConnection 表示对端正在终止连接, 没有好用的错误, 可以用这个错误码表示
	TerminatingConnection StatusCode = 1008
	// TooBigMessage  消息太大, 不能处理, 关闭连接
	TooBigMessage StatusCode = 1009
	// NoExtensions 只用于客户端, 服务端返回扩展消息
	NoExtensions StatusCode = 1010
	// ServerTerminating 服务端遇到意外情况, 中止请求
	ServerTerminating StatusCode = 1011
)
