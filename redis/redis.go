package redis

import (
	"context"
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
	Get(ctx context.Context, name string) (string, error)
	Set(ctx context.Context, name string, value string) (bool, error)
	SetNX(ctx context.Context, name string, value string, expiry time.Duration) (bool, error)
	Eval(ctx context.Context, script *Script, keysAndArgs ...interface{}) (interface{}, error)
	PTTL(ctx context.Context, name string) (time.Duration, error)
	Close() error
}

type Script struct {
	KeyCount int
	Src      string
	Hash     string
}

func NewScript(keyCount int, src string) *Script {
	h := sha1.New()
	_, _ = io.WriteString(h, src)
	return &Script{keyCount, src, hex.EncodeToString(h.Sum(nil))}
}
