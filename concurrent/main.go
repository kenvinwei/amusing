package main

import (
	"fmt"
	"time"
)
var num = 0
var n = 0

func main() {

	for j := 0; j < 50; j++ {
		go func(h int) {
				num = num + 1
				fmt.Printf("h=%d and num=%d\n", h, num)
		}(j)
	}

	for j := 0; j < 50; j++ {
		n = n + 1
	}
	time.Sleep(time.Second)
	fmt.Println(num)
	fmt.Printf("n=%d", n)


}
