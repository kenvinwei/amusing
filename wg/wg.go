package main

import (
	"fmt"
	"sync"
	"time"
)

var shutdownChannel = make(chan struct{}, 0)
var wg = &sync.WaitGroup{}

func start() {
	wg.Add(1)
	go func() {
		fmt.Printf("%s\n", "runing")
		ticker := time.Tick(100*time.Millisecond)

		for shutdown := false; !shutdown; {
			select {
			case <-ticker:
				fmt.Println("tick")
			case <-shutdownChannel:
				fmt.Println("tock")
				shutdown = true
			}
		}
		wg.Done()
	}()
}

func stop() {
	close(shutdownChannel)
}

func wait() {
	wg.Wait()
}

func main() {
	start()
	time.Sleep(time.Second)
	stop()
	wait()
}
