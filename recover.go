package grit

import (
	"context"
)

func Recover() Guard {
	return GuardFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		defer func() {
			recover()
		}()
		return fn(ctx)
	})
}
