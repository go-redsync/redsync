package redsync

import "errors"

// ErrFailed is the error resulting if Redsync fails to acquire the lock after
// exhausting all retries.
var ErrFailed = errors.New("redsync: failed to acquire lock")

// ErrCtxDone is the error resulting if the context is done
var ErrCtxDone = errors.New("redsync: ctx is done")

// ErrExtendFailed is the error resulting if Redsync fails to extend the
// lock.
var ErrExtendFailed = errors.New("redsync: failed to extend lock")
