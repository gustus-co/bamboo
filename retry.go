package granit

import (
	"context"
	"math/rand/v2"
	"time"
)

type RetryOpt func(*retryConfig)

type retryConfig struct {
	backoff BackoffStrategy
	jitter  float64
}

// BackoffStrategy defines how much time to wait before each retry
// attempt. The function receives the attempt number (0-based) and
// returns a delay duration.
type BackoffStrategy func(attempt int) time.Duration

// Constant returns a BackoffStrategy that always waits for the
// same duration between retries. Useful for stable but gentle retry
// intervals or in tests.
func Constant(d time.Duration) BackoffStrategy {
	return func(_ int) time.Duration { return d }
}

// Exponential returns a BackoffStrategy that doubles the delay
// with each attempt (1<<i * base). It is the preferred strategy
// for transient failures where rapid successive retries could
// overload a downstream service.
func Exponential(base time.Duration) BackoffStrategy {
	return func(i int) time.Duration {
		return time.Duration(1<<i) * base
	}
}

// WithBackoff specifies a BackoffStrategy used to delay between
// retries. It can be combined with WithJitter to randomize the
// delay slightly, reducing contention when many operations retry
// simultaneously.
func WithBackoff(strategy BackoffStrategy) RetryOpt {
	return func(c *retryConfig) { c.backoff = strategy }
}

// WithJitter randomizes the backoff delay by up to the given
// factor (0.3 = ±30%). Adding jitter prevents retry storms and
// spreads load when multiple clients retry at the same time.
func WithJitter(factor float64) RetryOpt {
	return func(c *retryConfig) { c.jitter = factor }
}

// Retry re-executes the operation up to the
// specified number of attempts when an error occurs.
//
// Retry is the most common Policy for transient failures. It should
// be used for operations that are likely to succeed on a subsequent
// attempt (e.g., network or I/O calls). For permanent errors,
// combine it with ShortCircuitIf to avoid useless retries.
func Retry(attempts int, opts ...RetryOpt) Policy {
	cfg := retryConfig{
		backoff: Constant(0),
		jitter:  0,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return PolicyFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		var lastErr error
		for i := range attempts {
			res, err := fn(ctx)
			if err == nil {
				return res, nil
			}

			lastErr = err
			delay := cfg.backoff(i)
			if cfg.jitter > 0 {
				jitter := rand.Float64()*cfg.jitter*float64(delay) - (cfg.jitter/2)*float64(delay)
				delay += time.Duration(jitter)
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
		return nil, lastErr
	})
}
