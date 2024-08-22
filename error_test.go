package redsync

import (
	"errors"
	"fmt"
	"testing"
)

func TestRedisErrorIs(t *testing.T) {
	cases := map[string]error{
		"defined_error":  ErrFailed,
		"other_error":    errors.New("other error"),
		"wrapping_error": fmt.Errorf("wrapping: %w", ErrFailed),
	}

	for k, v := range cases {

		t.Run(k, func(t *testing.T) {
			err := &RedisError{
				Node: 0,
				Err:  v,
			}

			if !errors.Is(err, v) {
				t.Errorf("errors.Is must be true")
			}
		})
	}
}

func TestRedisErrorAs(t *testing.T) {
	derr := dummyError{}

	cases := map[string]error{
		"simple_error":   derr,
		"wrapping_error": fmt.Errorf("wrapping: %w", derr),
	}

	for k, v := range cases {

		t.Run(k, func(t *testing.T) {
			err := &RedisError{
				Node: 0,
				Err:  v,
			}

			var target dummyError
			if !errors.As(err, &target) {
				t.Errorf("errors.As must be true")
			}

			if target != derr {
				t.Errorf("Expected target == %q, got %q", derr, target)
			}
		})
	}
}

type dummyError struct{}

func (err dummyError) Error() string {
	return "dummy"
}
