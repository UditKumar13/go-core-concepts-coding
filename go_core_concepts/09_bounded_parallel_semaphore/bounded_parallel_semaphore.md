# Bounded Parallelism & Semaphore in Go

## What is a Semaphore?

A Semaphore is a concurrency primitive that controls how many goroutines can access a resource or run a section of code **simultaneously**.

Think of it as a nightclub with a capacity limit:
```
Nightclub capacity = 3
Door policy: only let someone in when someone leaves

[G1 enters] → capacity used: 1/3
[G2 enters] → capacity used: 2/3
[G3 enters] → capacity used: 3/3
[G4 arrives] → WAIT at the door
[G1 leaves]  → G4 can now enter
```

In code terms: at most N goroutines inside the critical section at any time.

---

## What is Bounded Parallelism?

Bounded Parallelism = running work concurrently but **capping the maximum concurrency** at N.

```
Without bound: 1000 goroutines all run at once → memory explodes, CPU thrashes
With bound N=5: only 5 goroutines run at once → 1000 jobs complete steadily
```

It answers the question: **"How do I process many tasks concurrently without overwhelming the system?"**

---

## How Go Implements a Semaphore — Buffered Channel

Go has no built-in semaphore type. The idiomatic Go semaphore is a **buffered channel of empty structs**.

```go
sem := make(chan struct{}, N) // N = max concurrent goroutines

// Acquire (enter): send into channel — blocks when channel is full
sem <- struct{}{}

// Release (exit): receive from channel — unblocks one waiting goroutine
<-sem
```

Why `struct{}`? It occupies **zero bytes** of memory. We only care about the count, not the value.

---

## Mental Model

```
sem := make(chan struct{}, 3)  // capacity 3

Goroutine 1: sem <- {} → [■][ ][ ]  running
Goroutine 2: sem <- {} → [■][■][ ]  running
Goroutine 3: sem <- {} → [■][■][■]  running
Goroutine 4: sem <- {} → BLOCKS (channel full, waits)
Goroutine 5: sem <- {} → BLOCKS (channel full, waits)

Goroutine 1 finishes: <-sem → [■][■][ ]
Goroutine 4 unblocks: sem <- {} → [■][■][■]  now running
```

The channel itself enforces the limit — no extra logic needed.

---

## The Pattern

```go
sem := make(chan struct{}, maxConcurrent)

for _, task := range tasks {
    sem <- struct{}{}          // acquire: blocks if N already running

    go func(t Task) {
        defer func() { <-sem }() // release on exit
        process(t)
    }(task)
}

// Wait for all in-flight goroutines to finish
for i := 0; i < maxConcurrent; i++ {
    sem <- struct{}{}
}
```

Or combine with WaitGroup for cleaner shutdown:

```go
sem := make(chan struct{}, maxConcurrent)
var wg sync.WaitGroup

for _, task := range tasks {
    sem <- struct{}{}   // acquire
    wg.Add(1)

    go func(t Task) {
        defer wg.Done()
        defer func() { <-sem }() // release
        process(t)
    }(task)
}

wg.Wait()
```

---

## Acquire and Release — The Contract

| Operation | Channel op | Meaning |
|---|---|---|
| Acquire (enter) | `sem <- struct{}{}` | "I am starting — take one slot" |
| Release (exit) | `<-sem` | "I am done — free one slot" |

**Always release in a `defer`** so the slot is freed even if the goroutine panics.

```go
sem <- struct{}{}          // acquire
defer func() { <-sem }()  // release — guaranteed to run
```

---

## Semaphore vs Worker Pool

Both bound concurrency. The difference is structure:

| | Worker Pool | Semaphore |
|---|---|---|
| Goroutines | Fixed N goroutines, reused | New goroutine per task, bounded by sem |
| Jobs | Sent through a channel | Spawned inline in a loop |
| Lifetime | Workers live for program duration | Goroutine dies after each task |
| Control | Channel size = worker count | Semaphore size = max concurrent |
| Flexibility | All workers do same job | Each goroutine can be different |

**Use Worker Pool** when: tasks are uniform and you want to reuse goroutines.
**Use Semaphore** when: tasks vary or you want to bound goroutines spawned dynamically.

---

## Semaphore vs Rate Limiter

| | Semaphore | Rate Limiter |
|---|---|---|
| Controls | How many run simultaneously | How many run per second |
| Tool | Buffered channel (count) | time.Ticker (time) |
| Question | "How many at once?" | "How many per second?" |
| Can combine? | Yes — both together in production |  |

---

## Why struct{} and Not int or bool?

```go
sem := make(chan struct{}, N) // correct — zero memory per token
sem := make(chan bool, N)     // works but wastes 1 byte per slot
sem := make(chan int, N)      // works but wastes 8 bytes per slot
```

`struct{}` is the Go idiom for "signal without data". It communicates intent: the value doesn't matter, only the count.

---

## Lifecycle Step by Step

```
1. Create buffered channel of capacity N
2. Loop over all tasks:
   a. sem <- struct{}{}   → acquire slot (blocks if N already running)
   b. go func() {         → spawn goroutine
       defer <-sem         → release slot when done
       process(task)
      }()
3. wg.Wait()              → wait for all goroutines to finish
```

---

## Draining Pattern — Wait Without WaitGroup

If you don't use WaitGroup, you can wait by filling the semaphore completely:

```go
// after loop, drain remaining slots
for i := 0; i < cap(sem); i++ {
    sem <- struct{}{} // blocks until a running goroutine releases
}
```

When all N slots are filled with your drain tokens, all goroutines have finished. Less common — WaitGroup is cleaner.

---

## Common Mistakes

1. **Releasing before work is done** — slot freed early, N+1 goroutines run simultaneously
2. **Not using defer for release** — if goroutine panics, slot is never freed → semaphore leaks → system grinds to halt
3. **Acquire inside the goroutine** — goroutine is already spawned before acquiring, concurrency is not actually bounded
4. **Wrong cap** — `make(chan struct{}, 0)` is unbuffered, every send blocks immediately → nothing runs
5. **Forgetting to wait** — main exits before goroutines finish

---

## Real-World Uses

- **HTTP request fan-out** — call 100 APIs but max 10 at once
- **File processing** — open max N files simultaneously (OS file descriptor limit)
- **Database connections** — max N concurrent queries (connection pool)
- **Image/video encoding** — max N CPU-intensive jobs (CPU core count)
- **Scraping** — max N concurrent fetches per domain
- **CI/CD** — max N parallel test suites
