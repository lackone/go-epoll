package main

import (
	"bufio"
	"fmt"
	go_epoll "github.com/lackone/go-epoll"
	"io"
	"log"
)

type Handler struct {
}

func (h *Handler) OnConnect(conn *go_epoll.Conn) {
	fmt.Println("连接成功", conn.GetAddr())
}

func (h *Handler) OnData(conn *go_epoll.Conn) {
	//conn连接实现了Read()和Write()方法
	//没有数据了，一定要退出for循环，如果有数据来了，事件反应堆会回调一次OnData方法，如果多个OnData一直都没退出，会出问题。
	reader := bufio.NewReader(conn)
	for {
		readString, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
		}
		fmt.Println("get data : ", readString)

		conn.Write([]byte(readString))
	}
}

func (h *Handler) OnError(conn *go_epoll.Conn) {
	fmt.Println("出错", conn.GetAddr())
}

func (h *Handler) OnClose(conn *go_epoll.Conn) {
	fmt.Println("关闭连接", conn.GetAddr())
}

func main() {
	server, err := go_epoll.NewTcpServer("127.0.0.1:8080", go_epoll.EpollType, 10, 256, 20)
	if err != nil {
		log.Fatalln(err)
	}
	defer server.Close()

	server.SetHandler(&Handler{})

	server.Run()
}
