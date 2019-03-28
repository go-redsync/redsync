package redsync

import "errors"

var ErrFailed = errors.New("redsync: failed to acquire lock")
var ErrLockAlreadyAcquired = errors.New("redsync: lock is already acquired")
