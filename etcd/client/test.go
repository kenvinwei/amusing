package main

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"log"
	"fmt"
	"context"
)

func main() {
	var (
		err error
		EtcdKey = "/dev/testKey"
	)
	cli , err := clientv3.New(clientv3.Config{
		Endpoints:[]string{"http://localhost:2380"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Fatalln("connect err:" , err)
	}

	defer cli.Close()


	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err = cli.Put(ctx, EtcdKey, string("192.168.20.2"))
	cancel()

	if err != nil {
		fmt.Println("put failed, err:", err)
		return
	}

	fmt.Println("Set key succ!")
}
