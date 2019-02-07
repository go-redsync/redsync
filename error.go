package redsync

import (
	"errors"
	"fmt"
	"strings"
)

const takenMsg = "lock already taken"

// ErrTaken happens when the lock is already taken in a quorum on nodes.
var ErrTaken = errors.New(takenMsg)

// A NodeTaken happens when the lock is already taken in one of the cluster's
// nodes.
type NodeTaken struct {
	Node int
}

func (err NodeTaken) Error() string {
	return fmt.Sprintf("node #%d: "+takenMsg, err.Node)
}

// A RedisError is an error communicating with one of the Redis nodes.
type RedisError struct {
	Node int
	Err  error
}

func (err RedisError) Error() string {
	return fmt.Sprintf("node #%d: %s", err.Node, err.Err)
}

// NoQuorum is a series of errors, either NodeTaken or RedisError, that happened
// while trying to acquire or extend a lock in a cluster and prevent the lock to
// be taken in a quorum of nodes.
type NoQuorum []error

func (errs NoQuorum) Error() string {
	msgs := make([]string, 0, len(errs))
	for _, err := range errs {
		msgs = append(msgs, err.Error())
	}
	return "redsync: " + strings.Join(msgs, "; ")
}
