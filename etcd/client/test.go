package main

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"log"
	"fmt"
	"context"
)
//这里用客户端的etcdctl 并不生效原因是 api版本不一致，请设置 export ETCDCTL_API=3 即可
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
