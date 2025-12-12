package redsync

import (
	"context"
	"testing"

	"github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib_v9 "github.com/redis/go-redis/v9"
	"github.com/stvp/tempredis"
)

func TestScripts_Hash(t *testing.T) {
	testScriptHash(t, "deleteScript", deleteScript)
	testScriptHash(t, "touchWithSetNXScript", touchWithSetNXScript)
	testScriptHash(t, "touchScript", touchScript)
}

func testScriptHash(t *testing.T, name string, script *redis.Script) {
	t.Run(name, func(t *testing.T) {
		server, err := tempredis.Start(tempredis.Config{})
		if err != nil {
			t.Fatal(err)
		}
		defer server.Term()

		client := goredislib_v9.NewClient(&goredislib_v9.Options{
			Network: "unix",
			Addr:    server.Socket(),
		})

		pool := goredis.NewPool(client)
		conn, err := pool.Get(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		dupscript := redis.NewScript(script.KeyCount, script.Src, "")
		err = conn.ScriptLoad(dupscript)
		if err != nil {
			t.Fatal(err)
		}

		if dupscript.Hash == "" {
			t.Fatal("script hash is empty")
		}
		if dupscript.Hash != script.Hash {
			t.Fatalf("script hash mismatch; want %s, got %s", script.Hash, dupscript.Hash)
		}
	})
}
