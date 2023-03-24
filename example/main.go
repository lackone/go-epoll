package main

import (
	"fmt"
	"sync/atomic"
)

func main() {

	var a uint32

	atomic.AddUint32(&a, 1)
	atomic.AddUint32(&a, 1)
	fmt.Println(a)

	atomic.AddUint32(&a, -uint32(1))
	fmt.Println(a)
}
