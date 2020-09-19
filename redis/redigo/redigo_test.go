package redigo

import "github.com/go-redsync/redsync/v4/redis"

var _ redis.Conn = (*conn)(nil)

var _ redis.Pool = (*pool)(nil)
