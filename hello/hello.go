package main

import "fmt"

func main() {
	var n int = 0
	for i := 0; i < 1000000; i++ {
		fmt.Println(n)
		n++
	}

}
