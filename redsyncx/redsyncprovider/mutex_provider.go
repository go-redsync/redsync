package redsyncprovider

import (
	"context"

	"github.com/cabify/redsync/v2/redsyncx"
)

//go:generate mockery -case underscore -output redsyncprovidertest -outpkg redsyncprovidertest -name Provider

// Provider should provide Mutex for a key
type Provider interface {
	For(ctx context.Context, key string) redsyncx.Mutex
}
