package radix

import (
	"context"
	"fmt"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	"github.com/mediocregopher/radix/v4"
	"strconv"
	"time"
)

type PoolWrapper struct {
	cluster *radix.Cluster
}

type ConnWrapper struct {
	cluster *radix.Cluster
	ctx     context.Context
}

// NewPool returns a Redix-based pool implementation.
func NewPool(cluster *radix.Cluster) redsyncredis.Pool {
	return &PoolWrapper{cluster: cluster}
}

func (p *PoolWrapper) Get(ctx context.Context) (redsyncredis.Conn, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return &ConnWrapper{cluster: p.cluster, ctx: ctx}, nil
}

func (c *ConnWrapper) Get(name string) (string, error) {
	var ret string
	action := radix.Cmd(&ret, "GET", name)
	err := c.cluster.Do(c.ctx, action)
	return ret, err
}

func (c *ConnWrapper) Set(name string, value string) (bool, error) {
	var ret string
	action := radix.Cmd(&ret, "SET", name, value)
	err := c.cluster.Do(c.ctx, action)
	return ret == "OK", err
}

func (c *ConnWrapper) SetNX(name string, value string, expiry time.Duration) (bool, error) {
	var ret string
	action := radix.Cmd(&ret, "SET", name, value, "NX", "PX", strconv.FormatInt(int64(expiry/time.Millisecond), 10))
	err := c.cluster.Do(c.ctx, action)
	return ret == "OK", err
}

func (c *ConnWrapper) PTTL(name string) (time.Duration, error) {
	var expiry int64
	action := radix.Cmd(&expiry, "PTTL", name)
	err := c.cluster.Do(c.ctx, action)
	return time.Duration(expiry) * time.Millisecond, err
}

func (c *ConnWrapper) Eval(script *redsyncredis.Script, keysAndArgs ...interface{}) (interface{}, error) {
	evScript := radix.NewEvalScript(script.Src)
	var ret interface{}
	keys, args := args(script, keysAndArgs)
	err := c.cluster.Do(c.ctx, evScript.Cmd(&ret, keys, args...))
	return ret, err
}

func (c *ConnWrapper) Close() error {
	return nil
}

func args(script *redsyncredis.Script, keysAndArgs []interface{}) ([]string, []string) {
	var args []string
	var keys []string
	for i, arg := range keysAndArgs {
		strArg := toString(arg)
		if i >= script.KeyCount {
			args = append(args, strArg)
		} else {
			keys = append(keys, strArg)
		}
	}

	return keys, args
}

func toString(arg interface{}) string {
	switch arg.(type) {
	case string:
		return arg.(string)
	case int:
		return strconv.Itoa(arg.(int))
	}
	//FIXME: add more types if required
	return fmt.Sprintf("%v", arg)
}
