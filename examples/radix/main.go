package main

import (
	"context"
	"github.com/go-redsync/redsync/v4"
	radixlib "github.com/go-redsync/redsync/v4/redis/radix"
	"github.com/mediocregopher/radix/v4"
)

func main() {
	var clusterAddr = []string{""}
	pool, err := radix.ClusterConfig{
		PoolConfig: radix.PoolConfig{
			Size: 4,
			//Dialer: radix.Dialer{
			//	AuthPass: "",
			//},
		},
	}.New(context.Background(), clusterAddr)
	if err != nil {
		panic(err)
	}

	rs := redsync.New(radixlib.NewPool(pool))

	mutex := rs.NewMutex("test-redsync")

	if err = mutex.Lock(); err != nil {
		panic(err)
	}

	if _, err = mutex.Unlock(); err != nil {
		panic(err)
	}
}
