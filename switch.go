package granit

import "context"

// Switch chooses between two other policies
// based on the given condition function. If cond(ctx) evaluates
// to true, Policy a is used. Otherwise, Policy b is.
//
// Switch is useful for runtime selection of resilience behavior,
// such as applying different retry strategies based on tenant,
// feature flag, or execution context. Unlike Fallback, which reacts
// after an error, Switch decides before the function runs.
func Switch(cond func(ctx context.Context) bool, a Policy, b Policy) Policy {
	return PolicyFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		if cond(ctx) {
			return a.Do(ctx, fn)
		}
		return b.Do(ctx, fn)
	})
}
