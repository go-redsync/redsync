package redsync

import "time"

type Redsync struct {
}

func New(pools []Pool) *Redsync {
	panic("unimplemented")
}

func (r *Redsync) NewMutex(name string, options ...Option) *Mutex {
	panic("unimplemented")
}

type Option interface {
	Apply(*Mutex)
}

type OptionFunc func(*Mutex)

func (f OptionFunc) Apply(mutex *Mutex) {
	f(mutex)
}

func (f OptionFunc) SetExpiry(expiry time.Duration) Option {
	return OptionFunc(func(m *Mutex) {
		m.expiry = expiry
	})
}

func (f OptionFunc) SetTries(tries int) Option {
	return OptionFunc(func(m *Mutex) {
		m.tries = tries
	})
}

func (f OptionFunc) SetDelay(delay time.Duration) Option {
	return OptionFunc(func(m *Mutex) {
		m.delay = delay
	})
}

func (f OptionFunc) SetFactor(factor float64) Option {
	return OptionFunc(func(m *Mutex) {
		m.factor = factor
	})
}
