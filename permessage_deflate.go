package quickws

import (
	"net/http"
	"strconv"
)

// https://datatracker.ietf.org/doc/html/rfc7692#section-7.1
type permessageDeflate struct {
	// 服务端是否支持上下文接管
	// https://datatracker.ietf.org/doc/html/rfc7692#section-7.1.1.1
	// 客户端可以发送 server_no_context_takeover 参数，表示服务端不需要上下文接管
	serverNoContextTakeover bool
	// 客户端是否支持上下文接管
	// https://datatracker.ietf.org/doc/html/rfc7692#section-7.1.1.2
	// 客户端发关 client_no_context_takeover 参数，表示客户端不使用上下文接管
	// 即使服务端没有响应 client_no_context_takeover 参数，客户端也不会使用上下文接管
	clientNoContextTakeover bool

	// 客户端最大窗口位数， 8-15
	clientMaxWindowBits int
	// 服务端最大窗口位数，8-15
	serverMaxWindowBits int
}

func parseMaxWindowBits(val string) (int, error) {
	if val == "" {
		return 15, nil
	}
	bits, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	if bits < 8 || bits > 15 {
		return 0, http.ErrNotSupported
	}
	return bits, nil
}

func parsePermessageDeflate(header http.Header) (pmd permessageDeflate, err error) {
	params := parseExtensions(header)
	for _, param := range params {
		switch param.key {
		case "server_no_context_takeover":
			pmd.serverNoContextTakeover = true
		case "client_no_context_takeover":
			pmd.clientNoContextTakeover = true
		case "client_max_window_bits":
			if pmd.clientMaxWindowBits, err = parseMaxWindowBits(param.val); err != nil {
				return
			}
		case "server_max_window_bits":
			if pmd.serverMaxWindowBits, err = parseMaxWindowBits(param.val); err != nil {
				return
			}
		default:
			err = http.ErrNotSupported
			return
		}
	}
	return
}
