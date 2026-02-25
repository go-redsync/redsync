package main

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/valkeygo"
	"github.com/stvp/tempredis"
	valkeylib "github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeycompat"
)

func main() {
	server, err := tempredis.Start(tempredis.Config{})
	if err != nil {
		panic(err)
	}
	defer server.Term()

	client, err := valkeylib.NewClient(valkeylib.ClientOption{
		InitAddress: []string{server.Socket()},
		DialCtxFn: func(ctx context.Context, s string, d *net.Dialer, c *tls.Config) (conn net.Conn, err error) {
			return d.DialContext(ctx, "unix", s)
		},
	})
	if err != nil {
		panic(err)
	}

	pool := valkeygo.NewPool(valkeycompat.NewAdapter(client))

	rs := redsync.New(pool)

	mutex := rs.NewMutex("test-redsync")

	if err = mutex.Lock(); err != nil {
		panic(err)
	}

	if _, err = mutex.Unlock(); err != nil {
		panic(err)
	}
}
