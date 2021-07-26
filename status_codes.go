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
)
