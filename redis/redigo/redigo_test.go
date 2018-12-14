package redigo

import "github.com/go-redsync/redsync/redis"

var _ (redis.Conn) = (*RedigoConn)(nil)

var _ (redis.Pool) = (*RedigoPool)(nil)
