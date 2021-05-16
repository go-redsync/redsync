package redsync

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/weylan/redsync/redis"
)

// A DelayFunc is used to decide the amount of time to wait between retries.
type DelayFunc func(tries int) time.Duration

// A Mutex is a distributed mutual exclusion lock.
type Mutex struct {
	name        string
	expiry      time.Duration
	processTime time.Duration

	tries     int
	delayFunc DelayFunc

	factor float64

	quorum int

	genValueFunc func() (string, error)
	version      int64
	until        time.Time

	pools redis.Pool
}

// Name returns mutex name (i.e. the Redis key).
func (m *Mutex) Name() string {
	return m.name
}

func (m *Mutex) Version() int64 {
	return m.version
}

// Lock locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *Mutex) Lock() error {
	return m.LockContext(nil)
}

// LockContext locks m. In case it returns an error on failure, you may retry to acquire the lock by calling this method again.
func (m *Mutex) LockContext(ctx context.Context) error {
	var err error
	var ok bool

	for i := 0; i < m.tries; i++ {
		if i != 0 {
			time.Sleep(m.delayFunc(i))
		}

		start := time.Now()

		ok, err = m.acquire(ctx, m.pools)
		if !ok && err != nil {
			return err
		}

		now := time.Now()
		until := now.Add(m.expiry - now.Sub(start) - time.Duration(int64(float64(m.expiry)*m.factor)))
		if ok && now.Before(until) {
			m.until = until
			return nil
		}
		m.release(ctx, m.pools)
	}

	return ErrFailed
}

// Unlock unlocks m and returns the status of unlock.
func (m *Mutex) Unlock() (bool, error) {
	return m.UnlockContext(nil)
}

// UnlockContext unlocks m and returns the status of unlock.
func (m *Mutex) UnlockContext(ctx context.Context) (bool, error) {
	return m.release(ctx, m.pools)
}

// Extend resets the mutex's expiry and returns the status of expiry extension.
func (m *Mutex) Extend() (bool, error) {
	return m.ExtendContext(nil)
}

// ExtendContext resets the mutex's expiry and returns the status of expiry extension.
func (m *Mutex) ExtendContext(ctx context.Context) (bool, error) {
	return m.touch(ctx, m.pools, int(m.expiry/time.Second))
}

func (m *Mutex) Valid() (bool, error) {
	return m.ValidContext(nil)
}

func (m *Mutex) ValidContext(ctx context.Context) (bool, error) {
	return m.valid(ctx, m.pools)
}

func (m *Mutex) valid(ctx context.Context, pool redis.Pool) (bool, error) {
	if time.Now().After(m.until) {
		return false, nil
	}
	conn, err := pool.Get(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	reply, err := conn.HGet(m.name, "version")
	if err != nil {
		return false, err
	}
	version, err := strconv.Atoi(reply)
	if err != nil {
		return false, err
	}
	return m.version == int64(version*-1), nil
}

func genValue() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

var lockScript = redis.NewScript(1, `
	local key = KEYS[1]
	local result = redis.call("HMGET", key, "version", "expire")
	local cur = redis.call("time")[1]
	local value = result[1]
	local expire = result[2]

	if type(value) == "string" then
		value = tonumber(value)
	end

	if value and value <= 0 and cur <= expire then
		return 0
	end

	if not value or value == 0 then
		value = 1
	elseif value <= 0 then
		value = value * -1 + 1
	end

	expire = ARGV[1] + cur
	redis.call("HMSET", key, "version", value * -1, "expire", expire)
	return value
`)

func (m *Mutex) acquire(ctx context.Context, pool redis.Pool) (bool, error) {
	conn, err := pool.Get(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	status, err := conn.Eval(lockScript, m.name, int(m.expiry/time.Second))
	if err != nil {
		return false, err
	}
	if version, ok := status.(int64); ok && version != 0 {
		m.version = version
	}
	return status != int64(0), nil
}

var releaseScript = redis.NewScript(1, `
	local key = KEYS[1]
	local result = redis.call("HMGET", key, "version", "expire")
	local value = result[1]
	local expire = result[2]
	
	if not value then
		return 0
	end

	if type(value) == "string" then
		value = tonumber(value)
	end

	if value >= 0 then
		return 0
	end

	if value == -1 * ARGV[1] then
		value = ARGV[1] + 1
	else
		return -1 * value
	end

	redis.call("HMSET", key, "version", value, "expire", 0)
	return value
`)

func (m *Mutex) release(ctx context.Context, pool redis.Pool) (bool, error) {
	conn, err := pool.Get(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	status, err := conn.Eval(releaseScript, m.name, int(m.version))
	if err != nil {
		return false, err
	}
	if version, ok := status.(int64); ok && version != 0 {
		return m.version+1 == version, nil
	}
	return status != int64(0), nil
}

var touchScript = redis.NewScript(1, `
	local key = KEYS[1]
	local result = redis.call("HMGET", key, "version", "expire")
	local value = result[1]
	local expire = result[2]
	
	if type(value) == "string" then
		value = tonumber(value)
	end

	if value == -1 * ARGV[1] then
		local cur = redis.call("time")[1]
		if cur > expire then
			return 0
		end
		redis.call("HSET", key, "expire", cur + ARGV[2])
		return value * -1
	else
		return 0
	end
`)

func (m *Mutex) touch(ctx context.Context, pool redis.Pool, expiry int) (bool, error) {
	conn, err := pool.Get(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	status, err := conn.Eval(touchScript, m.name, int(m.version), expiry)
	if err != nil {
		return false, err
	}
	return status != int64(0), nil
}

//func (m *Mutex) actOnPoolsAsync(actFn func(redis.Pool) (bool, error)) (int, error) {
//	type result struct {
//		Status bool
//		Err    error
//	}
//
//	ch := make(chan result)
//	for _, pool := range m.pools {
//		go func(pool redis.Pool) {
//			r := result{}
//			r.Status, r.Err = actFn(pool)
//			ch <- r
//		}(pool)
//	}
//	n := 0
//	var err error
//	for range m.pools {
//		r := <-ch
//		if r.Status {
//			n++
//		} else if r.Err != nil {
//			err = multierror.Append(err, r.Err)
//		}
//	}
//	return n, err
//}
