package tinyws

import "errors"

var (
	ErrWrongStatusCode      = errors.New("Wrong status code")
	ErrUpgradeFieldValue    = errors.New("The value of the upgrade field is not 'websocket'")
	ErrConnectionFieldValue = errors.New("The value of the connection field is not 'upgrade'")
	ErrSecWebSocketAccept   = errors.New("The value of Sec-WebSocketAaccept field is invalid")

	ErrHostCannotBeEmpty   = errors.New("Host cannot be empty")
	ErrSecWebSocketKey     = errors.New("The value of SEC websocket key field is wrong")
	ErrSecWebSocketVersion = errors.New("The value of SEC websocket version field is wrong, not 13")

	ErrHTTPProtocolNotSupported = errors.New("HTTP protocol not supported")

	ErrOnlyGETSupported = errors.New("Only get methods are supported")
)
