package goredis

import (
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
)

type pool struct {
	delegate *redis.Client
}

func (p *pool) Get(ctx context.Context) (redsyncredis.Conn, error) {
	c := p.delegate
	if ctx != nil {
		c = c.WithContext(ctx)
	}
	return &conn{c}, nil
}

func NewPool(delegate *redis.Client) redsyncredis.Pool {
	return &pool{delegate}
}

type conn struct {
	delegate *redis.Client
}

func (c *conn) Get(name string) (string, error) {
	value, err := c.delegate.Get(name).Result()
	return value, noErrNil(err)
}

func (c *conn) Set(name string, value string) (bool, error) {
	reply, err := c.delegate.Set(name, value, 0).Result()
	return reply == "OK", noErrNil(err)
}

func (c *conn) SetNX(name string, value string, expiry time.Duration) (bool, error) {
	ok, err := c.delegate.SetNX(name, value, expiry).Result()
	return ok, noErrNil(err)
}

func (c *conn) PTTL(name string) (time.Duration, error) {
	expiry, err := c.delegate.PTTL(name).Result()
	return expiry, noErrNil(err)
}

func (c *conn) Eval(script *redsyncredis.Script, keysAndArgs ...interface{}) (interface{}, error) {
	keys := make([]string, script.KeyCount)
	args := keysAndArgs

	if script.KeyCount > 0 {
		for i := 0; i < script.KeyCount; i++ {
			keys[i] = keysAndArgs[i].(string)
		}

		args = keysAndArgs[script.KeyCount:]
	}

	v, err := c.delegate.EvalSha(script.Hash, keys, args...).Result()
	if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		v, err = c.delegate.Eval(script.Src, keys, args...).Result()
	}
	return v, noErrNil(err)
}

func (c *conn) Close() error {
	// Not needed for this library
	return nil
}

func noErrNil(err error) error {
	if err == redis.Nil {
		return nil
	} else {
		return err
	}
}
