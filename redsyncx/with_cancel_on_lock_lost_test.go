package redsyncx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.otters.xyz/product/journey/timex.git"
	"gitlab.otters.xyz/product/journey/timex.git/timextest"
	"gitlab.otters.xyz/product/match/library/gox.git/chantest"
)

func TestUnlockCancelsContext(t *testing.T) {
	timexmock := timextest.Mock(time.Now())
	defer timexmock.RestoreDefaultImplementation()

	const ttl = 10 * time.Minute

	var lock mockExtendableMutex
	defer lock.AssertExpectations(t)

	lock.On("Lock").Return(nil).Once()
	lock.On("Unlock").Return(nil).Once()

	mtx := withCancelOnLockLost(ttl, &lock, func(f func()) { go f() })
	ctx, err := mtx.Lock(context.Background())
	assert.NoError(t, err)

	chantest.AssertNoRecv(t, ctx.Done(), "context is cancelled before calling unlock")

	<-timexmock.AfterCalls
	mtx.Unlock()

	chantest.AssertRecv(t, ctx.Done(), "context isn't cancelled when unlock is called")
	chantest.AssertNoRecv(t, timexmock.AfterCalls, "cancelling context didn't kill context goroutine")
}

func TestParentCancelledCancelsContext(t *testing.T) {
	timexmock := timextest.Mock(time.Now())
	defer timexmock.RestoreDefaultImplementation()

	const ttl = 10 * time.Minute

	var lock mockExtendableMutex
	defer lock.AssertExpectations(t)

	lock.On("Lock").Return(nil).Once()

	parent, cancel := context.WithCancel(context.Background())

	mtx := withCancelOnLockLost(ttl, &lock, func(f func()) { go f() })
	ctx, err := mtx.Lock(parent)
	assert.NoError(t, err)

	chantest.AssertNoRecv(t, ctx.Done(), "context is cancelled before cancelling parent")

	<-timexmock.AfterCalls
	cancel()

	chantest.AssertRecv(t, ctx.Done(), "context isn't cancelled when parent is cancelled")
	chantest.AssertNoRecv(t, timexmock.AfterCalls, "cancelling context didn't kill context goroutine")
}

func TestFailedExtendCancelsContext(t *testing.T) {
	timexmock := timextest.Mock(time.Now())
	defer timexmock.RestoreDefaultImplementation()

	const ttl = 10 * time.Minute

	var lock mockExtendableMutex
	defer lock.AssertExpectations(t)

	lock.On("Lock").Return(nil).Once()

	mtx := withCancelOnLockLost(ttl, &lock, func(f func()) { go f() })
	ctx, err := mtx.Lock(context.Background())
	assert.NoError(t, err)

	chantest.AssertNoRecv(t, ctx.Done(), "context is cancelled before failing to extend")

	// Let's ensure first that a successful extend doesn't cancel anything.

	after := <-timexmock.AfterCalls
	assert.Equal(t, after.Duration, ttl/2)

	lock.On("Extend").Return(nil).Once()
	after.Done(timex.Now())

	chantest.AssertNoRecv(t, ctx.Done(), "context is cancelled even though extending the lock succeeded")

	// A subsequent failed extend should cancel; we've lost the lock.

	after = <-timexmock.AfterCalls
	assert.Equal(t, after.Duration, ttl/2)

	lock.On("Extend").Return(errors.New("nope")).Once()
	after.Done(timex.Now())

	chantest.AssertRecv(t, ctx.Done(), "context isn't cancelled when extending the lock fails")
	chantest.AssertNoRecv(t, timexmock.AfterCalls, "cancelling context didn't kill context goroutine")
}

func TestOnExtendError(t *testing.T) {
	timexmock := timextest.Mock(time.Now())
	defer timexmock.RestoreDefaultImplementation()

	const ttl = 10 * time.Minute

	var lock mockExtendableMutex
	defer lock.AssertExpectations(t)

	expectedErr := errors.New("bad extend")

	lock.On("Lock").Return(nil).Once()
	lock.On("Extend").Return(expectedErr).Once()

	var gotErr error

	mtx := withCancelOnLockLost(ttl, &lock, func(f func()) { go f() }, onExtendErrorOption(func(err error) {
		gotErr = err
	}))

	ctx, err := mtx.Lock(context.Background())
	assert.NoError(t, err)

	after := <-timexmock.AfterCalls

	after.Done(timex.Now())

	chantest.AssertRecv(t, ctx.Done(), "context isn't cancelled when extending the lock fails")
	chantest.AssertNoRecv(t, timexmock.AfterCalls, "cancelling context didn't kill context goroutine")

	assert.Equal(t, expectedErr, gotErr)
}
