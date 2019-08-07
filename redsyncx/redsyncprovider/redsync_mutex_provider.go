package redsyncprovider

import (
	"context"
	"time"

	"github.com/cabify/redsync/v2/redsyncx"
)

// Config configures a Provider.
type Config struct {
	// TTL of the external lock.
	TTL time.Duration `default:"10s"`
	// RetryDelay is the wait between retries to get the lock before
	// giving up.
	RetryDelay time.Duration `default:"100ms"`
	// Tries is the amount of intents to try to acquire the lock
	Tries int `default:"5"`
}

type provider struct {
	wrapped redsyncx.Redsync
	cfg     Config
	logFor  func(context.Context) Logger
}

// New is a constructor of Provider.
func New(redsync redsyncx.Redsync, cfg Config, logFor func(context.Context) Logger) Provider {
	return provider{
		wrapped: redsync,
		cfg:     cfg,
		logFor:  logFor,
	}
}

// For implements Provider and returns redsync mutexes that implement Mutex
func (p provider) For(ctx context.Context, key string) redsyncx.Mutex {
	log := p.logFor(ctx)
	return p.wrapped.NewMutex(ctx, key, p.cfg.TTL,
		redsyncx.SetTries(p.cfg.Tries),
		redsyncx.SetRetryDelay(p.cfg.RetryDelay),
		redsyncx.OnExtendError(redsyncx.LogExtendError(log.Infof, log.Errorf)),
	)
}

// Logger is the interface this package depends on to log extend errors.
type Logger interface {
	Infof(string, ...interface{})
	Errorf(string, ...interface{})
}
