package grit

import (
	"context"
	"math/rand/v2"
	"time"
)

type RetryOpt func(*retryConfig)

type retryConfig struct {
	backoff BackoffStrategy
	jitter  float64
	retryIf func(error) bool
}

type BackoffStrategy func(attempt int) time.Duration

func Constant(d time.Duration) BackoffStrategy {
	return func(_ int) time.Duration { return d }
}

func Exponential(base time.Duration) BackoffStrategy {
	return func(i int) time.Duration {
		return time.Duration(1<<i) * base
	}
}

func WithBackoff(strategy BackoffStrategy) RetryOpt {
	return func(c *retryConfig) { c.backoff = strategy }
}

func WithExponentialBackoff(base time.Duration) RetryOpt {
	return func(c *retryConfig) { c.backoff = Exponential(base) }
}

func WithJitter(factor float64) RetryOpt {
	return func(c *retryConfig) { c.jitter = factor }
}

func WithRetryIf(fn func(error) bool) RetryOpt {
	return func(c *retryConfig) { c.retryIf = fn }
}

func Retry(attempts int, opts ...RetryOpt) Guard {
	cfg := retryConfig{
		backoff: Constant(0),
		jitter:  0,
		retryIf: func(err error) bool { return true },
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return GuardFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		var lastErr error
		for i := range attempts {
			res, err := fn(ctx)
			if err == nil {
				return res, nil
			}
			if !cfg.retryIf(err) {
				return nil, err
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
