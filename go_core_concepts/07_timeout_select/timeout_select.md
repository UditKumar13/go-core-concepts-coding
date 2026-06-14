# Timeout & Select Pattern in Go

## What is Select?

`select` is like a `switch` but for channels. It waits on multiple channel operations simultaneously and executes whichever one is ready first.

```go
select {
case val := <-ch1:
    // ch1 had a value
case val := <-ch2:
    // ch2 had a value
case ch3 <- x:
    // successfully sent x into ch3
default:
    // none of the above were ready — runs immediately (non-blocking)
}
```

If multiple cases are ready at the same time, Go picks one **at random** (not first-in-list).

---

## What is a Timeout?

A timeout says: "wait for this operation, but give up after N duration".

Without timeout — goroutine blocks forever if no response:
```go
result := <-slowOperation()  // what if it never returns?
```

With timeout — give up after 2 seconds:
```go
select {
case result := <-slowOperation():
    fmt.Println("got result:", result)
case <-time.After(2 * time.Second):
    fmt.Println("timed out")
}
```

---

## time.After — The Timeout Tool

`time.After(d)` returns a `<-chan time.Time` that receives one value after duration `d`.

```go
ch := time.After(2 * time.Second)
// after 2 seconds, ch gets a value
// select can race this against any other channel
```

It is the standard Go idiom for implementing timeouts.

---

## The Select + Timeout Pattern

```go
select {
case result := <-work():          // case 1: work finishes first
    fmt.Println("result:", result)
case <-time.After(3 * time.Second): // case 2: timeout fires first
    fmt.Println("timed out after 3s")
}
```

Whichever happens first wins. Go's scheduler handles the race.

---

## Mental Model

```
time.After(3s) ─────────────────────────────────────────────→ fires at t=3s
work()         ──────────────────→ finishes at t=1s (wins)

time.After(3s) ──────────────────────────────────────────────→ fires at t=3s (wins)
work()         ──────────────────────────────────────────────────→ finishes at t=5s (too late)
```

`select` is always watching both. First to arrive wins.

---

## Select with Default — Non-Blocking Check

`default` makes select non-blocking — if no channel is ready, it falls through immediately.

```go
select {
case val := <-ch:
    fmt.Println("got:", val)
default:
    fmt.Println("nothing ready yet, moving on")
}
```

Use case: polling a channel without blocking the goroutine.

---

## For + Select — The Event Loop

`select` by itself handles one event. Wrap it in `for` to handle events continuously.

```go
for {
    select {
    case job := <-jobs:
        process(job)
    case <-quit:
        fmt.Println("shutting down")
        return
    case <-time.After(5 * time.Second):
        fmt.Println("idle timeout — no jobs for 5s")
        return
    }
}
```

This is the **Go event loop** pattern. Used in almost every long-running goroutine.

---

## Context — The Production Timeout Tool

`time.After` is simple but `context` is the production standard. It supports:
- Timeouts (`context.WithTimeout`)
- Deadlines (`context.WithDeadline`)
- Cancellation (`context.WithCancel`)
- Propagation across function calls and goroutines

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel() // always cancel to free resources

select {
case result := <-doWork(ctx):
    fmt.Println("result:", result)
case <-ctx.Done():
    fmt.Println("timed out:", ctx.Err()) // context.DeadlineExceeded
}
```

`ctx.Done()` is a channel that closes when timeout fires or cancel() is called.

---

## context.WithTimeout vs context.WithCancel vs context.WithDeadline

| Function | What it does | When to use |
|---|---|---|
| `WithTimeout(ctx, d)` | Cancels after duration d | "Give up after 3 seconds" |
| `WithDeadline(ctx, t)` | Cancels at absolute time t | "Give up at 2pm" |
| `WithCancel(ctx)` | Cancels when you call cancel() | Manual cancellation |

All three return `(ctx, cancelFunc)`. Always call `defer cancel()`.

---

## Why Always defer cancel()?

If you don't call `cancel()`, the context's internal timer goroutine leaks until the parent context is cancelled.

```go
ctx, cancel := context.WithTimeout(parent, 3*time.Second)
defer cancel() // even if work finishes in 1s, resources are freed immediately
```

`defer cancel()` is safe to call multiple times — idempotent.

---

## Select Key Rules

1. **Evaluates all channel expressions first**, then blocks waiting for one to be ready
2. **Random selection** if multiple cases ready simultaneously — no priority
3. **Default runs immediately** if no other case is ready
4. **Nil channel** in a case blocks forever — that case is never selected (useful trick to disable a case)
5. **No fallthrough** — unlike switch, select cases do not fall through

---

## The Nil Channel Trick

Setting a channel to nil disables that select case without removing it from code:

```go
var ch1 <-chan int = someChannel
var ch2 <-chan int = nil // disabled

select {
case val := <-ch1:  // active
    ...
case val := <-ch2:  // never selected — nil channel blocks forever
    ...
}

// Later: enable ch2 by assigning it
ch2 = anotherChannel
```

Useful for dynamically enabling/disabling channels based on program state.

---

## Timeout vs Rate Limiter vs Context

| | Timeout/Select | Rate Limiter | Context |
|---|---|---|---|
| Controls | How long to wait for one op | How fast ops run overall | Cancellation + timeout propagation |
| Tool | `time.After` + `select` | `time.Ticker` | `context.WithTimeout` |
| Scope | Single operation | Stream of operations | Tree of goroutines |
| Passes through layers | No | No | Yes — passes via function args |

---

## Common Mistakes

1. **Missing `default` when you need non-blocking** — goroutine blocks forever
2. **Not calling `defer cancel()`** — goroutine leak from context timer
3. **Assuming case order matters** — Go picks randomly when multiple cases ready
4. **Reusing `time.After` in a loop** — creates a new timer every iteration, leaks timers. Use `time.NewTimer` and reset it instead
5. **Forgetting select needs at least one non-default case** — `select {}` blocks forever (valid use: keep main alive)
6. **Cancelling a child context thinking it cancels parent** — cancellation only flows downward

---

## Real-World Uses

- **HTTP client timeouts** — give up if server doesn't respond in 5s
- **Database query timeouts** — cancel slow queries
- **Graceful shutdown** — wait for goroutines to finish or force-quit after 10s
- **Heartbeat / idle detection** — fire if no message received in N seconds
- **Circuit breaker** — stop calling a failing service after repeated timeouts
- **Fan-out with deadline** — call 3 services, use first response, cancel the rest
