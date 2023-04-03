package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	go_epoll "github.com/lackone/go-epoll"
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

// 注意，只有解码这里有点特殊，因为必须要解码一个完整的包，因为binary.Read会移动buffer中的start下标，如果解码未成功，需要重置start
func (e *EnDecode) Decode(reader *go_epoll.Buffer) (data []byte, err error) {
	start := reader.GetStart()

	defer func() {
		if err != nil {
			//如果出错，则重置start
			reader.SetStart(start)
		}
	}()

	var length uint32
	//尝试读取4个字节数据
	_, err = reader.TryGet(4)
	if err != nil {
		return nil, err
	}

	err = binary.Read(reader, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}

	data = make([]byte, length-4)

	//尝试读取data长度的字节数据
	_, err = reader.TryGet(len(data))
	if err != nil {
		return nil, err
	}
	err = binary.Read(reader, binary.BigEndian, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type Handler struct {
}

func (h *Handler) OnConnect(conn *go_epoll.Conn) {
	fmt.Println("连接成功", conn.GetAddr())
}

func (h *Handler) OnData(conn *go_epoll.Conn, data []byte) {
	fmt.Printf("get client[%s] data : %s\n", conn.GetAddr(), string(data))
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
