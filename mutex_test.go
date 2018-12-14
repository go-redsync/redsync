package redsync

import (
	"strconv"
	"testing"
	"time"

	redsyncredis "github.com/go-redsync/redsync/redis"
)

func TestMutex(t *testing.T) {
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			mutexes := newTestMutexes(v.pool, "test-mutex", v.poolCount)
			orderCh := make(chan int)
			for i, mutex := range mutexes {
				go func(i int, mutex *Mutex) {
					err := mutex.Lock()
					if err != nil {
						t.Fatalf("Expected err == nil, got %q", err)
					}
					defer mutex.Unlock()

					assertAcquired(t, v.pool, mutex)

					orderCh <- i
				}(i, mutex)
			}
			for range mutexes {
				<-orderCh
			}

		})
	}
}

func TestMutexExtend(t *testing.T) {
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			mutexes := newTestMutexes(v.pool, "test-mutex-extend", 1)
			mutex := mutexes[0]

			err := mutex.Lock()
			if err != nil {
				t.Fatalf("Expected err == nil, got %q", err)
			}
			defer mutex.Unlock()

			time.Sleep(1 * time.Second)

			expiries := getPoolExpiries(v.pool, mutex.name)
			ok := mutex.Extend()
			if !ok {
				t.Fatalf("Expected ok == true, got %v", ok)
			}
			expiries2 := getPoolExpiries(v.pool, mutex.name)

			for i, expiry := range expiries {
				if expiry >= expiries2[i] {
					t.Fatalf("Expected expiries[%d] > expiry, got %d %d", i, expiries2[i], expiry)
				}
			}

		})
	}
}

func TestMutexQuorum(t *testing.T) {
	for k, v := range makeCases(4) {
		t.Run(k, func(t *testing.T) {
			for mask := 0; mask < 1<<uint(len(v.pool)); mask++ {
				mutexes := newTestMutexes(v.pool, "test-mutex-partial-"+strconv.Itoa(mask), 1)
				mutex := mutexes[0]
				mutex.tries = 1

				n := clogPools(v.pool, mask, mutex)

				if n >= len(v.pool)/2+1 {
					err := mutex.Lock()
					if err != nil {
						t.Fatalf("Expected err == nil, got %q", err)
					}
					assertAcquired(t, v.pool, mutex)
				} else {
					err := mutex.Lock()
					if err != ErrFailed {
						t.Fatalf("Expected err == %q, got %q", ErrFailed, err)
					}
				}
			}
		})
	}
}

func getPoolValues(pools []redsyncredis.Pool, name string) []string {
	values := []string{}
	for _, pool := range pools {
		conn := pool.Get()
		value, err := conn.Get(name)
		conn.Close()
		if err != nil {
			panic(err)
		}
		values = append(values, value)
	}
	return values
}

func getPoolExpiries(pools []redsyncredis.Pool, name string) []int {
	expiries := []int{}
	for _, pool := range pools {
		conn := pool.Get()
		expiry, err := conn.PTTL(name)
		conn.Close()
		if err != nil {
			panic(err)
		}
		expiries = append(expiries, int(expiry))
	}
	return expiries
}

func clogPools(pools []redsyncredis.Pool, mask int, mutex *Mutex) int {
	n := 0
	for i, pool := range pools {
		if mask&(1<<uint(i)) == 0 {
			n++
			continue
		}
		conn := pool.Get()
		_, err := conn.Set(mutex.name, "foobar")
		conn.Close()
		if err != nil {
			panic(err)
		}
	}
	return n
}

func newTestMutexes(pools []redsyncredis.Pool, name string, n int) []*Mutex {
	mutexes := []*Mutex{}
	for i := 0; i < n; i++ {
		mutexes = append(mutexes, &Mutex{
			name:      name,
			expiry:    8 * time.Second,
			tries:     32,
			delayFunc: func(tries int) time.Duration { return 500 * time.Millisecond },
			factor:    0.01,
			quorum:    len(pools)/2 + 1,
			pools:     pools,
		})
	}
	return mutexes
}

func assertAcquired(t *testing.T, pools []redsyncredis.Pool, mutex *Mutex) {
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
}
