package granit

import (
	"context"
	"time"
)

func Timeout(d time.Duration) GuardFunc {
	return func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		ctx, cancel := context.WithTimeout(ctx, d)
		defer cancel()
		return fn(ctx)
	}
}
