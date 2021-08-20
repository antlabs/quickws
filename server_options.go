package tinyws

type ServerOption interface {
	apply(*ConnOption)
}

type serverReplyPing bool

func (r serverReplyPing) apply(o *ConnOption) {
	o.replyPing = bool(r)
}

// 配置自动回应ping frame, 当收到ping， 回一个pong
func WithServerReplyPing() ServerOption {
	b := true

	return serverReplyPing(b)
}
