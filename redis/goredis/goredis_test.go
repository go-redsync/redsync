package goredis

import "github.com/go-redsync/redsync/redis"

var _ (redis.Conn) = (*GoredisConn)(nil)

var _ (redis.Pool) = (*GoredisPool)(nil)
