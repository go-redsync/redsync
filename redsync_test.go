package redsync

import (
	goredislib "github.com/go-redis/redis"
	"github.com/go-redsync/redsync/redis/goredis"
	"github.com/go-redsync/redsync/redis/redigo"
	redigolib "github.com/gomodule/redigo/redis"
	"os"
	"testing"
	"time"

	"github.com/go-redsync/redsync/redis"
	"github.com/stvp/tempredis"
)

var servers []*tempredis.Server

type testCase struct {
	poolCount int
	pool      []redis.Pool
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

// Maintain seprate blocks of servers for each type of driver
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
		server.Term()
	}
	os.Exit(result)
}

func TestRedsync(t *testing.T) {

	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			rs := New(v.pool)

			mutex := rs.NewMutex("test-redsync")
			err := mutex.Lock()
			if err != nil {

			}

			assertAcquired(t, v.pool, mutex)
		})
	}
}

func newMockPoolsRedigo(n int) []redis.Pool {
	pools := []redis.Pool{}

	offset := REDIGO_BLOCK * SERVER_POOL_SIZE

	for i := offset; i < offset+SERVER_POOL_SIZE; i++ {
		func(server *tempredis.Server) {
			pools = append(pools, redigo.NewRedigoPool(&redigolib.Pool{
				MaxIdle:     3,
				IdleTimeout: 240 * time.Second,
				Dial: func() (redigolib.Conn, error) {
					return redigolib.Dial("unix", server.Socket())
				},
				TestOnBorrow: func(c redigolib.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
			}))
		}(servers[i])
		if len(pools) == n {
			break
		}

	}
	return pools
}

func newMockPoolsGoredis(n int) []redis.Pool {
	pools := []redis.Pool{}

	offset := GOREDIS_BLOCK * SERVER_POOL_SIZE

	for i := offset; i < offset+SERVER_POOL_SIZE; i++ {
		func(server *tempredis.Server) {

			client := goredislib.NewClient(&goredislib.Options{
				Network: "unix",
				Addr:    server.Socket(),
			})

			pools = append(pools, goredis.NewGoredisPool(client))
		}(servers[i])
		if len(pools) == n {
			break
		}

	}
	return pools
}
