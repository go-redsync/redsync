package redsync

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/go-redsync/redsync/v4/redis"
	"golang.org/x/sync/errgroup"
)

func TestMutex(t *testing.T) {
	ctx := context.Background()
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			mutexes := newTestMutexes(v.pools, k+"-test-mutex", v.poolCount)
			eg := errgroup.Group{}
			for _, mutex := range mutexes {
				eg.Go(func() error {
					err := mutex.Lock()
					if err != nil {
						return err
					}
					defer mutex.Unlock()

					if !isAcquired(ctx, v.pools, mutex) {
						return fmt.Errorf("Expected n >= %d, got %d", mutex.quorum, countAcquiredPools(ctx, v.pools, mutex))
					}
					return nil
				})
			}
			err := eg.Wait()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
		})
	}
}

func TestTryLock(t *testing.T) {
	ctx := context.Background()
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			mutexes := newTestMutexes(v.pools, k+"-test-trylock", v.poolCount)
			eg := errgroup.Group{}
			for _, mutex := range mutexes {
				eg.Go(func() error {
					err := mutex.TryLockContext(context.Background())
					for {
						if err == nil {
							break
						}

						time.Sleep(100 * time.Millisecond) // 1ms
					}

					ok, err := mutex.Extend()
					if err != nil || !ok {
						t.Fatalf("extend failed: %v", err)
					}

					defer mutex.Unlock()

					if !isAcquired(ctx, v.pools, mutex) {
						return fmt.Errorf("Expected n >= %d, got %d", mutex.quorum, countAcquiredPools(ctx, v.pools, mutex))
					}
					return nil
				})
			}
			err := eg.Wait()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
		})
	}
}

func TestMutexAlreadyLocked(t *testing.T) {
	ctx := context.Background()
	for k, v := range makeCases(4) {
		t.Run(k, func(t *testing.T) {
			rs := New(v.pools...)
			key := "test-lock"

			mutex1 := rs.NewMutex(key)
			err := mutex1.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			assertAcquired(ctx, t, v.pools, mutex1)

			mutex2 := rs.NewMutex(key)
			err = mutex2.TryLock()
			var errTaken *ErrTaken
			if !errors.As(err, &errTaken) {
				t.Fatalf("mutex was not already locked: %s", err)
			}
		})
	}
}

func TestMutexExtend(t *testing.T) {
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			mutexes := newTestMutexes(v.pools, k+"-test-mutex-extend", 1)
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
			mutexes := newTestMutexes(v.pools, k+"-test-mutex-extend", 1)
			mutex := mutexes[0]
			mutex.expiry = 500 * time.Millisecond

			err := mutex.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			defer mutex.Unlock()

			time.Sleep(1 * time.Second)

			_, err = mutex.Extend()
			if err == nil {
				t.Fatalf("mutex extend didn't fail")
			}
		})
	}
}

func TestSetNXOnExtendAcquiresLockWhenKeyIsExpired(t *testing.T) {
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			mutexes := newTestMutexes(v.pools, k+"-test-mutex-extend", 1)
			mutex := mutexes[0]
			mutex.setNXOnExtend = true
			mutex.expiry = 500 * time.Millisecond

			err := mutex.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			defer mutex.Unlock()

			time.Sleep(1 * time.Second)

			_, err = mutex.Extend()
			if err != nil {
				t.Fatalf("mutex didn't extend")
			}
		})
	}
}

func TestSetNXOnExtendFailsToAcquireLockWhenKeyIsTaken(t *testing.T) {
	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			firstMutex := newTestMutexes(v.pools, k+"-test-mutex-extend", 1)[0]
			firstMutex.expiry = 500 * time.Millisecond

			err := firstMutex.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			defer firstMutex.Unlock()

			time.Sleep(1 * time.Second)

			secondMutex := newTestMutexes(v.pools, k+"-test-mutex-extend", 1)[0]
			firstMutex.expiry = 500 * time.Millisecond
			err = secondMutex.Lock()
			defer secondMutex.Unlock()
			if err != nil {
				t.Fatalf("second mutex couldn't lock")
			}

			ok, err := firstMutex.Extend()
			firstMutex.setNXOnExtend = true
			if err == nil {
				t.Fatalf("mutex extend didn't fail")
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
			mutexes := newTestMutexes(v.pools, k+"-test-mutex-extend", 1)
			mutex := mutexes[0]
			mutex.expiry = 500 * time.Millisecond

			err := mutex.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			defer mutex.Unlock()

			time.Sleep(1 * time.Second)

			ok, err := mutex.Unlock()
			if err == nil {
				t.Fatalf("mutex unlock didn't fail")
			}
			if ok {
				t.Fatalf("Expected ok == false, got %v", ok)
			}
		})
	}
}

func TestMutexQuorum(t *testing.T) {
	ctx := context.Background()
	for k, v := range makeCases(4) {
		t.Run(k, func(t *testing.T) {
			for mask := range 1 << len(v.pools) {
				mutexes := newTestMutexes(v.pools, k+"-test-mutex-partial-"+strconv.Itoa(mask), 1)
				mutex := mutexes[0]
				mutex.tries = 1

				n := clogPools(v.pools, mask, mutex)

				if n >= len(v.pools)/2+1 {
					err := mutex.Lock()
					if err != nil {
						t.Fatalf("mutex lock failed: %s", err)
					}
					assertAcquired(ctx, t, v.pools, mutex)
				} else {
					err := mutex.Lock()
					if errors.Is(err, &ErrNodeTaken{}) {
						t.Fatalf("Expected err == %q, got %q", ErrNodeTaken{}, err)
					}
				}
			}
		})
	}
}

func TestValid(t *testing.T) {
	ctx := context.Background()
	for k, v := range makeCases(4) {
		t.Run(k, func(t *testing.T) {
			rs := New(v.pools...)
			key := k + "-test-shared-lock"

			mutex1 := rs.NewMutex(key, WithExpiry(time.Hour))
			err := mutex1.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			assertAcquired(ctx, t, v.pools, mutex1)

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

func TestMutexLockUnlockSplit(t *testing.T) {
	ctx := context.Background()
	for k, v := range makeCases(4) {
		t.Run(k, func(t *testing.T) {
			rs := New(v.pools...)
			key := "test-split-lock"

			mutex1 := rs.NewMutex(key, WithExpiry(time.Hour))
			err := mutex1.Lock()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
			assertAcquired(ctx, t, v.pools, mutex1)

			mutex2 := rs.NewMutex(key, WithExpiry(time.Hour), WithValue(mutex1.Value()))
			ok, err := mutex2.Unlock()
			if err != nil {
				t.Fatalf("mutex unlock failed: %s", err)
			}
			if !ok {
				t.Fatalf("Expected a valid mutex")
			}
		})
	}
}

func getPoolValues(ctx context.Context, pools []redis.Pool, name string) []string {
	values := make([]string, len(pools))
	for i, pool := range pools {
		conn, err := pool.Get(ctx)
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
		conn, err := pool.Get(context.TODO())
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
		conn, err := pool.Get(context.Background())
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
	for i := range n {
		mutexes[i] = &Mutex{
			name:          name + "-" + strconv.Itoa(i),
			expiry:        8 * time.Second,
			tries:         32,
			delayFunc:     func(tries int) time.Duration { return 500 * time.Millisecond },
			genValueFunc:  genValue,
			driftFactor:   0.01,
			timeoutFactor: 0.05,
			quorum:        len(pools)/2 + 1,
			pools:         pools,
		}
	}
	return mutexes
}

func isAcquired(ctx context.Context, pools []redis.Pool, mutex *Mutex) bool {
	n := countAcquiredPools(ctx, pools, mutex)
	return n >= mutex.quorum
}

func countAcquiredPools(ctx context.Context, pools []redis.Pool, mutex *Mutex) int {
	n := 0
	values := getPoolValues(ctx, pools, mutex.name)
	for _, value := range values {
		if value == mutex.value {
			n++
		}
	}
	return n
}

func assertAcquired(ctx context.Context, t *testing.T, pools []redis.Pool, mutex *Mutex) {
	n := 0
	values := getPoolValues(ctx, pools, mutex.name)
	for _, value := range values {
		if value == mutex.value {
			n++
		}
	}
	if !isAcquired(ctx, pools, mutex) {
		t.Fatalf("Expected n >= %d, got %d", mutex.quorum, n)
	}
}
