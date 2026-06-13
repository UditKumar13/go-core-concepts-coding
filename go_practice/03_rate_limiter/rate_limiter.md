# Rate Limiter Pattern in Go

## What is a Rate Limiter?

A Rate Limiter controls **how fast** work is allowed to happen — not just how many goroutines run, but how many operations are permitted per unit of time.

Example: "Process at most 5 jobs per second", regardless of how many workers or jobs exist.

---

## The Core Problem it Solves

You have a Worker Pool running 10 workers hitting an external API.
The API allows only 3 requests per second or it bans you.

Without a rate limiter:
```
t=0ms  → Worker1, Worker2, Worker3, Worker4... all fire at once → API bans you
```

With a rate limiter:
```
t=0ms   → Worker1 gets token → fires request
t=333ms → Worker2 gets token → fires request
t=666ms → Worker3 gets token → fires request
t=1000ms → Worker4 gets token → fires request
```

---

## Mental Model — The Token Bucket

Imagine a bucket that gets filled with tokens at a fixed rate.
Each worker must take one token before doing work.
If the bucket is empty, the worker waits.

```
time.Ticker (every 200ms)
       ↓
  [token][token][token]   ← token channel (the bucket)
       ↓       ↓       ↓
  Worker1  Worker2  Worker3
  (each consumes one token before proceeding)
```

The **Ticker fires at a fixed rate** → puts tokens into the channel → workers consume them.
Rate of work = rate of the ticker. Simple and precise.

---

## Two Common Approaches in Go

### 1. Simple Rate Limiter — time.Ticker

```go
ticker := time.NewTicker(200 * time.Millisecond) // 5 per second
defer ticker.Stop()

for job := range jobs {
    <-ticker.C   // wait for next tick = wait for permission
    process(job)
}
```

Every job must wait for the next tick. Clean, simple, no channel needed.

### 2. Bursty Rate Limiter — Buffered Token Channel

```go
limiter := make(chan time.Time, 3) // allows burst of 3

// fill initial burst
for i := 0; i < 3; i++ {
    limiter <- time.Now()
}

// refill at fixed rate
go func() {
    ticker := time.NewTicker(200 * time.Millisecond)
    for t := range ticker.C {
        limiter <- t // add one token every 200ms
    }
}()

// worker consumes a token before each job
<-limiter
process(job)
```

Allows an initial burst (3 jobs immediately), then throttles to the ticker rate.

---

## Key Components

| Component | Type | Role |
|---|---|---|
| `time.Ticker` | `*time.Ticker` | Fires at fixed interval — the clock of the limiter |
| `ticker.C` | `<-chan time.Time` | Channel that receives a value on each tick |
| `limiter` | `chan time.Time` (buffered) | Token bucket — buffered size = allowed burst |
| Worker | goroutine | Blocks on `<-limiter` before doing work |

---

## Lifecycle — Simple Rate Limiter

```
1. Create ticker with desired interval (interval = 1s / rate)
2. Spin up workers
3. Each worker: before processing a job → <-ticker.C (blocks until next tick)
4. Tick fires → one worker unblocks → processes one job
5. Next tick fires → next worker unblocks → processes next job
6. On shutdown: ticker.Stop() to free resources
```

---

## Calculating the Interval

```
rate  = 5 jobs/second
interval = 1s / 5 = 200ms per job

rate  = 100 jobs/minute  
interval = 60s / 100 = 600ms per job

rate  = 10 jobs/second
interval = 1s / 10 = 100ms per job
```

---

## Simple vs Bursty — When to Use Which

| | Simple (Ticker only) | Bursty (Buffered + Ticker) |
|---|---|---|
| Behavior | Strict — one job per tick, always | Allows N jobs immediately, then throttles |
| Use case | Strict API rate limits | User-facing features where initial speed matters |
| Complexity | Very simple | Slightly more complex |
| Example | External API: max 5 req/sec | Login attempts: allow 3 fast, then slow down |

---

## Rate Limiter vs Worker Pool

| Aspect | Worker Pool | Rate Limiter |
|---|---|---|
| Controls | **How many** run simultaneously | **How fast** jobs are processed |
| Mechanism | Fixed number of goroutines | Ticker / token channel |
| Question answered | "How many concurrent?" | "How many per second?" |
| Can combine? | Yes — use both together | Yes — pool of workers each respecting a shared limiter |

**Combined pattern** (very common in production):
```
Worker Pool (bounds concurrency) + Rate Limiter (bounds speed)
= "Max 5 workers, max 10 jobs/second"
```

---

## time.Ticker vs time.Sleep

You might think: why not just `time.Sleep(200ms)` in each worker?

```go
// naive approach
process(job)
time.Sleep(200 * time.Millisecond)
```

Problem: if `process(job)` takes 150ms, sleep adds another 200ms → actual rate = 1 job per 350ms, not 200ms. The rate drifts based on processing time.

`time.Ticker` fires on a wall-clock schedule regardless of how long the work took. It is **self-correcting**. If work takes 150ms and tick interval is 200ms, the worker only waits 50ms for the next token. Rate stays accurate.

---

## Common Mistakes

1. **Forgetting `ticker.Stop()`** — ticker goroutine leaks, keeps firing forever
2. **One ticker per worker** — each worker gets its own rate, total rate = N × rate. Use one shared ticker/limiter
3. **Buffered limiter too large** — defeats the purpose; burst becomes unlimited
4. **Using Sleep instead of Ticker** — rate drifts with processing time (see above)
5. **Not accounting for startup burst** — simple ticker makes first job wait for first tick; pre-fill if you want immediate start

---

## Real-World Uses

- External API calls (GitHub, Stripe, Twitter — all have rate limits)
- Web scraping without getting IP-banned
- Database write throttling during bulk imports
- SMS / email sending (carriers limit per second)
- Login attempt throttling (security)
- Video encoding job queues (CPU budget per second)
