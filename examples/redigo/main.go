package main

import (
	"time"

	redigolib "github.com/gomodule/redigo/redis"
	"github.com/stvp/tempredis"
	"github.com/weylan/redsync"
	"github.com/weylan/redsync/redis/redigo"
)

func main() {
	server, err := tempredis.Start(tempredis.Config{})
	if err != nil {
		panic(err)
	}
	defer server.Term()

	pool := redigo.NewPool(&redigolib.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigolib.Conn, error) {
			return redigolib.Dial("unix", server.Socket())
		},
		TestOnBorrow: func(c redigolib.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	})

	rs := redsync.New(pool)

	mutex := rs.NewMutex("test-redsync")

	if err = mutex.Lock(); err != nil {
		panic(err)
	}

	if _, err = mutex.Unlock(); err != nil {
		panic(err)
	}
}
