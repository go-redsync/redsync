package redsync

import "errors"

var ErrFailed = errors.New("redsync: failed to acquire lock")
var ErrExtendFailed = errors.New("redsync: failed to extend lock")
