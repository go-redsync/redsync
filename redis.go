package redsync

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

// A Pool maintains a pool of Redis connections.
type Pool interface {
	GetContext(context.Context) (redis.Conn, error)
}
