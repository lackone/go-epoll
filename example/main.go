package main

import "fmt"

func main() {
	b := make([]byte, 3)

	c := b[:0]
	
	fmt.Println(c, len(c), cap(c))
}
