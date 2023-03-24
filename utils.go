package go_epoll

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"net"
	"net/http"
	"net/netip"
	"reflect"
)

// 获取连接的FD
func GetSocketFD(ln interface{}) (int, error) {
	//通过反射，获取类型
	t := reflect.Indirect(reflect.ValueOf(ln)).Type().String()
	switch t {
	case "net.TCPConn":
		//通过反射，获取net.TCPConn结构体中的conn字段
		conn := reflect.Indirect(reflect.ValueOf(ln)).FieldByName("conn")
		return ValueTCPFD(conn), nil
	case "net.TCPListener":
		return PointerTCPFD(ln), nil
	case "tls.Conn":
		conn := reflect.Indirect(reflect.ValueOf(ln)).FieldByName("conn")
		//注意，这里使用Elem()获取指针变量指向的值，因为tls.Conn结构体conn的类型为net.Conn，是一个interface
		//所以，比net.TCPConn多一步
		conn = reflect.Indirect(conn.Elem())
		return ValueTCPFD(conn), nil
	case "tls.listener":
		tln := reflect.Indirect(reflect.ValueOf(ln)).FieldByName("Listener")
		if ln, ok := tln.Interface().(*net.TCPListener); ok {
			return PointerTCPFD(ln), nil
		}
	}
	return 0, errors.New(fmt.Sprintf("get socket fd error type : %s", t))
}

func PointerTCPFD(ln interface{}) int {
	fdVal := reflect.Indirect(reflect.ValueOf(ln)).FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

func ValueTCPFD(ln reflect.Value) int {
	fdVal := ln.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

// 获取内网IP
func GetInternalIP() (string, error) {
	conn, err := net.Dial("udp", "114.114.114.114:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	addrPort, err := netip.ParseAddrPort(conn.LocalAddr().String())
	if err != nil {
		return "", err
	}
	return addrPort.Addr().String(), nil
}

// 获取外网IP
func GetExternalIP() (string, error) {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(all), nil
}

// 获取地址
func GetSockAddrInet4(addr string) (*unix.SockaddrInet4, error) {
	addrPort, err := netip.ParseAddrPort(addr)
	if err != nil {
		return nil, err
	}

	inet4 := &unix.SockaddrInet4{
		Port: int(addrPort.Port()),
	}
	copy(inet4.Addr[:], net.ParseIP(addrPort.Addr().String()).To4())

	return inet4, nil
}

// 获取IP
func GetIPBySockAddr(sa unix.Sockaddr) string {
	inet4 := sa.(*unix.SockaddrInet4)
	addr := net.IPv4(inet4.Addr[0], inet4.Addr[1], inet4.Addr[2], inet4.Addr[3]).To4().String()
	return addr
}
