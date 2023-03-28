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
