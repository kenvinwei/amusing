package main

import (
	"net"
	"fmt"
)


var conn net.Conn
var chMsg chan string = make(chan string, 5)

func init()  {
	var e error
	conn, e = net.Dial("tcp", ":8888")
	if e != nil {
		fmt.Println("Dial tcp :8888 error", e)
	}
}

func SendMsg() {

	var data string
	if msg, ok := <- chMsg; ok {
		data = msg
	}
	len := len(data)
	fmt.Println("send before message:", data)
	n, err := conn.Write([]byte(data))
	fmt.Println("send after message:", data)

	if err!=nil {
		fmt.Println("send msg err:", err)
	}

	if n != len {
		fmt.Println("send msg len err:", err)
	}
}

func ReadMsg() {
	var data = make([]byte, 1024)
	n, _ := conn.Read(data)
	fmt.Printf("recv:%s\n", string(data[0:n]))
	Entry()
}


func Entry () {
	var msg string
	fmt.Println("input your msg:")
	fmt.Scanf("%s", &msg)
	chMsg <- msg
}

func main() {

	Entry()
	for{
		SendMsg()
		ReadMsg()
	}

	defer conn.Close()

}
