package granit

import "context"

type Guard interface {
	Do(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error)
}

type GuardFunc func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error)

func (gf GuardFunc) Do(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	return gf(ctx, fn)
}

// Chain wraps multiple guards together into a single one.
// The first argument becomes the outermost guard, and the last argument
// becomes the innermost. For example, Chain(A, B, C) produces a policy
// that executes in the order: A -> B -> C
func Chain(guards ...Guard) Guard {
	return GuardFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		wrapped := fn
		for i := len(guards) - 1; i >= 0; i-- {
			p := guards[i]
			inner := wrapped
			wrapped = func(ctx context.Context) (any, error) {
				return p.Do(ctx, inner)
			}
		}
		return wrapped(ctx)
	})
}
