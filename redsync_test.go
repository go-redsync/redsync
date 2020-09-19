package redsync

import (
	"os"
	"testing"
	"time"

	goredislib "github.com/go-redis/redis"
	goredislib_v7 "github.com/go-redis/redis/v7"
	goredislib_v8 "github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	goredis_v7 "github.com/go-redsync/redsync/v4/redis/goredis/v7"
	goredis_v8 "github.com/go-redsync/redsync/v4/redis/goredis/v8"
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
	}
}

// Maintain separate blocks of servers for each type of driver
const SERVER_POOLS = 4
const SERVER_POOL_SIZE = 8
const REDIGO_BLOCK = 0
const GOREDIS_BLOCK = 1
const GOREDIS_V7_BLOCK = 2
const GOREDIS_V8_BLOCK = 3

func TestMain(m *testing.M) {
	for i := 0; i < SERVER_POOL_SIZE*SERVER_POOLS; i++ {
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

	offset := REDIGO_BLOCK * SERVER_POOL_SIZE

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

	offset := GOREDIS_BLOCK * SERVER_POOL_SIZE

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

	offset := GOREDIS_V7_BLOCK * SERVER_POOL_SIZE

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

	offset := GOREDIS_V8_BLOCK * SERVER_POOL_SIZE

	for i := 0; i < n; i++ {
		client := goredislib_v8.NewClient(&goredislib_v8.Options{
			Network: "unix",
			Addr:    servers[i+offset].Socket(),
		})
		pools[i] = goredis_v8.NewPool(client)
	}
	return pools
}
