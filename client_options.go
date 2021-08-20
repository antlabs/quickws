package tinyws

import (
	"crypto/tls"
	"net/http"
	"time"
)

type Option interface {
	apply(*DialOption)
}

type tlsConfig tls.Config

func (t *tlsConfig) apply(o *DialOption) {
	o.tlsConfig = (*tls.Config)(t)
}

// 配置tls.config
func WithTLSConfig(tls *tls.Config) Option {
	return (*tlsConfig)(tls)
}

type httpHeader http.Header

func (h httpHeader) apply(o *DialOption) {
	o.Header = (http.Header)(h)
}

// 配置http.Header
func WithHTTPHeader(h http.Header) Option {
	return (httpHeader)(h)
}

type timeDuration time.Duration

func (t timeDuration) apply(o *DialOption) {
	o.dialTimeout = (time.Duration)(t)
}

// 配置握手时的timeout
func WithDialTimeout(t time.Duration) Option {
	return (timeDuration)(t)
}

type replyPing bool

func (r replyPing) apply(o *DialOption) {
	o.replyPing = bool(r)
}

// 配置自动回应ping frame, 当收到ping， 回一个pong
func WithReplyPing() Option {
	var b bool
	return replyPing(b)
}
