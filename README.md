# Bamboo

[![Go Reference](https://pkg.go.dev/badge/github.com/gustus-co/bamboo.svg)](https://pkg.go.dev/github.com/gustus-co/bamboo)
[![Go Report Card](https://goreportcard.com/badge/github.com/gustus-co/bamboo)](https://goreportcard.com/report/github.com/gustus-co/bamboo)
[![Go Coverage](https://coder.github.io/websocket/coverage.svg)](https://coder.github.io/websocket/coverage.html)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Bamboo** is a lightweight, composable Go library for building resilient systems.
It provides primitives like **Retry**, **Circuit Breaker**, **Timeout**, and more. It is designed to be idiomatic, and dependency-free.

---

## Features

- **Retry** with exponential backoff and jitter
- **Circuit Breaker** with half-open recovery
- **Timeout** guards for bounded latency
- **Limiter** for concurrency control
- **Composable** with `Chain()` for custom resilience stacks
- Extensible with `PolicyFunc` — create your own policies

---

## Installation

```bash
go get github.com/gustus-co/bamboo
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gustus-co/bamboo"
)

func main() {
	ctx := context.Background()

	resilience := bamboo.Chain(
		bamboo.Timeout(5*time.Second),
		bamboo.CircuitBreaker(3,
			bamboo.WithOpenDuration(10*time.Second),
			bamboo.WithTripWhen(isSystemError),
		),
		bamboo.Retry(3,
			bamboo.WithBackoff(bamboo.Exponential(100*time.Millisecond)),
			bamboo.WithJitter(0.2),
			bamboo.WithRetryUntil(isPermanentError),
		),
	)

	result, err := resilience.Do(ctx, func(ctx context.Context) (any, error) {
		return callRemoteAPI(ctx)
	})
	if err != nil {
		fmt.Println("failed:", err)
		return
	}

	fmt.Println("success:", result)
}
```
 
## Core Concepts

### **Policy**

A **Policy** defines a reusable resilience behavior.  
Every operation is executed through:

```go
result, err := policy.Do(ctx, func(ctx context.Context) (any, error) {
	return "foo", nil
})
res := result.(string)
```

Bamboo provides several built-in policies:
| Policy | Description |
|---|---|
| `Retry` |Retries a failing operation using backoff strategies|
| `CircuitBreaker` |Halts requests after consecutive failures|
| `Timeout` |Cancels an operation after a fixed duration|
| `Limiter` |Restricts concurrent executions|
| `Fallback` |Provides alternate logic when an operation fails|


### **Chain**
Policies can be composed using the `Chain` function:
```go
resilience := bamboo.Chain(
	bamboo.Timeout(3*time.Second),
	bamboo.Retry(3, bamboo.WithBackoff(bamboo.Exponential(100*time.Millisecond))),
	bamboo.CircuitBreaker(3),
)
result, err := resilience.Do(ctx, func(ctx context.Context) (any, error) {
	time.Sleep(200 * time.Millisecond)
	return "bar", nil
})
```
Execution order:
```go
Timeout(Retry(CircuitBreaker(yourFunction)))
```

### **Custom Policies**
You can easily define your own `Policy` using `PolicyFunc`:

```go
func LogPolicy() bamboo.Policy {
	return bamboo.PolicyFunc(func(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
		start := time.Now()
		defer func() { fmt.Println("operation took:", time.Since(start)) }()
		return fn(ctx)
	})
}
```

## License

Bamboo is licensed under the [MIT License](LICENSE).

## Contributing

Issues and PRs are welcome! Bamboo aims to stay small and dependency-free.
if you have a useful pattern or improvement, please open a discussion before submitting a PR.
