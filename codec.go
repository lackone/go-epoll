package go_epoll

type Encoder interface {
	Encode(data []byte) ([]byte, error)
}

type Decoder interface {
	Decode(reader *Buffer) ([]byte, error)
}

type EnDecoder interface {
	Encoder
	Decoder
}
