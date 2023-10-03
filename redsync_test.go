package redsync

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	goredislib "github.com/go-redis/redis"
	goredislib_v7 "github.com/go-redis/redis/v7"
	goredislib_v8 "github.com/go-redis/redis/v8"
	redigolib "github.com/gomodule/redigo/redis"
	"github.com/n-h-n/redsync/v4/redis"
	"github.com/n-h-n/redsync/v4/redis/goredis"
	goredis_v7 "github.com/n-h-n/redsync/v4/redis/goredis/v7"
	goredis_v8 "github.com/n-h-n/redsync/v4/redis/goredis/v8"
	goredis_v9 "github.com/n-h-n/redsync/v4/redis/goredis/v9"
	"github.com/n-h-n/redsync/v4/redis/redigo"
	rueidis "github.com/n-h-n/redsync/v4/redis/rueidis"
	goredislib_v9 "github.com/redis/go-redis/v9"
	rueidislib "github.com/rueian/rueidis"
	"github.com/rueian/rueidis/rueidiscompat"
	"github.com/stvp/tempredis"
)

var servers []*tempredis.Server

type testCase struct {
	poolCount int
	pools     []*redis.Pool
}

func makeCases(poolCount int) map[string]*testCase {
	return map[string]*testCase{
		"redigo": {
			poolCount,
			newMockPoolsRedigo(poolCount),
		},
		"goredis": {
			poolCount,
			newMockPoolsGoredis(poolCount),
		},
		"goredis_v7": {
			poolCount,
			newMockPoolsGoredisV7(poolCount),
		},
		"goredis_v8": {
			poolCount,
			newMockPoolsGoredisV8(poolCount),
		},
		"goredis_v9": {
			poolCount,
			newMockPoolsGoredisV9(poolCount),
		},
		"rueidis": {
			poolCount,
			newMockPoolsRueidis(poolCount),
		},
	}
}

// Maintain separate blocks of servers for each type of driver
const (
	ServerPools    = 6
	ServerPoolSize = 8
	RedigoBlock    = 0
	GoredisBlock   = 1
	GoredisV7Block = 2
	GoredisV8Block = 3
	GoredisV9Block = 4
	RueidisBlock   = 5
)

func TestMain(m *testing.M) {
	for i := 0; i < ServerPoolSize*ServerPools; i++ {
		server, err := tempredis.Start(tempredis.Config{
			"port": strconv.Itoa(51200 + i),
		})
		if err != nil {
			panic(err)
		}
		servers = append(servers, server)
	}
	result := m.Run()
	for _, server := range servers {
		_ = server.Term()
	}
	os.Exit(result)
}

func TestRedsync(t *testing.T) {
	ctx := context.Background()
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			rs := New(v.pools...)

			mutex := rs.NewMutex("test-redsync")
			_ = mutex.Lock()

			assertAcquired(ctx, t, v.pools, mutex)
		})
	}
}

func newMockPoolsRedigo(n int) []*redis.Pool {
	pools := make([]*redis.Pool, n)

	offset := RedigoBlock * ServerPoolSize

	for i := 0; i < n; i++ {
		server := servers[i+offset]
		p := redigo.NewPool(&redigolib.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redigolib.Conn, error) {
				return redigolib.Dial("unix", server.Socket())
			},
			TestOnBorrow: func(c redigolib.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		})
		pools[i] = &p
	}
	return pools
}

func newMockPoolsGoredis(n int) []*redis.Pool {
	pools := make([]*redis.Pool, n)

	offset := GoredisBlock * ServerPoolSize

	for i := 0; i < n; i++ {
		client := goredislib.NewClient(&goredislib.Options{
			Network: "unix",
			Addr:    servers[i+offset].Socket(),
		})
		p := goredis.NewPool(client)
		pools[i] = &p
	}
	return pools
}

func newMockPoolsGoredisV7(n int) []*redis.Pool {
	pools := make([]*redis.Pool, n)

	offset := GoredisV7Block * ServerPoolSize

	for i := 0; i < n; i++ {
		client := goredislib_v7.NewClient(&goredislib_v7.Options{
			Network: "unix",
			Addr:    servers[i+offset].Socket(),
		})
		p := goredis_v7.NewPool(client)
		pools[i] = &p
	}
	return pools
}

func newMockPoolsGoredisV8(n int) []*redis.Pool {
	pools := make([]*redis.Pool, n)

	offset := GoredisV8Block * ServerPoolSize

	for i := 0; i < n; i++ {
		client := goredislib_v8.NewClient(&goredislib_v8.Options{
			Network: "unix",
			Addr:    servers[i+offset].Socket(),
		})
		p := goredis_v8.NewPool(client)
		pools[i] = &p
	}
	return pools
}

func newMockPoolsGoredisV9(n int) []*redis.Pool {
	pools := make([]*redis.Pool, n)

	offset := GoredisV9Block * ServerPoolSize

	for i := 0; i < n; i++ {
		client := goredislib_v9.NewClient(&goredislib_v9.Options{
			Network: "unix",
			Addr:    servers[i+offset].Socket(),
		})
		p := goredis_v9.NewPool(client)
		pools[i] = &p
	}
	return pools
}

func newMockPoolsRueidis(n int) []*redis.Pool {
	pools := make([]*redis.Pool, n)

	offset := RueidisBlock * ServerPoolSize

	for i := 0; i < n; i++ {
		client, err := rueidislib.NewClient(rueidislib.ClientOption{
			InitAddress: []string{"127.0.0.1:" + strconv.Itoa(51200+i+offset)},
		})
		if err != nil {
			panic(err)
		}
		p := rueidis.NewPool(rueidiscompat.NewAdapter(client))
		pools[i] = &p
	}
	return pools
}
