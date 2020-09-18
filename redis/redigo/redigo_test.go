package redigo

import "github.com/go-redsync/redsync/v3/redis"

var _ redis.Conn = (*conn)(nil)

var _ redis.Pool = (*pool)(nil)
