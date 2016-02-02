package redsync

import (
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/stvp/tempredis"
)

func TestMutex(t *testing.T) {
	pools := newMockPools()
	mutexes := newTestMutexes(pools, 8)
	orderCh := make(chan int)
	for i, mutex := range mutexes {
		go func(i int, mutex *Mutex) {
			err := mutex.Lock()
			if err != nil {
				t.Fatalf("Expected err == nil, got %q", err)
			}
			defer mutex.Unlock()

			n := 0
			values := getPoolValues(pools, mutex.name)
			for _, value := range values {
				if value == mutex.value {
					n++
				}
			}
			if n < mutex.quorum {
				t.Fatalf("Expected n >= %d, got %d", mutex.quorum, n)
			}

			orderCh <- i
		}(i, mutex)
	}
	for range mutexes {
		<-orderCh
	}
}

func newMockPools() []Pool {
	pools := []Pool{}
	for _, server := range servers {
		func(server *tempredis.Server) {
			pools = append(pools, &redis.Pool{
				MaxIdle:     3,
				IdleTimeout: 240 * time.Second,
				Dial: func() (redis.Conn, error) {
					return redis.Dial("unix", server.Socket())
				},
				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
			})
		}(server)
	}
	return pools
}

func getPoolValues(pools []Pool, name string) []string {
	values := []string{}
	for _, pool := range pools {
		conn := pool.Get()
		value, err := redis.String(conn.Do("GET", name))
		if err != nil && err != redis.ErrNil {
			panic(err)
		}
		values = append(values, value)
	}
	return values
}

func newTestMutexes(pools []Pool, n int) []*Mutex {
	mutexes := []*Mutex{}
	for i := 0; i < n; i++ {
		mutexes = append(mutexes, &Mutex{
			name:   "test-mutex",
			expiry: 8 * time.Second,
			tries:  16,
			delay:  512 * time.Millisecond,
			factor: 0.01,
			quorum: len(pools)/2 + 1,
			pools:  pools,
		})
	}
	return mutexes
}
