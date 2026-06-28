# Retry with Exponential Backoff in Go

## What is it?

When a network call fails, instead of giving up immediately or hammering the server repeatedly, you **retry with increasing delays** between each attempt.

```
Attempt 1 fails → wait 1s
Attempt 2 fails → wait 2s
Attempt 3 fails → wait 4s
Attempt 4 fails → wait 8s
Attempt 5 fails → give up
```

The delay **doubles** each time — this is the "exponential" part.

---

## Why Not Just Retry Immediately?

```
BAD:  retry every 10ms  → floods the server (makes it worse)
GOOD: retry with backoff → gives server time to recover
```

---

## The RetryConfig Struct

```go
type RetryConfig struct {
    MaxRetries int           // How many times to try total
    BaseDelay  time.Duration // Starting delay (e.g. 1s)
    MaxDelay   time.Duration // Cap the delay (e.g. 30s)
    Multiplier float64       // How fast delay grows (2.0 = doubles)
    Jitter     bool          // Add randomness to delay
}
```

### Default values:

| Field | Value | Meaning |
|-------|-------|---------|
| MaxRetries | 5 | Try up to 5 times |
| BaseDelay | 1s | Start with 1 second wait |
| MaxDelay | 30s | Never wait more than 30s |
| Multiplier | 2.0 | Double the delay each retry |
| Jitter | true | Add randomness |

---

## How the Delay Grows

With `BaseDelay = 1s` and `Multiplier = 2.0`:

```
Attempt 1 fails → wait 1s
Attempt 2 fails → wait 2s   (1s × 2)
Attempt 3 fails → wait 4s   (2s × 2)
Attempt 4 fails → wait 8s   (4s × 2)
Attempt 5 fails → give up
```

Code that does this:

```go
delay = time.Duration(float64(delay) * cfg.Multiplier)
```

---

## What is Jitter?

Without jitter, if 100 clients all fail at the same time, they all retry at exactly the same moment — called the **thundering herd problem**.

```
No Jitter:   100 clients retry at T=1s, T=2s, T=4s (all together) ← BAD
With Jitter: 100 clients retry at T=1.2s, T=1.7s, T=2.3s (spread out) ← GOOD
```

Code that adds jitter:

```go
jitter := time.Duration(rand.Int63n(int64(delay) / 2))
wait = delay + jitter  // adds a random extra delay
```

---

## Context Cancellation

The `Retry` function accepts a `context.Context`. If the context times out or is cancelled, retrying stops immediately:

```go
select {
case <-ctx.Done():
    return fmt.Errorf("retry cancelled: %w", ctx.Err())
case <-time.After(wait):
    // Continue waiting
}
```

This is important in production — you don't want retries running forever if the caller has already moved on.

---

## Three Scenarios

### 1. Unstable API (succeeds eventually)

```
Attempt 1: FAILED — connection timeout
Waiting 507ms...
Attempt 2: FAILED — connection timeout
Waiting 1.3s...
Attempt 3: SUCCESS
```

### 2. Always failing (max retries exhausted)

```
Attempt 1: FAILED
Waiting 253ms...
Attempt 2: FAILED
Waiting 414ms...
Attempt 3: FAILED
Final Error: all 3 attempts failed: server unavailable
```

### 3. Context timeout cancels early

```
Attempt 1: FAILED
Waiting 1.3s... ← Context times out during this wait
Final Error: retry cancelled: context deadline exceeded
```

---

## Real World Use Cases

| Scenario | Why exponential backoff? |
|----------|--------------------------|
| HTTP API call fails | Server might be temporarily overloaded |
| Database connection fails | DB might be restarting |
| Message queue unavailable | Broker might be restarting |
| File upload fails | Network blip, retry makes sense |
| Kubernetes readiness probe | Pod not ready yet, keep trying |

---

## Key Takeaways

- **Exponential** — delay grows by multiplier each attempt (1s → 2s → 4s → 8s)
- **Backoff** — you wait before retrying, not hammering immediately
- **Jitter** — adds randomness to spread out retries across many clients
- **MaxDelay** — caps the wait so it doesn't grow forever
- **Context** — cancels retry if caller gives up or times out
