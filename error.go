package redsync

import (
	"errors"
	"fmt"
)

// ErrFailed is the error resulting if Redsync fails to acquire the lock after
// exhausting all retries.
var ErrFailed = errors.New("redsync: failed to acquire lock")

// ErrExtendFailed is the error resulting if Redsync fails to extend the
// lock.
var ErrExtendFailed = errors.New("redsync: failed to extend lock")

// ErrTaken happens when the lock is already taken in a quorum on nodes.
type ErrTaken struct {
	Nodes []int
}

func (err ErrTaken) Error() string {
	return fmt.Sprintf("lock already taken, locked nodes: %v", err.Nodes)
}

// ErrNodeTaken is the error resulting if the lock is already taken in one of
// the cluster's nodes
type ErrNodeTaken struct {
	Node int
}

func (err ErrNodeTaken) Error() string {
	return fmt.Sprintf("node #%d: lock already taken", err.Node)
}

// A RedisError is an error communicating with one of the Redis nodes.
type RedisError struct {
	Node int
	Err  error
}

func (err RedisError) Error() string {
	return fmt.Sprintf("node #%d: %s", err.Node, err.Err)
}
