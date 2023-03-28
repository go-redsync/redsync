package redsync

import "sync"

var locksMap sync.Map

type MixinMutex struct {
	crossProcess *Mutex
	inProcess    *sync.Mutex
}

// at any moment,only one goroutine can run Do func with distributed lock
func (m *MixinMutex) Do(action func(l *Mutex) error) error {
	m.inProcess.Lock()
	defer m.inProcess.Unlock()
	return action(m.crossProcess)
}
