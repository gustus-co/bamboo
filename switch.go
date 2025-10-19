package grit

import "context"

func Switch(cond func(ctx context.Context) bool, a Guard, b Guard) Guard {
    return GuardFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
        if cond(ctx) {
            return a.Do(ctx, fn)
        }
        return b.Do(ctx, fn)
    })
}
