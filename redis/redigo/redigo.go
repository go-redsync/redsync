package redigo

import (
	"strings"
	"time"

	redsyncredis "github.com/go-redsync/redsync/v3/redis"
	"github.com/gomodule/redigo/redis"
)

type RedigoPool struct {
	delegate *redis.Pool
}

func (self *RedigoPool) Get() redsyncredis.Conn {
	return &RedigoConn{self.delegate.Get()}
}

func NewRedigoPool(delegate *redis.Pool) *RedigoPool {
	return &RedigoPool{delegate}
}

type RedigoConn struct {
	delegate redis.Conn
}

func (self *RedigoConn) Get(name string) (string, error) {
	value, err := redis.String(self.delegate.Do("GET", name))
	err = noErrNil(err)
	return value, err
}

func (self *RedigoConn) Set(name string, value string) (bool, error) {
	reply, err := redis.String(self.delegate.Do("SET", name, value))
	err = noErrNil(err)
	return err == nil && reply == "OK", nil
}

func (self *RedigoConn) SetNX(name string, value string, expiry time.Duration) (bool, error) {
	reply, err := redis.String(self.delegate.Do("SET", name, value, "NX", "PX", int(expiry/time.Millisecond)))
	err = noErrNil(err)
	return err == nil && reply == "OK", nil
}

func (self *RedigoConn) PTTL(name string) (time.Duration, error) {
	expiry, err := redis.Int64(self.delegate.Do("PTTL", name))
	err = noErrNil(err)
	return time.Duration(expiry) * time.Millisecond, err
}

func (self *RedigoConn) Eval(script *redsyncredis.Script, keysAndArgs ...interface{}) (interface{}, error) {
	v, err := self.delegate.Do("EVALSHA", args(script, script.Hash, keysAndArgs)...)
	if e, ok := err.(redis.Error); ok && strings.HasPrefix(string(e), "NOSCRIPT ") {
		v, err = self.delegate.Do("EVAL", args(script, script.Src, keysAndArgs)...)
	}
	return v, err

}

func (self *RedigoConn) Close() error {
	err := self.delegate.Close()
	return noErrNil(err)
}

func noErrNil(err error) error {
	if err == redis.ErrNil {
		return nil
	} else {
		return err
	}
}

func args(script *redsyncredis.Script, spec string, keysAndArgs []interface{}) []interface{} {
	var args []interface{}
	if script.KeyCount < 0 {
		args = make([]interface{}, 1+len(keysAndArgs))
		args[0] = spec
		copy(args[1:], keysAndArgs)
	} else {
		args = make([]interface{}, 2+len(keysAndArgs))
		args[0] = spec
		args[1] = script.KeyCount
		copy(args[2:], keysAndArgs)
	}
	return args
}
