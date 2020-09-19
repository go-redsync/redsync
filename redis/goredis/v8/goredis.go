package goredis

import (
	"context"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
)

type pool struct {
	delegate *redis.Client
}

func (p *pool) Get(ctx context.Context) (redsyncredis.Conn, error) {
	return &conn{p.delegate}, nil
}

func NewPool(delegate *redis.Client) redsyncredis.Pool {
	return &pool{delegate}
}

type conn struct {
	delegate *redis.Client
}

func (c *conn) Get(ctx context.Context, name string) (string, error) {
	value, err := c.delegate.Get(c._context(ctx), name).Result()
	return value, noErrNil(err)
}

func (c *conn) Set(ctx context.Context, name string, value string) (bool, error) {
	reply, err := c.delegate.Set(c._context(ctx), name, value, 0).Result()
	return reply == "OK", err
}

func (c *conn) SetNX(ctx context.Context, name string, value string, expiry time.Duration) (bool, error) {
	return c.delegate.SetNX(c._context(ctx), name, value, expiry).Result()
}

func (c *conn) PTTL(ctx context.Context, name string) (time.Duration, error) {
	return c.delegate.PTTL(c._context(ctx), name).Result()
}

func (c *conn) Eval(ctx context.Context, script *redsyncredis.Script, keysAndArgs ...interface{}) (interface{}, error) {
	keys := make([]string, script.KeyCount)
	args := keysAndArgs

	if script.KeyCount > 0 {
		for i := 0; i < script.KeyCount; i++ {
			keys[i] = keysAndArgs[i].(string)
		}
		args = keysAndArgs[script.KeyCount:]
	}

	ctx = c._context(ctx)
	v, err := c.delegate.EvalSha(ctx, script.Hash, keys, args...).Result()
	if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		v, err = c.delegate.Eval(ctx, script.Src, keys, args...).Result()
	}
	return v, noErrNil(err)
}

func (c *conn) Close() error {
	// Not needed for this library
	return nil
}

func (c *conn) _context(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return c.delegate.Context()
}

func noErrNil(err error) error {
	if err == redis.Nil {
		return nil
	}
	return err
}
