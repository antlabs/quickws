package quickws

import "io"

// 限制读取
type limitReader struct {
	r io.Reader // 包装的io.Reader
	m int       //最大值
}

// 读取
func (l *limitReader) Read(p []byte) (n int, err error) {
	if l.m <= 0 {
		return 0, io.EOF
	}
	if len(p) > l.m {
		p = p[:l.m]
	}
	n, err = l.r.Read(p)
	l.m -= n
	return
}
