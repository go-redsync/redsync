package goredis

import (
	"strings"
	"time"

	"github.com/go-redis/redis"
	redsyncredis "github.com/go-redsync/redsync/v3/redis"
)

type GoredisPool struct {
	delegate *redis.Client
}

func (self *GoredisPool) Get() redsyncredis.Conn {
	return &GoredisConn{self.delegate}
}

func NewGoredisPool(delegate *redis.Client) *GoredisPool {
	return &GoredisPool{delegate}
}

type GoredisConn struct {
	delegate *redis.Client
}

func (self *GoredisConn) Get(name string) (string, error) {
	value, err := self.delegate.Get(name).Result()
	return value, noErrNil(err)
}

func (self *GoredisConn) Set(name string, value string) (bool, error) {
	reply, err := self.delegate.Set(name, value, 0).Result()
	return reply == "OK", noErrNil(err)
}

func (self *GoredisConn) SetNX(name string, value string, expiry time.Duration) (bool, error) {
	ok, err := self.delegate.SetNX(name, value, expiry).Result()
	return ok, noErrNil(err)
}

func (self *GoredisConn) PTTL(name string) (time.Duration, error) {
	expiry, err := self.delegate.PTTL(name).Result()
	return expiry, noErrNil(err)
}

func (self *GoredisConn) Eval(script *redsyncredis.Script, keysAndArgs ...interface{}) (interface{}, error) {
	keys := make([]string, script.KeyCount)
	args := keysAndArgs

	if script.KeyCount > 0 {
		for i := 0; i < script.KeyCount; i++ {
			keys[i] = keysAndArgs[i].(string)
		}

		args = keysAndArgs[script.KeyCount:]
	}

	v, err := self.delegate.EvalSha(script.Hash, keys, args...).Result()
	if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		v, err = self.delegate.Eval(script.Src, keys, args...).Result()
	}
	return v, noErrNil(err)
}

func (self *GoredisConn) Close() error {
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
