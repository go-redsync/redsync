// Code generated by mockery v1.0.0. DO NOT EDIT.

package redsyncprovidertest

import context "context"
import mock "github.com/stretchr/testify/mock"

import redsyncx "github.com/cabify/redsync/v2/redsyncx"

// Provider is an autogenerated mock type for the Provider type
type Provider struct {
	mock.Mock
}

// For provides a mock function with given fields: ctx, key
func (_m *Provider) For(ctx context.Context, key string) redsyncx.Mutex {
	ret := _m.Called(ctx, key)

	var r0 redsyncx.Mutex
	if rf, ok := ret.Get(0).(func(context.Context, string) redsyncx.Mutex); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(redsyncx.Mutex)
		}
	}

	return r0
}
