package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"learngo/go-websocket/impl"
)

var (
	upgrader = websocket.Upgrader{
		//允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)
func wsHandler(w http.ResponseWriter, r *http.Request)  {
	var(
		wsConn *websocket.Conn
		err error
		data []byte
		conn *impl.Connection
	)
	//upgrade:websocket
	if wsConn,err = upgrader.Upgrade(w,r,nil);err != nil{
		return
	}

	if conn,err = impl.InitConnection(wsConn); err != nil{
		goto ERR
	}

	for{

		if data, err = conn.ReadMessage(); err != nil{
			//fmt.Println(data)
			goto ERR
		}


		if err = conn.WriteMessage(data); err != nil{
			goto ERR
		}
	}

ERR:
//关闭连接的操作
	conn.Close()

}

func main() {
	http.HandleFunc("/ws", wsHandler)

	http.ListenAndServe("0.0.0.0:7777", nil)
}