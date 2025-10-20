package granit

import (
	"context"
)

// Fallback invokes the provided handler function
// when the wrapped operation returns an error. It allows graceful
// degradation by producing an alternate value or error.
func Fallback(fallback func(ctx context.Context, err error) (any, error)) PolicyFunc {
	return PolicyFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		v, err := fn(ctx)
		if err != nil {
			return fallback(ctx, err)
		}
		return v, nil
	})
}
