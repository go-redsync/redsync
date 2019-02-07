package redsync

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

// A DelayFunc is used to decide the amount of time to wait between retries.
type DelayFunc func(tries int) time.Duration

// A Mutex is a distributed mutual exclusion lock.
type Mutex struct {
	name   string
	expiry time.Duration

	tries     int
	delayFunc DelayFunc

	factor float64

	quorum int

	value string
	until time.Time

	nodem sync.Mutex

	pools []Pool
}

// Lock locks m. In case it returns an error on failure, you may retry to
// acquire the lock by calling this method again.
//
// If not nil, the error will be either ErrTaken or a NoQuorum, which in turn
// contains NodeTaken or RedisError.
//
// Only the last try's errors are returned.
func (m *Mutex) Lock() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	value, err := m.genValue()
	if err != nil {
		return err
	}

	err = nil

	for i := 0; i < m.tries; i++ {
		if i != 0 {
			time.Sleep(m.delayFunc(i))
		}

		start := time.Now()

		var n int
		n, err = m.actOnPoolsAsync(func(pool Pool) (bool, error) {
			return m.acquire(pool, value)
		})

		until := time.Now().Add(m.expiry - time.Now().Sub(start) - time.Duration(int64(float64(m.expiry)*m.factor)) + 2*time.Millisecond)
		if n >= m.quorum && time.Now().Before(until) {
			m.value = value
			m.until = until
			return nil
		}
		m.actOnPoolsAsync(func(pool Pool) (bool, error) {
			return m.release(pool, value)
		})
	}

	return err
}

// Unlock unlocks m.
//
// If not nil, the error will be either ErrTaken or a NoQuorum, which in turn
// contains NodeTaken or RedisError.
func (m *Mutex) Unlock() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	n, errs := m.actOnPoolsAsync(func(pool Pool) (bool, error) {
		return m.release(pool, m.value)
	})
	if n >= m.quorum {
		return nil
	}
	return errs
}

// Extend resets the mutex's expiry.
//
// If not nil, the error will be either ErrTaken or a NoQuorum, which in turn
// contains NodeTaken or RedisError.
func (m *Mutex) Extend() error {
	m.nodem.Lock()
	defer m.nodem.Unlock()

	n, errs := m.actOnPoolsAsync(func(pool Pool) (bool, error) {
		return m.touch(pool, m.value, int(m.expiry/time.Millisecond))
	})
	if n >= m.quorum {
		return nil
	}
	return errs
}

func (m *Mutex) genValue() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (m *Mutex) acquire(pool Pool, value string) (bool, error) {
	conn := pool.Get()
	defer conn.Close()
	reply, err := redis.String(conn.Do("SET", m.name, value, "NX", "PX", int(m.expiry/time.Millisecond)))
	if err != nil && err != redis.ErrNil {
		return false, err
	}
	return reply == "OK", nil
}

var deleteScript = redis.NewScript(1, `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
`)

func (m *Mutex) release(pool Pool, value string) (bool, error) {
	conn := pool.Get()
	defer conn.Close()
	status, err := deleteScript.Do(conn, m.name, value)
	if err != nil {
		return false, err
	}
	return status != 0, nil
}

var touchScript = redis.NewScript(1, `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("SET", KEYS[1], ARGV[1], "XX", "PX", ARGV[2])
	else
		return "ERR"
	end
`)

func (m *Mutex) touch(pool Pool, value string, expiry int) (bool, error) {
	conn := pool.Get()
	defer conn.Close()
	status, err := redis.String(touchScript.Do(conn, m.name, value, expiry))
	if err != nil {
		return false, err
	}
	return status != "ERR", nil
}

func (m *Mutex) actOnPoolsAsync(actFn func(Pool) (bool, error)) (ok int, err error) {
	type resp struct {
		node int
		ok   bool
		err  error
	}

	ch := make(chan resp)
	for node, pool := range m.pools {
		go func(node int, pool Pool) {
			ok, err := actFn(pool)
			ch <- resp{node: node, ok: ok, err: err}
		}(node, pool)
	}

	var errs []error
	hasRedisErrors := false

	for range m.pools {
		resp := <-ch
		if err := resp.err; err != nil {
			hasRedisErrors = true
			errs = append(errs, RedisError{Node: resp.node, Err: err})
		} else if !resp.ok {
			errs = append(errs, NodeTaken{Node: resp.node})
		} else {
			ok++
		}
	}

	if !hasRedisErrors && len(errs) > 0 {
		return ok, ErrTaken
	}

	return ok, NoQuorum(errs)
}
