package redsync

import (
	"os"
	"testing"
	"time"

	goredislib "github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v3/redis"
	"github.com/go-redsync/redsync/v3/redis/goredis"
	"github.com/go-redsync/redsync/v3/redis/redigo"
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
	}
}

// Maintain separate blocks of servers for each type of driver
const SERVER_POOLS = 2
const SERVER_POOL_SIZE = 8
const REDIGO_BLOCK = 0
const GOREDIS_BLOCK = 1

func TestMain(m *testing.M) {
	// Create 16 here, goredis will use the last 8
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
			rs := New(v.pools)

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
		pools[i] = redigo.NewRedigoPool(&redigolib.Pool{
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
		pools[i] = goredis.NewGoredisPool(client)
	}
	return pools
}
