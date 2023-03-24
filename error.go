package go_epoll

import "errors"

var (
	DemultiplexerTypeUnknown = errors.New("demultiplexer type unknown")
	DemultiplexerSizeError   = errors.New("demultiplexer size ge 1")
	HandlerNotFound          = errors.New("handler not found")
)
