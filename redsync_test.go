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
		/*
			"goredis": {
				poolCount,
				newMockPoolsGoredis(poolCount),
			},
		*/
	}
}

func TestMain(m *testing.M) {
	for i := 0; i < 8; i++ {
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
	for _, server := range servers {
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
		}(server)
		if len(pools) == n {
			break
		}
	}
	return pools
}

func newMockPoolsGoredis(n int) []redis.Pool {
	pools := []redis.Pool{}

	for _, server := range servers {
		func(server *tempredis.Server) {

			client := goredislib.NewClient(&goredislib.Options{
				Addr: server.Socket(),
			})

			pools = append(pools, goredis.NewGoredisPool(client))
		}(server)
		if len(pools) == n {
			break
		}
	}
	return pools
}
