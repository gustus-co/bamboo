package granit

import "context"

// Switch returns a Guard that chooses between two other Guards
// based on the given condition function. If cond(ctx) evaluates
// to true, Guard a is used. Otherwise, Guard b is.
//
// Switch is useful for runtime selection of resilience behavior,
// such as applying different retry strategies based on tenant,
// feature flag, or execution context. Unlike Fallback, which reacts
// after an error, Switch decides before the guarded function runs.
func Switch(cond func(ctx context.Context) bool, a Guard, b Guard) Guard {
	return GuardFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		if cond(ctx) {
			return a.Do(ctx, fn)
		}
		return b.Do(ctx, fn)
	})
}
