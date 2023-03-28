# go-epoll
go 实现 epoll，基于边缘触发模式

### 示例

服务端代码：
```go
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

//编码
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
//解码
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
//连接成功
func (h *Handler) OnConnect(conn *go_epoll.Conn) {
	fmt.Println("连接成功", conn.GetAddr())
}
//数据到来
func (h *Handler) OnData(conn *go_epoll.Conn, data []byte) {
	fmt.Println("get client data : ", string(data))
	conn.Write(data)
}
//发生错误
func (h *Handler) OnError(conn *go_epoll.Conn) {
	fmt.Println("出错", conn.GetAddr())
}
//连接关闭
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

```

客户端代码：
```go
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

//编码
func Encode(data []byte) ([]byte, error) {
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
//解码
func Decode(reader io.Reader) ([]byte, error) {
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

func main() {
	dial, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		encode, _ := Encode([]byte(time.Now().Format("2006-01-02 15:04:05")))
		dial.Write(encode)

		decode, err := Decode(dial)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("get server data : ", string(decode))

		time.Sleep(time.Second)
	}
}
```