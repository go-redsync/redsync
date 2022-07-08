package redsync

import (
	"os"
	"testing"
	"time"

	goredislib "github.com/go-redis/redis"
	goredislib_v7 "github.com/go-redis/redis/v7"
	goredislib_v8 "github.com/go-redis/redis/v8"
	goredislib_v9 "github.com/go-redis/redis/v9"
	"github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	goredis_v7 "github.com/go-redsync/redsync/v4/redis/goredis/v7"
	goredis_v8 "github.com/go-redsync/redsync/v4/redis/goredis/v8"
	goredis_v9 "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	redigolib "github.com/gomodule/redigo/redis"
	"github.com/stvp/tempredis"
)

var servers []*tempredis.Server

type testCase struct {
	poolCount int
	pools     []redis.Pool
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
	}
}

// Maintain separate blocks of servers for each type of driver
const ServerPools = 5
const ServerPoolSize = 8
const RedigoBlock = 0
const GoredisBlock = 1
const GoredisV7Block = 2
const GoredisV8Block = 3
const GoredisV9Block = 4

func TestMain(m *testing.M) {
	for i := 0; i < ServerPoolSize*ServerPools; i++ {
		server, err := tempredis.Start(tempredis.Config{})
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
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			rs := New(v.pools...)

			mutex := rs.NewMutex("test-redsync")
			_ = mutex.Lock()

			assertAcquired(t, v.pools, mutex)
		})
	}
}

func newMockPoolsRedigo(n int) []redis.Pool {
	pools := make([]redis.Pool, n)

	offset := RedigoBlock * ServerPoolSize

	for i := 0; i < n; i++ {
		server := servers[i+offset]
		pools[i] = redigo.NewPool(&redigolib.Pool{
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
	}
	return pools
}

func newMockPoolsGoredis(n int) []redis.Pool {
	pools := make([]redis.Pool, n)

	offset := GoredisBlock * ServerPoolSize

	for i := 0; i < n; i++ {
		client := goredislib.NewClient(&goredislib.Options{
			Network: "unix",
			Addr:    servers[i+offset].Socket(),
		})
		pools[i] = goredis.NewPool(client)
	}
	return pools
}

func newMockPoolsGoredisV7(n int) []redis.Pool {
	pools := make([]redis.Pool, n)

	offset := GoredisV7Block * ServerPoolSize

	for i := 0; i < n; i++ {
		client := goredislib_v7.NewClient(&goredislib_v7.Options{
			Network: "unix",
			Addr:    servers[i+offset].Socket(),
		})
		pools[i] = goredis_v7.NewPool(client)
	}
	return pools
}

func newMockPoolsGoredisV8(n int) []redis.Pool {
	pools := make([]redis.Pool, n)

	offset := GoredisV8Block * ServerPoolSize

	for i := 0; i < n; i++ {
		client := goredislib_v8.NewClient(&goredislib_v8.Options{
			Network: "unix",
			Addr:    servers[i+offset].Socket(),
		})
		pools[i] = goredis_v8.NewPool(client)
	}
	return pools
}

func newMockPoolsGoredisV9(n int) []redis.Pool {
	pools := make([]redis.Pool, n)

	offset := GoredisV9Block * ServerPoolSize

	for i := 0; i < n; i++ {
		client := goredislib_v9.NewClient(&goredislib_v9.Options{
			Network: "unix",
			Addr:    servers[i+offset].Socket(),
		})
		pools[i] = goredis_v9.NewPool(client)
	}
	return pools
}
