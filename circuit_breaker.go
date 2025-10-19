package grit

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

// --- Internal state ---

type breakerState int

const (
	Closed breakerState = iota
	Opened
	HalfOpened
)

type CircuitBreakerOpt func(*circuitBreakerConfig)

type circuitBreakerConfig struct {
	failures     uint
	openDuration time.Duration
	maxRequests  uint
	resetInterval     time.Duration
}

func defaultCircuitBreakerConfig() circuitBreakerConfig {
	return circuitBreakerConfig{
		failures:     5,
		openDuration: 30 * time.Second,
		maxRequests:  1,
		resetInterval:     0,
	}
}

func WithFailures(n uint) CircuitBreakerOpt {
	return func(c *circuitBreakerConfig) {
		if n > 0 {
			c.failures = n
		}
	}
}

func WithOpenDuration(d time.Duration) CircuitBreakerOpt {
	return func(c *circuitBreakerConfig) {
		if d > 0 {
			c.openDuration = d
		}
	}
}

func WithMaxRequests(n uint) CircuitBreakerOpt {
	return func(c *circuitBreakerConfig) {
		if n > 0 {
			c.maxRequests = n
		}
	}
}

func WithResetInterval(d time.Duration) CircuitBreakerOpt {
	return func(c *circuitBreakerConfig) {
		c.resetInterval = d
	}
}

func CircuitBreaker(opts ...CircuitBreakerOpt) Guard {
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

	return GuardFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		mu.Lock()
		now := time.Now()

		// reset failure counters periodically (if configured)
		if state == Closed && cfg.resetInterval > 0 && now.Sub(lastReset) >= cfg.resetInterval {
			failures = 0
			lastReset = now
		}

		// move from Open → HalfOpen when timeout expires
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
			if failures >= cfg.failures {
				state = Opened
				nextAttempt = now.Add(cfg.openDuration)
				return v, ErrCircuitBreakerTripped
			}
		}

		return v, err
	})
}
