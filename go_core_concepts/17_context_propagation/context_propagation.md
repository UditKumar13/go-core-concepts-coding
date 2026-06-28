# Context Propagation in Go

## What is Context?

`context.Context` is a way to carry **deadlines, cancellation signals, and values** across function calls and goroutines.

```
Request comes in
     ↓
Handler creates context
     ↓
Passes ctx to function A
     ↓
A passes ctx to function B
     ↓
B passes ctx to function C
```

Every function in the chain shares the same context — same deadline, same values, same cancellation signal.

---

## Three Things Context Carries

| Thing | How | Use Case |
|-------|-----|---------|
| Values | `context.WithValue` | Pass userID, requestID through layers |
| Timeout | `context.WithTimeout` | Cancel if takes too long |
| Cancellation | `context.WithCancel` | Stop all children when parent cancels |

---

## Example 1: Value Propagation

```go
ctx = context.WithValue(ctx, "userID", "user-42")
ctx = context.WithValue(ctx, "requestID", "req-abc")

handleRequest(ctx)
  → processData(ctx)
    → saveToDB(ctx)
```

Each layer reads from ctx without being explicitly passed the values:

```go
userID := ctx.Value("userID")  // works at any depth
```

**Output:**
```
handleRequest: userID=user-42 requestID=req-abc123
processData:   userID=user-42 (propagated from parent)
saveToDB:      userID=user-42 (propagated through all layers)
```

---

## Example 2: Timeout Propagation

```go
// Parent has 5s timeout
parentCtx, _ := context.WithTimeout(context.Background(), 5*time.Second)

// Child has 1s timeout (stricter)
childCtx, _ := context.WithTimeout(parentCtx, 1*time.Second)

fetchData(childCtx)  // takes 2s → cancelled after 1s
```

The child context expires first. If the parent expired first, it would cancel the child too.

**Rule:** Child timeout can only be equal or shorter than parent. Never longer.

---

## Example 3: Cancellation Propagation

```go
rootCtx, rootCancel := context.WithCancel(context.Background())

go worker(rootCtx, 1)  // all workers share same ctx
go worker(rootCtx, 2)
go worker(rootCtx, 3)

rootCancel()  // cancels ALL three workers at once
```

**Output:**
```
Worker 1 doing work...
Worker 2 doing work...
Worker 3 doing work...
Parent cancelling context...
Worker 1 stopped: context canceled
Worker 2 stopped: context canceled
Worker 3 stopped: context canceled
```

One `cancel()` call stops all goroutines.

---

## How Workers Listen for Cancellation

```go
func worker(ctx context.Context, id int) {
    for {
        select {
        case <-ctx.Done():   // ← Stop signal received
            return
        case <-time.After(500 * time.Millisecond):
            doWork()
        }
    }
}
```

`ctx.Done()` returns a channel that **closes** when the context is cancelled or timed out. Any goroutine blocking on `<-ctx.Done()` will unblock immediately.

---

## Context Hierarchy

```
context.Background()
    └── WithCancel → rootCtx
            └── WithTimeout(5s) → parentCtx
                    └── WithTimeout(1s) → childCtx
                            └── WithValue("userID") → requestCtx
```

When a parent is cancelled → ALL children are cancelled.
When a child is cancelled → parent is NOT affected.

---

## Always Pass ctx as First Argument

```go
// Standard Go convention
func fetchUser(ctx context.Context, userID string) (*User, error)
func saveOrder(ctx context.Context, order Order) error
func sendEmail(ctx context.Context, to string) error
```

This allows any caller to control the deadline/cancellation of any operation.

---

## Key Rules

| Rule | Reason |
|------|--------|
| Always `defer cancel()` | Prevents context leak (memory/goroutine leak) |
| Never store ctx in a struct | Context is per-request, not per-object |
| Pass ctx as first param | Go convention, keeps API consistent |
| Check `ctx.Err()` after Done | Tells you WHY it was cancelled (timeout vs cancel) |
