package bamboo

import (
	"context"
)

func Recover() Policy {
	return PolicyFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		defer func() {
			recover()
		}()
		return fn(ctx)
	})
}
