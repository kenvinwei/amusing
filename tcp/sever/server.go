package main

import (
	"net"
	"log"
	"strings"
	"fmt"
	"time"
)

var allConn = make([]net.Conn, 5)
var Msg chan string = make(chan string, 5)

func handle(conn net.Conn) {

	fmt.Println("handle start......")
	allConn = append(allConn, conn)

	go ReadMsg(conn)

	go Notify(allConn)

}

func Notify(conns []net.Conn) {
	for {
		userMsg := <-Msg

		//通知所有链接
		for _, v := range allConn {
			if _, ok := v.(net.Conn); ok {
				v.Write([]byte(userMsg + "\n"))
			}
		}
	}
}

func ReadMsg(conn net.Conn) {
	for {
		data := make([]byte, 1024)
		n, err := conn.Read(data)
		if n == 0 {
			time.Sleep(2 * time.Millisecond)
			continue
		}

		fmt.Println("read byte len:", n)

		if err != nil {
			log.Println("conn read failed!")

		}

		userMsg := strings.TrimSpace(string(data[0:n]))
		go func() {
			Msg <- userMsg
		}()
		//log.Println("recv msg : ", userMsg)
	}
}

func main() {

	listener, e := net.Listen("tcp", ":8888")

	if e != nil {
		log.Fatalln("server listen error! err:", e)
	}

	for {

		conn, i := listener.Accept()

		if i != nil {
			log.Fatalln("listen accept error!")
		}

		go handle(conn)

	}



}
