# Job Scheduler in Go

## What is a Job Scheduler?

A job scheduler runs tasks automatically at fixed time intervals — without you manually triggering them.

```
CleanupJob   → runs every 2 seconds
HeartbeatJob → runs every 3 seconds
ReportJob    → runs every 5 seconds
```

All jobs run **concurrently** in the background until the scheduler is stopped.

---

## Real World Examples

| Job | Interval | Purpose |
|-----|----------|---------|
| Cleanup temp files | Every 1 hour | Free up disk space |
| Heartbeat ping | Every 30 seconds | Tell server "I'm alive" |
| Generate report | Every 24 hours | Daily summary email |
| Database backup | Every 6 hours | Data safety |
| Cache invalidation | Every 5 minutes | Keep data fresh |

---

## Components

### Job Struct

```go
type Job struct {
    Name     string
    Interval time.Duration
    Task     func()       // The actual work to do
}
```

Each job has a name, how often to run, and the function to execute.

---

### Scheduler Struct

```go
type Scheduler struct {
    jobs   []Job
    wg     sync.WaitGroup   // Wait for all jobs to stop
    cancel context.CancelFunc // Signal to stop all jobs
    ctx    context.Context    // Carries the stop signal
}
```

- `wg` — waits for all goroutines to finish before `Stop()` returns
- `ctx` + `cancel` — used to signal all jobs to stop at once

---

## How Each Job Runs

```go
func (s *Scheduler) run(job Job) {
    defer s.wg.Done()
    ticker := time.NewTicker(job.Interval)  // Fires every N seconds
    defer ticker.Stop()

    for {
        select {
        case <-s.ctx.Done():     // Stop signal received
            return
        case t := <-ticker.C:   // Ticker fired → run the task
            job.Task()
        }
    }
}
```

### The select block:

| Channel | What it means |
|---------|---|
| `s.ctx.Done()` | Scheduler was stopped → exit goroutine |
| `ticker.C` | Interval elapsed → run the task |

The goroutine **blocks** until one of these fires, then acts accordingly.

---

## Flow

```
scheduler.Start()
    ↓
For each job → launch goroutine
    ↓
Each goroutine:
    ┌─────────────────────────────┐
    │  Wait for ticker or stop    │
    │  ticker fires → run task    │
    │  stop signal → exit         │
    └─────────────────────────────┘

scheduler.Stop()
    ↓
cancel() → sends stop signal to ALL goroutines via ctx
    ↓
wg.Wait() → blocks until all goroutines exit
    ↓
"All jobs stopped"
```

---

## Timeline Example (from output)

```
T=0s   Scheduler starts
T=2s   CleanupJob runs
T=3s   HeartbeatJob runs
T=4s   CleanupJob runs
T=5s   ReportJob runs
T=6s   CleanupJob + HeartbeatJob run
T=8s   CleanupJob runs
T=9s   HeartbeatJob runs
T=10s  Scheduler stops → all jobs get stop signal → exit
```

---

## Key Concepts Used

### time.NewTicker

```go
ticker := time.NewTicker(2 * time.Second)

// ticker.C sends a value every 2 seconds
case t := <-ticker.C:
    // runs every 2s
```

Different from `time.Sleep` — ticker fires repeatedly on a fixed interval, not just once.

### context.WithCancel

```go
ctx, cancel := context.WithCancel(context.Background())

// Later:
cancel()  // ctx.Done() channel closes → all goroutines listening exit
```

One `cancel()` call stops ALL goroutines simultaneously.

### sync.WaitGroup

```go
s.wg.Add(1)      // Before starting goroutine
defer s.wg.Done() // When goroutine exits

s.wg.Wait()      // Block until all goroutines call Done()
```

Ensures `Stop()` doesn't return until every job has fully exited.

---

## WaitGroup vs Context — Who Does What?

| Tool | Role |
|------|------|
| `context.Cancel` | **Signals** goroutines to stop |
| `sync.WaitGroup` | **Waits** for goroutines to actually finish |

Cancel tells them to stop. WaitGroup confirms they stopped.
