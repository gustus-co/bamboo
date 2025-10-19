package granit

import (
	"context"
)

// Fallback returns a Guard that invokes the provided handler function
// when the wrapped operation returns an error. It allows graceful
// degradation by producing an alternate value or error.
func Fallback(handler func(ctx context.Context, err error) (any, error)) Guard {
	return GuardFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		v, err := fn(ctx)
		if err != nil {
			return handler(ctx, err)
		}
		return v, nil
	})
}
