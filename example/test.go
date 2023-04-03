package main

import (
	"fmt"
	go_epoll "github.com/lackone/go-epoll"
)

func main() {
	buf := make([]byte, 6)
	buffer := go_epoll.NewBuffer(buf)
	fmt.Println(buffer)

	buffer.Write([]byte("123"))
	fmt.Println(buffer)
	fmt.Println(buffer.Cap())

	buffer.Write([]byte("456"))
	fmt.Println(buffer)
	fmt.Println(buffer.Cap())

	buffer.Write([]byte("789"))
	fmt.Println(buffer)
	fmt.Println(buffer.Cap())

	buffer.Write([]byte("abc"))
	fmt.Println(buffer)
	fmt.Println(buffer.Cap())

	buffer.Write([]byte("def"))
	fmt.Println(buffer)
	fmt.Println(buffer.Cap())

	fmt.Println(buffer.TryGet(32))

	tt := make([]byte, 2)
	n, err := buffer.Read(tt)
	fmt.Println(n, err)
	fmt.Println(tt[:n])
	fmt.Println(buffer)

	fmt.Println(buffer.ReadAt(0, 4))
	fmt.Println(buffer)

	fmt.Println(buffer.ReadAll())
	fmt.Println(buffer)

	n, err = buffer.Read(tt)
	fmt.Println(n, err)
	fmt.Println(buffer)
}
