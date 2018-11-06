package main

import (
	"context"
	"time"
	"fmt"
)

var resultChan chan string = make(chan string)

func do() {
	time.Sleep(3 * time.Second)
	resultChan <- "do something..."
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()
	go do()

	select {
	case <-ctx.Done():
		fmt.Println("Timeout!")
	case <-resultChan:
		fmt.Println("Working")
	}

}
