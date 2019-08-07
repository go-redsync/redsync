package redsyncx

import (
	"context"
	"fmt"
	"time"

	"github.com/cabify/redsync/v2"
)

//go:generate mockery -case underscore -output redsynctest -outpkg redsynctest -name Redsync

// Redsync is the interface implemented by *redsync.Redsync, except concrete
// types have been converted to interfaces and it's mandatory to specify a TTL.
type Redsync interface {
	NewMutex(ctx context.Context, name string, ttl time.Duration, options ...Option) Mutex
}

//go:generate mockery -case underscore -output redsynctest -outpkg redsynctest -name Mutex

// Mutex describes a mutex.
//
// Any error returned will be one of the errors defined in the redsync package.
type Mutex interface {
	// Lock locks the mutex.
	//
	// If the provided context finishes, or Unlock is called, or the lock is lost
	// for some other reason, the returned context is cancelled.
	Lock(context.Context) (context.Context, error)
	Unlock() error
}

// New is a wrapper for redsync.New that returns an interface.
//
// goFunc is a function that must start a goroutine running the function given
// to it. It should keep the basic semantics of the plain go statement.
func New(pools []redsync.Pool, goFunc func(func())) Redsync {
	if len(pools) == 0 {
		panic("redsyncx: no pools specified")
	}
	return wrapper{r: redsync.New(pools), goFunc: goFunc}
}

type wrapper struct {
	r      *redsync.Redsync
	goFunc func(func())
}

func (w wrapper) NewMutex(ctx context.Context, name string, ttl time.Duration, options ...Option) Mutex {
	bridgedOpts := make([]redsync.Option, 0, len(options)+1)
	withCancelOpts := make([]func(*mutexWithCancelOnLockLost), 0, len(options))
	for _, o := range options {
		switch o := o.(type) {
		case bridgedOption:
			bridgedOpts = append(bridgedOpts, o.bridged)
		case withCancelOption:
			withCancelOpts = append(withCancelOpts, o)
		default:
			panic(fmt.Errorf("unexpected Option: %T", o))
		}
	}

	bridgedOpts = append(bridgedOpts, redsync.SetExpiry(ttl))

	mtx := w.r.NewMutex(ctx, name, bridgedOpts...)
	return withCancelOnLockLost(ttl, mtx, w.goFunc, withCancelOpts...)
}

// An Option configures a Mutex.
type Option interface {
	isOption()
}

type bridgedOption struct {
	bridged redsync.Option
}

func (bridgedOption) isOption() {}

type withCancelOption func(*mutexWithCancelOnLockLost)

func (withCancelOption) isOption() {}

// SetDriftFactor is a wrapper for redsync.SetDriftFactor.
func SetDriftFactor(factor float64) Option {
	return bridgedOption{bridged: redsync.SetDriftFactor(factor)}
}

// SetRetryDelay is a wrapper for redsync.SetRetryDelay.
func SetRetryDelay(delay time.Duration) Option {
	return bridgedOption{bridged: redsync.SetRetryDelay(delay)}
}

// SetTries is a wrapper for redsync.SetTries.
func SetTries(tries int) Option {
	return bridgedOption{bridged: redsync.SetTries(tries)}
}

// OnExtendError sets a function that will be called with the non-nil errors
// returned when extending the lock.
func OnExtendError(f func(error)) Option {
	return withCancelOption(onExtendErrorOption(f))
}

// LogExtendError returns a function that can be used to log errors returned by
// Extend, to be used with OnExtendError.
//
// Bring your own log functions.
func LogExtendError(logTaken, logOther func(string, ...interface{})) func(error) {
	return func(err error) {
		log := logOther
		if err == redsync.ErrTaken {
			log = logTaken
		}
		log("Couldn't extend lock: %s", err)
	}
}
