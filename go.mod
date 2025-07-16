module github.com/go-redsync/redsync/v4

go 1.23.0

require (
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-redis/redis/v7 v7.4.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gomodule/redigo v1.9.2
	github.com/hashicorp/go-multierror v1.1.1
	github.com/redis/go-redis/v9 v9.11.0
	github.com/redis/rueidis v1.0.63
	github.com/redis/rueidis/rueidiscompat v1.0.63
	github.com/stvp/tempredis v0.0.0-20181119212430-b82af8480203
	golang.org/x/sync v0.16.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
)

replace github.com/stvp/tempredis => github.com/hjr265/tempredis v0.0.0-20231015061547-ad8aa5a343a2
