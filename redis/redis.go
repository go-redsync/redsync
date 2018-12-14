package redis

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"time"
)

// A Pool maintains a pool of Redis connections.
type Pool interface {
	Get() Conn
}

type Conn interface {
	Get(name string) (string, error)
	Set(name string, value string) (bool, error)
	SetNX(name string, value string, expiry time.Duration) (bool, error)
	Eval(script *Script, keysAndArgs ...interface{}) (interface{}, error)
	PTTL(name string) (time.Duration, error)
	Close() error
}

type Script struct {
	KeyCount int
	Src      string
	Hash     string
}

func NewScript(keyCount int, src string) *Script {
	h := sha1.New()
	io.WriteString(h, src)
	return &Script{keyCount, src, hex.EncodeToString(h.Sum(nil))}
}

func (self *Script) Args(spec string, keysAndArgs []interface{}) []interface{} {
	var args []interface{}
	if self.KeyCount < 0 {
		args = make([]interface{}, 1+len(keysAndArgs))
		args[0] = spec
		copy(args[1:], keysAndArgs)
	} else {
		args = make([]interface{}, 2+len(keysAndArgs))
		args[0] = spec
		args[1] = self.KeyCount
		copy(args[2:], keysAndArgs)
	}
	return args
}
