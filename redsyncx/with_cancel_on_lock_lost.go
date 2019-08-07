package redsyncx

import (
	"context"
	"time"

	"gitlab.otters.xyz/product/journey/timex.git"
)

// withCancelOnLockLost turns a mutex that can be extended into a
// statemachine.Mutex.
//
// When the returned mutex's Lock method is called with a context, a goroutine
// starts extending the lock each ttl/2. If this fails, Unlock is called, or the
// context provided to Lock finishes, the returned lock is cancelled.
func withCancelOnLockLost(ttl time.Duration, mtx extendableMutex, goFunc func(func()), options ...func(*mutexWithCancelOnLockLost)) Mutex {
	m := mutexWithCancelOnLockLost{
		ttl:      ttl,
		wrapped:  mtx,
		goFunc:   goFunc,
		unlocked: make(chan struct{}),
	}
	for _, o := range options {
		o(&m)
	}
	return m
}

type mutexWithCancelOnLockLost struct {
	ttl      time.Duration
	wrapped  extendableMutex
	goFunc   func(func())
	unlocked chan struct{}

	onExtendError func(error)
}

func (mtx mutexWithCancelOnLockLost) Lock(parent context.Context) (context.Context, error) {
	err := mtx.wrapped.Lock()
	if err != nil {
		return parent, err
	}

	ctx, cancel := context.WithCancel(parent)

	mtx.goFunc(func() {
		for {
			select {
			case <-parent.Done():
			case <-mtx.unlocked:
			case <-timex.After(mtx.ttl / 2):
				err := mtx.wrapped.Extend()
				if err == nil {
					continue
				}
				if on := mtx.onExtendError; on != nil {
					on(err)
				}
			}

			cancel()
			return
		}
	})

	return ctx, nil
}

func (mtx mutexWithCancelOnLockLost) Unlock() error {
	close(mtx.unlocked)
	return mtx.wrapped.Unlock()
}

//go:generate mockery -inpkg -testonly -case underscore -name extendableMutex

type extendableMutex interface {
	Lock() error
	Unlock() error
	Extend() error
}

func onExtendErrorOption(f func(error)) func(*mutexWithCancelOnLockLost) {
	return func(mtx *mutexWithCancelOnLockLost) {
		mtx.onExtendError = f
	}
}
