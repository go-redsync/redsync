package redsync

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/go-redsync/redsync/v4/redis"
)

func WithKeys(keys []string) Option {
	return OptionFunc(func(m *Mutex) {
		m.keysForMultipleLock = keys
	})
}

func (m *Mutex) multipleRelease(ctx context.Context, pool redis.Pool, keys []string) (bool, error) {
	conn, err := pool.Get(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	script := redis.NewScript(len(keys), multipleDeleteScript)

	keysForDelete := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		keysForDelete = append(keysForDelete, key)
	}

	_, err = conn.Eval(script, keysForDelete...)
	if err != nil {
		return false, err
	}

	return true, nil
}

var multipleDeleteScript = `
	local function delete_keys(keys)
		for i, key in ipairs(keys) do
			redis.call('DEL', key)
		end
	end
	local keys = ARGV
	delete_keys(keys)
	return #keys
	`

var multipleLockScript = `
	local key_value_pairs = {}
	for i, key in ipairs(KEYS) do
		key_value_pairs[#key_value_pairs + 1] = key
		key_value_pairs[#key_value_pairs + 1] = "" -- Пустое значение
	end
		
	local msetnx_result = redis.call("MSETNX", unpack(key_value_pairs))
		
	if msetnx_result == 0 then
		return 0
	end
		
	local timeout_ms = tonumber(ARGV[1])
	local timeout_seconds = math.floor(timeout_ms / 1000)

	for i, key in ipairs(KEYS) do
		redis.call("EXPIRE", key, timeout_seconds)
	end
	`

var (
	ErrNothingToLock = errors.New("nothing to lock")
	ErrEmptyExpire   = errors.New("empty expire")
)

func (m *Mutex) MultipleLockContext(ctx context.Context) error {
	return m.multipleLock(ctx, m.tries)
}

func (m *Mutex) UnlockMultipleContext(ctx context.Context) (bool, error) {
	n, err := m.actOnPoolsAsync(func(pool redis.Pool) (bool, error) {
		return m.multipleRelease(ctx, pool, m.keysForMultipleLock)
	})
	if n < m.quorum {
		return false, err
	}
	return true, nil
}

func (m *Mutex) multipleLock(ctx context.Context, tries int) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(m.keysForMultipleLock) == 0 {
		return ErrNothingToLock
	}

	if m.expiry == 0 {
		return ErrEmptyExpire
	}

	value, err := m.genValueFunc()
	if err != nil {
		return err
	}

	var timer *time.Timer
	for i := 0; i < tries; i++ {
		if i != 0 {
			if timer == nil {
				timer = time.NewTimer(m.delayFunc(i))
			} else {
				timer.Reset(m.delayFunc(i))
			}

			select {
			case <-ctx.Done():
				timer.Stop()
				// Exit early if the context is done.
				return ErrFailed
			case <-timer.C:
				// Fall-through when the delay timer completes.
			}
		}

		start := time.Now()

		n, err := func() (int, error) {
			ctx, cancel := context.WithTimeout(ctx, time.Duration(int64(float64(m.expiry)*m.timeoutFactor)))
			defer cancel()
			return m.actOnPoolsAsync(func(pool redis.Pool) (bool, error) {
				return m.multipleAcquire(ctx, pool)
			})
		}()

		now := time.Now()
		until := now.Add(m.expiry - now.Sub(start) - time.Duration(int64(float64(m.expiry)*m.driftFactor)))
		if n >= m.quorum && now.Before(until) {
			m.value = value
			m.until = until
			return nil
		}

		_, _ = func() (int, error) {
			ctx, cancel := context.WithTimeout(ctx, time.Duration(int64(float64(m.expiry)*m.timeoutFactor)))
			defer cancel()
			return m.actOnPoolsAsync(func(pool redis.Pool) (bool, error) {
				return m.multipleRelease(ctx, pool, m.keysForMultipleLock)
			})
		}()
		if i == tries-1 && err != nil {
			return err
		}
	}

	return ErrFailed
}

func (m *Mutex) multipleAcquire(ctx context.Context, pool redis.Pool) (bool, error) {
	conn, err := pool.Get(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	script := redis.NewScript(len(m.keysForMultipleLock), multipleLockScript)

	args := make([]interface{}, 0, len(m.keysForMultipleLock)+1)
	for _, value := range m.keysForMultipleLock {
		args = append(args, value)
	}
	args = append(args, strconv.Itoa(int(m.expiry.Milliseconds())))

	status, err := conn.Eval(script, args...)
	if err != nil {
		return false, err
	}

	return status != int64(0), nil
}
