package radix

import "github.com/go-redsync/redsync/v4/redis"

var _ redis.Conn = (*ConnWrapper)(nil)

var _ redis.Pool = (*PoolWrapper)(nil)
