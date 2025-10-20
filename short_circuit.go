package bamboo

import "context"

// ShortCircuitIf prevents outer policies from acting when the provided predicate returns true for the encountered error.
// It short-circuits the chain and returns the error immediately.
func ShortCircuitIf(predicate func(error) bool) Policy {
    return PolicyFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
        res, err := fn(ctx)
        if err != nil && predicate(err) {
            return nil, err
        }
        return res, err
    })
}
