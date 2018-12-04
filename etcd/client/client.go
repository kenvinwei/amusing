package main

import (
	"log"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
	"context"
	"go.etcd.io/etcd/mvcc/mvccpb"
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
	_, err = cli.Put(ctx, EtcdKey, string("192.168.20.1"))
	cancel()
	if err != nil {
		fmt.Println("put failed, err:", err)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, EtcdKey)
	cancel()

	if err != nil {
		fmt.Println("get failed, err:", err)
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("key is %s : value is %s\n", ev.Key, ev.Value)
	}

	for {
		rch := cli.Watch(context.Background(), EtcdKey)

		for wresp := range rch{
			for _, ev := range wresp.Events{
				if ev.Type == mvccpb.DELETE {
					fmt.Printf("Key (%s) is delete\n", EtcdKey)
				}

				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == EtcdKey {
					fmt.Printf("Key (%s) is update\n", EtcdKey)
				}
			}
		}
	}






}
