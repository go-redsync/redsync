package redsync

import "errors"

// ErrFailed is the error resulting if Redsync fails to acquire the lock afer
// exhausting all retries.
var ErrFailed = errors.New("redsync: failed to acquire lock")
