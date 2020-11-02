package main

import (
	"context"

	goredislib "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/stvp/tempredis"
)

func main() {
	server, err := tempredis.Start(tempredis.Config{})
	if err != nil {
		panic(err)
	}
	defer server.Term()
	clusterSlots := func(ctx context.Context) ([]goredislib.ClusterSlot, error) {
		slots := []goredislib.ClusterSlot{{
				Start: 0,
				End:   16383,
				Nodes: []goredislib.ClusterNode{{Addr: server.Socket()}, {Addr: server.Socket()}}},
		}
		return slots, nil
	}

	redisCluster := goredislib.NewClusterClient(&goredislib.ClusterOptions{
		ClusterSlots:  clusterSlots,
		RouteRandomly: true,
	})

	pool := goredis.NewPool(redisCluster)

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
