package bamboo

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrCircuitBreakerOpen    = errors.New("circuit breaker is open")
	ErrCircuitBreakerTooMany = errors.New("circuit breaker half-open: too many concurrent requests")
	ErrCircuitBreakerTripped = errors.New("circuit breaker tripped after consecutive failures")
)

type breakerState int

const (
	Closed breakerState = iota
	Opened
	HalfOpened
)

type CircuitBreakerOpt func(*circuitBreakerConfig)

type circuitBreakerConfig struct {
	openDuration  time.Duration
	maxRequests   uint
	resetInterval time.Duration
}

func defaultCircuitBreakerConfig() circuitBreakerConfig {
	return circuitBreakerConfig{
		openDuration:  30 * time.Second,
		maxRequests:   1,
		resetInterval: 0,
	}
}

// WithOpenDuration sets how long the circuit remains open
// after tripping before it transitions to half-open state
// to test recovery.
func WithOpenDuration(d time.Duration) CircuitBreakerOpt {
	return func(c *circuitBreakerConfig) {
		if d > 0 {
			c.openDuration = d
		}
	}
}

// WithMaxRequests defines how many test requests are allowed
// concurrently while the circuit is half-open. If any of them
// fail, the circuit reopens immediately.
func WithMaxRequests(n uint) CircuitBreakerOpt {
	return func(c *circuitBreakerConfig) {
		if n > 0 {
			c.maxRequests = n
		}
	}
}

// WithResetInterval resets the circuit breaker’s internal failure count after
// the given duration if the circuit is closed. This helps
// recover from old failures and prevents the circuit from
// tripping due to unrelated errors.
func WithResetInterval(d time.Duration) CircuitBreakerOpt {
	return func(c *circuitBreakerConfig) {
		c.resetInterval = d
	}
}

// CircuitBreaker monitors consecutive
// operation failures and temporarily halts new attempts when the
// failure threshold is exceeded. Once opened, it stays open for
// the configured duration before allowing limited test requests.
//
// CircuitBreaker protects systems from cascading failures and
// excessive retry storms. It differs from Retry in that it stops
// execution entirely when failures persist rather than retrying.
func CircuitBreaker(consecutiveFailures uint, opts ...CircuitBreakerOpt) Policy {
	cfg := defaultCircuitBreakerConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	var (
		mu              sync.Mutex
		state           = Closed
		nextAttempt     = time.Now()
		lastReset       = time.Now()
		failures        uint
		halfOpenRunning uint
	)

	return PolicyFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		mu.Lock()
		now := time.Now()

		// reset failure counters periodically (if configured)
		if state == Closed && cfg.resetInterval > 0 && now.Sub(lastReset) >= cfg.resetInterval {
			failures = 0
			lastReset = now
		}

		// move from Open to HalfOpen when timeout expires
		if state == Opened && now.After(nextAttempt) {
			state = HalfOpened
			halfOpenRunning = 0
		}

		// check if breaker allows the request
		switch state {
		case Opened:
			mu.Unlock()
			return nil, ErrCircuitBreakerOpen

		case HalfOpened:
			if halfOpenRunning >= cfg.maxRequests {
				mu.Unlock()
				return nil, ErrCircuitBreakerTooMany
			}
			halfOpenRunning++
		}
		mu.Unlock()

		// execute protected function
		v, err := fn(ctx)

		mu.Lock()
		defer mu.Unlock()

		if err == nil {
			// success: closed
			failures = 0
			state = Closed
			halfOpenRunning = 0
			return v, nil
		}

		// failure handling
		failures++
		switch state {
		case HalfOpened:
			// failure in half-opened: opened
			state = Opened
			nextAttempt = now.Add(cfg.openDuration)
			halfOpenRunning = 0

		case Closed:
			if failures >= consecutiveFailures {
				state = Opened
				nextAttempt = now.Add(cfg.openDuration)
				return v, ErrCircuitBreakerTripped
			}
		}

		return v, err
	})
}
