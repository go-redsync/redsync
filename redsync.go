package redsync

import "time"

type Redsync struct {
	pools []Pool
}

func New(pools []Pool) *Redsync {
	return &Redsync{
		pools: pools,
	}
}

func (r *Redsync) NewMutex(name string, options ...Option) *Mutex {
	m := &Mutex{
		name:   name,
		expiry: 8 * time.Second,
		tries:  32,
		delay:  500 * time.Millisecond,
		factor: 0.01,
		quorum: len(r.pools)/2 + 1,
		pools:  r.pools,
	}
	for _, o := range options {
		o.Apply(m)
	}
	return m
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
