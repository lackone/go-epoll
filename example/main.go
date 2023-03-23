package main

import "fmt"

func main() {

	v := make(map[int]int, 10)

	fmt.Println(cap(v))
}
