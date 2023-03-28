package go_epoll

import "io"

type Encoder interface {
	Encode(data []byte) ([]byte, error)
}

type Decoder interface {
	Decode(reader io.Reader) ([]byte, error)
}

type EnDecoder interface {
	Encoder
	Decoder
}
