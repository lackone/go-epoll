package go_epoll

import "errors"

var (
	DemultiplexerTypeUnknown = errors.New("demultiplexer type unknown")
	DemultiplexerSizeError   = errors.New("demultiplexer size ge 1")
	EventHandlerNotFound     = errors.New("handler not found")
	DataNotEnough            = errors.New("data Not enough")
)
