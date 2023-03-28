package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	go_epoll "github.com/lackone/go-epoll"
	"io"
	"log"
)

type EnDecode struct {
}

func (e *EnDecode) Encode(data []byte) ([]byte, error) {
	length := uint32(4 + len(data))
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.BigEndian, length)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *EnDecode) Decode(reader io.Reader) ([]byte, error) {
	var length uint32
	err := binary.Read(reader, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, length-4)
	err = binary.Read(reader, binary.BigEndian, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

type Handler struct {
}

func (h *Handler) OnConnect(conn *go_epoll.Conn) {
	fmt.Println("连接成功", conn.GetAddr())
}

func (h *Handler) OnData(conn *go_epoll.Conn, data []byte) {
	fmt.Println("get client data : ", string(data))
	conn.Write(data)
}

func (h *Handler) OnError(conn *go_epoll.Conn) {
	fmt.Println("出错", conn.GetAddr())
}

func (h *Handler) OnClose(conn *go_epoll.Conn) {
	fmt.Println("关闭连接", conn.GetAddr())
}

func main() {
	go_epoll.SetLimit()

	server, err := go_epoll.NewTcpServer("127.0.0.1:8080", go_epoll.EpollType, 10, 256, 20)
	if err != nil {
		log.Fatalln(err)
	}
	defer server.Close()

	server.SetHandler(&Handler{})

	server.SetEnDecoder(&EnDecode{})

	server.Run()
}
