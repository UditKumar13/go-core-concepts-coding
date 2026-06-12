# Worker Pool Pattern in Go

## What is a Worker Pool?

A Worker Pool is a concurrency pattern where a **fixed number of goroutines (workers)** process jobs from a shared channel. Instead of spawning a new goroutine per job (which can overwhelm the system), you create a bounded pool and reuse those goroutines for all incoming work.

---

## The Core Problem it Solves

Imagine you have 10,000 URLs to fetch. If you do:

```go
for _, url := range urls {
    go fetch(url) // spawns 10,000 goroutines — dangerous!
}
```

This creates 10,000 goroutines simultaneously — burning memory and CPU. A Worker Pool fixes this by capping concurrency to N workers.

---

## Mental Model

```
                  Jobs Channel (buffered)
  Dispatcher -->  [job1][job2][job3][job4][job5]...
                       ↓       ↓       ↓
                   Worker1  Worker2  Worker3   (fixed pool)
                       ↓       ↓       ↓
                  Results Channel
                       ↓
                  Main goroutine collects results
```

- Workers **compete** for jobs — whichever is free grabs the next one
- The pool size (N) is the knob that controls concurrency
- Jobs channel being closed signals workers to stop (range over channel exits on close)

---

## Key Components

| Component | Type | Role |
|---|---|---|
| `jobs` | `chan Job` (buffered) | Queue of pending work |
| `results` | `chan Result` (buffered) | Completed work output |
| `worker` | goroutine | Pulls from jobs, pushes to results |
| `WaitGroup` | `sync.WaitGroup` | Tracks when all workers are done |
| dispatcher | loop in main | Fills the jobs channel, then closes it |

---

## Lifecycle

```
1. Create jobs channel + results channel
2. Spin up N workers (each is a goroutine)
3. Dispatcher sends all jobs into jobs channel
4. Dispatcher closes jobs channel (signals no more work)
5. Workers range over jobs channel — exits when channel is closed
6. Each worker calls wg.Done() when it finishes
7. A closer goroutine: wg.Wait() → close(results)
8. Main goroutine ranges over results to collect output
```

---

## Worker Function Anatomy

```go
func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
    defer wg.Done()              // signal done when this goroutine exits

    for job := range jobs {      // blocks until a job arrives, exits when channel closes
        result := process(job)   // do the actual work
        results <- result        // send result out
    }
}
```

`range` over a channel is the idiomatic Go way to consume until the channel is closed.

---

## WaitGroup Flow

```
main:       wg.Add(N)       for each worker before go worker(...)
worker:     wg.Done()       deferred — runs when worker exits its range loop
closer:     wg.Wait()       blocks until all N workers call Done
closer:     close(results)  safe to close only after all writers are done
```

**Why a separate closer goroutine?**
`wg.Wait()` blocks. If you call it in main before ranging over results, you deadlock — main is stuck waiting, but workers can't finish because results is full and nobody is draining it.

---

## Buffered vs Unbuffered Channels

| Channel | Recommended size | Why |
|---|---|---|
| `jobs` | `numJobs` | Dispatcher can fill it without blocking |
| `results` | `numJobs` | Workers can write without blocking |

You can use unbuffered channels too, but then dispatcher and workers must interleave — harder to reason about.

---

## Worker Pool vs Producer/Consumer

| Aspect | Producer/Consumer | Worker Pool |
|---|---|---|
| Primary concern | Data pipeline / transformation | Bounded concurrency |
| Workers | Any number, role-based | Fixed N, all identical |
| Typical use | Streaming, ETL | HTTP fetches, image resize, batch CPU work |
| Shutdown signal | Producer closes jobs channel | Same — dispatcher closes jobs channel |

Worker Pool is essentially Producer/Consumer where the "consumer" is a fixed-size, reusable pool.

---

## Common Mistakes

1. **Closing results before all workers are done** — causes panic (send on closed channel)
2. **Not closing the jobs channel** — workers block forever on `range jobs`
3. **wg.Add inside the goroutine** — race condition; always `wg.Add` before `go`
4. **Deadlock from unbuffered channels** — dispatcher fills jobs, nobody draining yet

---

## Real-World Uses

- HTTP request fan-out (scraping, API calls)
- Image/video processing pipeline
- Database batch inserts
- File parsing (read N files in parallel)
- Any "do this N things at a time" problem
