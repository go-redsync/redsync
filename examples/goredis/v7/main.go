package main

import (
	"context"

	goredislib "github.com/go-redis/redis/v7"
	"github.com/stvp/tempredis"
	"github.com/weylan/redsync"
	"github.com/weylan/redsync/redis/goredis/v7"
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

	pool := goredis.NewPool(client)

	rs := redsync.New(pool)

	mutex := rs.NewMutex("test-redsync")
	ctx := context.Background()

	if err := mutex.LockContext(ctx); err != nil {
		panic(err)
	}

	if _, err := mutex.UnlockContext(ctx); err != nil {
		panic(err)
	}
}
