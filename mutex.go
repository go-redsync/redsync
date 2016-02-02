package redsync

import (
	"sync"
	"time"
)

type Mutex struct {
	name   string
	expiry time.Duration

	tries int
	delay time.Duration

	factor float64

	quorum int

	value string
	until time.Time

	nodem sync.Mutex

	pools []Pool
}

func (m *Mutex) Lock() error {
	panic("unimplemented")
}

func (m *Mutex) Unlock() {
	panic("unimplemented")
}

func (m *Mutex) touch() bool {
	return false
}
