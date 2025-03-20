package redsync

import (
	"context"
	"fmt"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func TestMutex_MultipleLockContext(t *testing.T) {
	ctx := context.Background()
	keys := []string{"1", "2", "3"}

	for k, v := range makeCases(8) {
		t.Run(k, func(t *testing.T) {
			mutexes := newTestMutexes(v.pools, k+"-test-mutex-multiple", v.poolCount)
			for _, v := range mutexes {
				v.keysForMultipleLock = keys
				v.expiry = time.Millisecond * 100
			}

			eg := errgroup.Group{}

			for i, mutex := range mutexes {
				func(i int, mutex *Mutex) {
					eg.Go(func() error {
						err := mutex.MultipleLockContext(nil)
						if err != nil {
							return err
						}
						defer mutex.UnlockMultipleContext(nil)

						if !isAcquired(ctx, v.pools, mutex) {
							return fmt.Errorf("Expected n >= %d, got %d", mutex.quorum, countAcquiredPools(ctx, v.pools, mutex))
						}
						return nil
					})
				}(i, mutex)
			}
			err := eg.Wait()
			if err != nil {
				t.Fatalf("mutex lock failed: %s", err)
			}
		})
	}
}
