package impl

import (
	"github.com/gorilla/websocket"
	"sync"
	"errors"
	"fmt"
)

type Connection struct {
	inChan chan[]byte
	outChan chan[]byte
	closeChan chan []byte
	mutex sync.Mutex
	isClosed bool
	wsConn *websocket.Conn
}
var client_map  =  make(map[*websocket.Conn]bool)


func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error){
	conn = &Connection{
		inChan:make(chan []byte, 1000),
		outChan:make(chan []byte, 10000),
		closeChan:make(chan []byte, 1),
		wsConn: wsConn,
	}
	client_map[wsConn] = true
	fmt.Println(client_map)
	//启动读和写协程
	go conn.readLoop()
	go conn.writeLoop()

	return
}

func (conn *Connection)ReadMessage()(data[]byte, err error){
	select {
	case data = <- conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}
	return
}

func(conn *Connection)WriteMessage(data[]byte) (err error)  {
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}
	return
}

func(conn *Connection)Close(){
	fmt.Println("close")
	for k,_ := range client_map{
		k.Close()
		//这一行只执行一次
		conn.mutex.Lock()
		if !conn.isClosed{
			close(conn.closeChan)
			conn.isClosed = true
		}
		conn.mutex.Unlock()
	}

}
//内部实现

func(conn *Connection) readLoop()  {
	var(
		data []byte
		err error

	)
	for {
		//fmt.Println("in readloop")

		if _,data,err = conn.wsConn.ReadMessage(); err != nil{
			goto ERR
		}
		//fmt.Println("read content")
		//阻塞在这里，等待inchan有空闲
		select {
		case conn.inChan <- data:
		case <-conn.closeChan:
			goto ERR
		}

	}

	ERR:
		conn.Close()

}
func(conn *Connection) writeLoop()  {
	var(
		data []byte
		err error
	)
	for {
		select {
		case data = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR
		}
		for k,_ := range client_map {

			if err = k.WriteMessage(websocket.TextMessage, data); err != nil {
				goto ERR
			}
		}

	}
	ERR:
		conn.Close()
}
