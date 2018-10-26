package main

import (
	"net"
	"log"
	"strings"
)

var allConn = make([]net.Conn, 5)

func handle(conn net.Conn) {

	allConn = append(allConn, conn)

	log.Println(allConn)


	data := make([]byte, 1024)
	n, err := conn.Read(data)
	if err != nil {
		log.Println("conn read failed!")
	}

	userMsg := strings.TrimSpace(string(data[0:n]))
	log.Printf("msg : %s\n", userMsg)
	//conn.Write([]byte(userMsg + "\n"))

	//通知所有链接
	for _, v := range allConn {
		if _, ok := v.(net.Conn); ok {
			v.Write([]byte(userMsg + "\n"))
		}
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
