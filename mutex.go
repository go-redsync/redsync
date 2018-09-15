package redsync

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

// A Mutex is a distributed mutual exclusion lock.
type Mutex struct {
	name   string
	expiry time.Duration

	tries int
	delay time.Duration

	factor float64

	quorum int

	value string
	until time.Time

	nodem sync.Mutex

	pools []Pool
}

// Lock locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *Mutex) Lock() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	value, err := m.genValue()
	if err != nil {
		return err
	}

	for i := 0; i < m.tries; i++ {
		if i != 0 {
			time.Sleep(m.delay)
		}

		start := time.Now()

		n := m.acquireAsync(value)

		until := time.Now().Add(m.expiry - time.Now().Sub(start) - time.Duration(int64(float64(m.expiry)*m.factor)) + 2*time.Millisecond)
		if n >= m.quorum && time.Now().Before(until) {
			m.value = value
			m.until = until
			return nil
		}
		for _, pool := range m.pools {
			m.release(pool, value)
		}
	}

	return ErrFailed
}

// Unlock unlocks m and returns the status of unlock.
func (m *Mutex) Unlock() bool {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	return m.releaseAsync(m.value) >= m.quorum
}

// Extend resets the mutex's expiry and returns the status of expiry extension.
func (m *Mutex) Extend() bool {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	return m.touchAsync(m.value, int(m.expiry/time.Millisecond)) >= m.quorum
}

func (m *Mutex) genValue() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (m *Mutex) acquireAsync(value string) int {
	ch := make(chan bool)
	for _, pool := range m.pools {
		go func(pool Pool) {
			ch <- m.acquire(pool, value)
		}(pool)
	}
	n := 0
	for range m.pools {
		if <-ch {
			n++
		}
	}
	return n
}

func (m *Mutex) acquire(pool Pool, value string) bool {
	conn := pool.Get()
	defer conn.Close()
	reply, err := redis.String(conn.Do("SET", m.name, value, "NX", "PX", int(m.expiry/time.Millisecond)))
	return err == nil && reply == "OK"
}

func (m *Mutex) releaseAsync(value string) int {
	ch := make(chan bool)
	for _, pool := range m.pools {
		go func(pool Pool) {
			ch <- m.release(pool, value)
		}(pool)
	}
	n := 0
	for range m.pools {
		if <-ch {
			n++
		}
	}
	return n
}

var deleteScript = redis.NewScript(1, `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
`)

func (m *Mutex) release(pool Pool, value string) bool {
	conn := pool.Get()
	defer conn.Close()
	status, err := deleteScript.Do(conn, m.name, value)
	return err == nil && status != 0
}

func (m *Mutex) touchAsync(value string, expiry int) int {
	ch := make(chan bool)
	for _, pool := range m.pools {
		go func(pool Pool) {
			ch <- m.touch(pool, value, expiry)
		}(pool)
	}
	n := 0
	for range m.pools {
		if <-ch {
			n++
		}
	}
	return n
}

var touchScript = redis.NewScript(1, `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("SET", KEYS[1], ARGV[1], "XX", "PX", ARGV[2])
	else
		return "ERR"
	end
`)

func (m *Mutex) touch(pool Pool, value string, expiry int) bool {
	conn := pool.Get()
	defer conn.Close()
	status, err := redis.String(touchScript.Do(conn, m.name, value, expiry))
	return err == nil && status != "ERR"
}
