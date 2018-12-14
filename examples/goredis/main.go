package main

import (
	goredislib "github.com/go-redis/redis"
	"github.com/go-redsync/redsync"
	"github.com/go-redsync/redsync/redis"
	"github.com/go-redsync/redsync/redis/goredis"
	"github.com/stvp/tempredis"
)

func main() {

	server, err := tempredis.Start(tempredis.Config{})
	if err != nil {
		panic(err)
	}
	defer server.Term()

	client := goredislib.NewClient(&goredislib.Options{
		Network: "unix",
		Addr:    server.Socket(),
	})

	pool := goredis.NewGoredisPool(client)

	rs := redsync.New([]redis.Pool{pool})

	mutex := rs.NewMutex("test-redsync")
	err = mutex.Lock()

	if err != nil {
		panic(err)
	}

	mutex.Unlock()
}
