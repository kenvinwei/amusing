package main

import (
	"net/rpc"
	"github.com/gpmgo/gopm/modules/log"
	"fmt"
)

func main() {
	client, err := rpc.Dial("tcp", ":9001")
	if err != nil {
		log.Fatal("Dial error : ", err)
	}

	var reply string

	err = client.Call("Hello.Hello", "hello", &reply)

	if err != nil{
		log.Fatal("Call func error : ", err)
	}

	fmt.Println(reply)
}
