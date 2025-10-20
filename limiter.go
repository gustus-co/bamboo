package bamboo

import "context"

// Limiter restricts the number of concurrent executions
// of the provided operation.
//
// If the context is canceled before a slot becomes available, the operation
// is skipped and the Policy returns ctx.Err().
//
// Limiter helps isolate overloads and prevent cascading failures by bounding concurrency.
func Limiter(limit int) Policy {
	sem := make(chan struct{}, limit)
	return PolicyFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			return fn(ctx)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
}
