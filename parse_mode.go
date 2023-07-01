package quickws

type parseMode int32

const (
	ParseModeBufio parseMode = iota
	ParseModeWindows
)
