package main

import (
	"net/rpc"
	"net"
	"log"
)

type HelloService struct {}

func (p *HelloService) Hello(request string, reply *string) error {
	*reply = "Hello:" + request
	return nil
}


func main() {
	rpc.RegisterName("Hello", new(HelloService))
 	listener, err :=	net.Listen("tcp", ":9001")

	if err!=nil {
		log.Fatal("ListenerTcp error : ", err)
	}

	conn, err := listener.Accept()

	if err != nil {
		log.Fatal("Accept error : ", err)
	}

	rpc.ServeConn(conn)
}
