package granit

import "context"

type Policy interface {
	Do(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error)
}

type PolicyFunc func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error)

func (f PolicyFunc) Do(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	return f(ctx, fn)
}

// Chain chains multiple policies together into a single one.
// The first argument becomes the outermost and the last argument
// becomes the innermost. For example, Chain(A, B, C) produces a policy
// that executes in the order: A -> B -> C -> fn
func Chain(policies ...Policy) Policy {
	return PolicyFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		wrapped := fn
		for i := len(policies) - 1; i >= 0; i-- {
			p := policies[i]
			inner := wrapped
			wrapped = func(ctx context.Context) (any, error) {
				return p.Do(ctx, inner)
			}
		}
		return wrapped(ctx)
	})
}
