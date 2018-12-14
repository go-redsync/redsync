package goredis

import (
	"github.com/go-redis/redis"
	redsyncredis "github.com/go-redsync/redsync/redis"
	"strings"
	"time"
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
	err = noErrNil(err)
	return value, err
}

func (self *GoredisConn) Set(name string, value string) (bool, error) {
	reply, err := self.delegate.Set(name, value, 0).Result()
	return err == nil && reply == "OK", nil
}

func (self *GoredisConn) SetNX(name string, value string, expiry time.Duration) (bool, error) {
	return self.delegate.SetNX(name, value, expiry).Result()
}

func (self *GoredisConn) PTTL(name string) (time.Duration, error) {
	return self.delegate.PTTL(name).Result()
}

func (self *GoredisConn) Eval(script *redsyncredis.Script, keysAndArgs ...interface{}) (interface{}, error) {
	var keys []string
	var args []interface{}

	if script.KeyCount > 0 {

		keys = []string{}

		for i := 0; i < script.KeyCount; i++ {
			keys = append(keys, keysAndArgs[i].(string))
		}

		args = keysAndArgs[script.KeyCount:]

	} else {
		keys = []string{}
		args = keysAndArgs
	}

	v, err := self.delegate.EvalSha(script.Hash, keys, args...).Result()
	if err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		v, err = self.delegate.Eval(script.Src, keys, args...).Result()
	}
	err = noErrNil(err)
	return v, err
}

func (self *GoredisConn) Close() error {
	// Not needed for this library
	return nil
}

func noErrNil(err error) error {

	if err != nil && err.Error() == "redis: nil" {
		return nil
	} else {
		return err
	}

}
