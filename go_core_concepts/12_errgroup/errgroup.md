# Errgroup in Go

## What is Errgroup?

`errgroup` is a package from `golang.org/x/sync` that simplifies managing a group of goroutines and collecting errors from all of them.

**Problem without errgroup:**
```go
var wg sync.WaitGroup
var mu sync.Mutex
var errors []error

// Need WaitGroup to track goroutines
// Need Mutex to safely store errors
// Need to manually check errors
```

**Solution with errgroup:**
```go
var eg errgroup.Group
// That's it! Handles goroutines + error collection
```

---

## Basic Usage

### Without Errgroup (verbose):

```go
var wg sync.WaitGroup
var mu sync.Mutex
var errs []error

wg.Add(1)
go func() {
    defer wg.Done()
    err := doSomething()
    if err != nil {
        mu.Lock()
        errs = append(errs, err)
        mu.Unlock()
    }
}()

wg.Add(1)
go func() {
    defer wg.Done()
    err := doAnotherThing()
    if err != nil {
        mu.Lock()
        errs = append(errs, err)
        mu.Unlock()
    }
}()

wg.Wait()
if len(errs) > 0 {
    // handle errors
}
```

### With Errgroup (clean):

```go
var eg errgroup.Group

eg.Go(func() error {
    return doSomething()
})

eg.Go(func() error {
    return doAnotherThing()
})

if err := eg.Wait(); err != nil {
    // handle error (first one encountered)
}
```

---

## Core Methods

| Method | What it does |
|--------|---|
| `eg.Go(func)` | Start a goroutine that returns an error |
| `eg.Wait()` | Block until all goroutines finish, return first error (or nil) |
| `eg.SetLimit(n)` | Limit concurrent goroutines to `n` at a time |
| `errgroup.WithContext(ctx)` | Create errgroup with context for cancellation |

---

## Four Patterns

### 1. Basic errgroup (no context)

```go
var eg errgroup.Group

eg.Go(func() error {
    return fetchUserData(1)
})

eg.Go(func() error {
    return fetchPostData(1)
})

if err := eg.Wait(); err != nil {
    fmt.Println("Error:", err)  // Returns FIRST error
}
```

**Behavior:** All goroutines run. If any fails, `Wait()` returns the first error (others still run).

---

### 2. With Context (cancellation)

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

eg, egCtx := errgroup.WithContext(ctx)

eg.Go(func() error {
    return fetchUserData(2)  // fails
})

eg.Go(func() error {
    select {
    case <-egCtx.Done():  // Context cancelled if any goroutine errors
        return egCtx.Err()
    case <-time.After(2 * time.Second):
        return fetchPostData(1)
    }
})

if err := eg.Wait(); err != nil {
    fmt.Println("Error:", err)
}
```

**Behavior:** If one goroutine errors, the context is cancelled. Other goroutines can check `egCtx.Done()` and stop early.

---

### 3. Collecting Results

```go
var eg errgroup.Group
results := make([]string, 0)
var mu sync.Mutex

for i := 1; i <= 3; i++ {
    id := i
    eg.Go(func() error {
        result := compute(id)
        mu.Lock()
        results = append(results, result)
        mu.Unlock()
        return nil
    })
}

eg.Wait()
fmt.Println(results)
```

**Behavior:** Each goroutine appends to a shared `results` slice (protected by mutex).

---

### 4. Limiting Concurrency

```go
var eg errgroup.Group
eg.SetLimit(2)  // Only 2 goroutines run at a time

for i := 1; i <= 10; i++ {
    id := i
    eg.Go(func() error {
        return doWork(id)
    })
}

eg.Wait()
```

**Behavior:** All 10 tasks are queued, but only 2 run concurrently. When one finishes, the next starts.

---

## Comparison: WaitGroup vs Errgroup

| Feature | sync.WaitGroup | errgroup.Group |
|---------|---|---|
| Track goroutines | ✓ | ✓ |
| Collect errors | ✗ (manual) | ✓ (automatic) |
| Cancel on error | ✗ | ✓ (with context) |
| Limit concurrency | ✗ | ✓ (SetLimit) |
| Return error | ✗ | ✓ (Wait returns error) |

---

## Real World Use Cases

| Use Case | Why errgroup? |
|----------|---|
| Parallel API calls | Fetch multiple endpoints, fail if any error |
| Batch processing | Process items concurrently, stop on error |
| Worker pools | Limit concurrent workers, collect errors |
| Migrations | Run multiple migrations, abort on first failure |
| Health checks | Check multiple services, report any failures |

---

## Key Insight

**Errgroup = WaitGroup + error handling + optional context cancellation + optional concurrency limit**

It's basically WaitGroup on steroids.
