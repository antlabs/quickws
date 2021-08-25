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

	ErrOnlyGETSupported    = errors.New("error:Only get methods are supported")
	ErrMaxControlFrameSize = errors.New("error:max control frame size > 125")
	ErrRsv123              = errors.New("error:rsv1 or rsv2 or rsv3 has a value")
	ErrOpcode              = errors.New("error:wrong opcode")
	ErrNOTBeFragmented     = errors.New("error:since control message MUST NOT be fragmented")
	ErrFrameOpcode         = errors.New("error:since all data frames after the initial data frame must have opcode 0.")
	ErrTextNotUTF8         = errors.New("error:text is not utf8 data")
)
