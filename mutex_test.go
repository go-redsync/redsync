package redsync

import (
	"strconv"
	"testing"
	"time"

	"github.com/go-redsync/redsync/v4/redis"
)

func TestMutex(t *testing.T) {
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			mutexes := newTestMutexes(v.pools, "test-mutex", v.poolCount)
			orderCh := make(chan int)
			for i, mutex := range mutexes {
				go func(i int, mutex *Mutex) {
					err := mutex.Lock()
					if err != nil {
						t.Fatalf("mutex lock failed: %s", err)
					}
					defer mutex.Unlock()

					assertAcquired(t, v.pools, mutex)

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
			mutexes := newTestMutexes(v.pools, "test-mutex-extend", 1)
			mutex := mutexes[0]

			err := mutex.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			defer mutex.Unlock()

			time.Sleep(1 * time.Second)

			expiries := getPoolExpiries(v.pools, mutex.name)
			ok, err := mutex.Extend()
			if err != nil {
				t.Fatalf("mutex extend failed: %s", err)
			}
			if !ok {
				t.Fatalf("Expected a valid mutex")
			}
			expiries2 := getPoolExpiries(v.pools, mutex.name)

			for i, expiry := range expiries {
				if expiry >= expiries2[i] {
					t.Fatalf("Expected expiries[%d] > %d, got %d", i, expiry, expiries2[i])
				}
			}
		})
	}
}

func TestMutexExtendExpired(t *testing.T) {
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			mutexes := newTestMutexes(v.pools, "test-mutex-extend", 1)
			mutex := mutexes[0]
			mutex.expiry = 500 * time.Millisecond

			err := mutex.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			defer mutex.Unlock()

			time.Sleep(1 * time.Second)

			ok, err := mutex.Extend()
			if err != nil {
				t.Fatalf("mutex extend failed: %s", err)
			}
			if ok {
				t.Fatalf("Expected ok == false, got %v", ok)
			}
		})
	}
}

func TestMutexUnlockExpired(t *testing.T) {
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			mutexes := newTestMutexes(v.pools, "test-mutex-extend", 1)
			mutex := mutexes[0]
			mutex.expiry = 500 * time.Millisecond

			err := mutex.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			defer mutex.Unlock()

			time.Sleep(1 * time.Second)

			ok, err := mutex.Unlock()
			if err != nil {
				t.Fatalf("mutex unlock failed: %s", err)
			}
			if ok {
				t.Fatalf("Expected ok == false, got %v", ok)
			}
		})
	}
}

func TestMutexQuorum(t *testing.T) {
	for k, v := range makeCases(4) {
		t.Run(k, func(t *testing.T) {
			for mask := 0; mask < 1<<uint(len(v.pools)); mask++ {
				mutexes := newTestMutexes(v.pools, "test-mutex-partial-"+strconv.Itoa(mask), 1)
				mutex := mutexes[0]
				mutex.tries = 1

				n := clogPools(v.pools, mask, mutex)

				if n >= len(v.pools)/2+1 {
					err := mutex.Lock()
					if err != nil {
						t.Fatalf("mutex lock failed: %s", err)
					}
					assertAcquired(t, v.pools, mutex)
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

func TestValid(t *testing.T) {
	for k, v := range makeCases(4) {
		t.Run(k, func(t *testing.T) {
			rs := New(v.pools...)
			key := "test-shared-lock"

			mutex1 := rs.NewMutex(key, WithExpiry(time.Hour))
			err := mutex1.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			assertAcquired(t, v.pools, mutex1)

			ok, err := mutex1.Valid()
			if err != nil {
				t.Fatalf("mutex valid failed: %s", err)
			}
			if !ok {
				t.Fatalf("Expected a valid mutex")
			}

			mutex2 := rs.NewMutex(key)
			err = mutex2.Lock()
			if err == nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
		})
	}
}

func getPoolValues(pools []redis.Pool, name string) []string {
	values := make([]string, len(pools))
	for i, pool := range pools {
		conn, err := pool.Get(nil)
		if err != nil {
			panic(err)
		}
		value, err := conn.Get(name)
		if err != nil {
			panic(err)
		}
		_ = conn.Close()
		values[i] = value
	}
	return values
}

func getPoolExpiries(pools []redis.Pool, name string) []int {
	expiries := make([]int, len(pools))
	for i, pool := range pools {
		conn, err := pool.Get(nil)
		if err != nil {
			panic(err)
		}
		expiry, err := conn.PTTL(name)
		if err != nil {
			panic(err)
		}
		_ = conn.Close()
		expiries[i] = int(expiry)
	}
	return expiries
}

func clogPools(pools []redis.Pool, mask int, mutex *Mutex) int {
	n := 0
	for i, pool := range pools {
		if mask&(1<<uint(i)) == 0 {
			n++
			continue
		}
		conn, err := pool.Get(nil)
		if err != nil {
			panic(err)
		}
		_, err = conn.Set(mutex.name, "foobar")
		if err != nil {
			panic(err)
		}
		_ = conn.Close()
	}
	return n
}

func newTestMutexes(pools []redis.Pool, name string, n int) []*Mutex {
	mutexes := make([]*Mutex, n)
	for i := 0; i < n; i++ {
		mutexes[i] = &Mutex{
			name:         name,
			expiry:       8 * time.Second,
			tries:        32,
			delayFunc:    func(tries int) time.Duration { return 500 * time.Millisecond },
			genValueFunc: genValue,
			factor:       0.01,
			quorum:       len(pools)/2 + 1,
			pools:        pools,
		}
	}
	return mutexes
}

func assertAcquired(t *testing.T, pools []redis.Pool, mutex *Mutex) {
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
