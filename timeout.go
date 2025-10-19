package granit

import (
	"context"
	"time"
)

// Timeout returns a GuardFunc that limits how long the operation
// is allowed to run. If the given duration elapses, the operation’s
// context is canceled and fn is interrupted.
//
// Timeout enforces an upper bound on execution time. Unlike Retry
// or CircuitBreaker, it does not analyze errors or state, it simply
// protects against hanging or slow operations. It pairs well with
// Retry to bound total execution time across multiple attempts.
func Timeout(d time.Duration) GuardFunc {
	return func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, d)
		defer cancel()
		return fn(ctx)
	}
}
