# Bamboo — Resilience Patterns for Go

**Bamboo** is a lightweight, composable Go library for building resilient systems.
It provides primitives like **Retry**, **Circuit Breaker**, **Timeout**, and more. It is designed to be clean, idiomatic, and dependency-free.

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
go get github.com/gustus-co/bamboo@latest
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
	return someOperation(ctx)
})
```

Bamboo provides several built-in policies:
|Policy|Description|
|---|---|
|Retry|Retries a failing operation using backoff strategies|
|CircuitBreaker|Halts requests after consecutive failures|
|Timeout|Cancels an operation after a fixed duration|
|Limiter|Restricts concurrent executions|
|Fallback|Provides alternate logic when an operation fails|


### **Chain**
Policies can be composed declaratively using the `Chain` function:
```go
resilience := bamboo.Chain(
	bamboo.Timeout(3*time.Second),
	bamboo.Retry(3, bamboo.WithBackoff(bamboo.Exponential(100*time.Millisecond))),
	bamboo.CircuitBreaker(3),
)
```
Execution order:
```
Timeout → Retry → CircuitBreaker → your function
```
