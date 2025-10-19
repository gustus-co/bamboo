package granit

import "context"

type Guard interface {
	Do(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error)
}

type GuardFunc func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error)

func (gf GuardFunc) Do(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
    return gf(ctx, fn)
}
