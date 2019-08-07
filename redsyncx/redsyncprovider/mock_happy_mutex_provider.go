package redsyncprovider

import (
	"context"

	"github.com/cabify/redsync/v2/redsyncx"
)

// MockHappyProvider is a mock for Provider that returs mutexes that always
// succeed, without actually locking anything
type MockHappyProvider struct {
}

func (MockHappyProvider) For(ctx context.Context, id string) redsyncx.Mutex {
	return &MockHappyMutex{
		id: id,
	}
}

// MockHappyMutex is a mock for a mutex that doesn't mutex anything, it just succeeds
type MockHappyMutex struct {
	id     string
	cancel func()
}

func (m *MockHappyMutex) Lock(ctx context.Context) (context.Context, error) {
	ctx, m.cancel = context.WithCancel(ctx)
	return ctx, nil
}

func (m *MockHappyMutex) Unlock() error {
	m.cancel()
	return nil
}
