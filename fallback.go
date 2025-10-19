package granit

import (
	"context"
)

func Fallback(handler func(ctx context.Context, err error) (any, error)) Guard {
	return GuardFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		v, err := fn(ctx)
		if err != nil {
			return handler(ctx, err)
		}
		return v, nil
	})
}
