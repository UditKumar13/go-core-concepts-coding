# Fan-In / Fan-Out Pattern in Go

## What is Fan-Out?

Fan-Out means taking **one input channel** and distributing its work across **multiple goroutines** in parallel.

One source → many workers

```
            jobs channel (one source)
                    ↓
       ┌────────────┼────────────┐
    Worker1      Worker2      Worker3     ← fanned out
       ↓            ↓            ↓
  results1      results2      results3
```

Each worker gets its own results channel. They all read from the same jobs channel.

---

## What is Fan-In?

Fan-In means **merging multiple channels into one** so a single consumer can read all results.

Many sources → one channel

```
  results1 ──┐
  results2 ──┼──→  merged channel  →  main (single consumer)
  results3 ──┘
```

A merger goroutine listens on all input channels and forwards everything into one output channel.

---

## Fan-Out + Fan-In Together

In practice you almost always use them together:

```
Input Channel
      ↓
   Fan-Out → [Worker1] [Worker2] [Worker3]   (parallel processing)
                 ↓          ↓         ↓
              [ch1]      [ch2]      [ch3]
                 ↓          ↓         ↓
              Fan-In  →  merged channel      (single result stream)
                              ↓
                           main()
```

This is the **pipeline with parallel stages** pattern.

---

## Why Not Just Use a Worker Pool?

Great question. Compare:

| | Worker Pool | Fan-Out + Fan-In |
|---|---|---|
| Workers share | One results channel | Each worker has its own channel |
| Merging | Not needed (shared channel) | Fan-In explicitly merges channels |
| Flexibility | All workers do same job | Each worker can be a different function |
| Use case | N identical workers on a queue | Parallel pipelines, scatter-gather |

**Worker Pool** = shared job queue, shared results channel, N identical workers.

**Fan-Out/Fan-In** = spread work across N goroutines each with own channel, then merge back.

Fan-Out/Fan-In gives you more control — each goroutine can be a completely different stage or operation.

---

## Fan-Out in Code (concept)

```go
// fanOut takes one input channel and spawns N workers.
// Each worker gets its own output channel.
// Returns a slice of those output channels.
func fanOut(input <-chan int, numWorkers int) []<-chan int {
    channels := make([]<-chan int, numWorkers)
    for i := 0; i < numWorkers; i++ {
        channels[i] = worker(i, input) // each worker returns its own chan
    }
    return channels
}
```

---

## Fan-In in Code (concept)

```go
// fanIn merges multiple input channels into one output channel.
// Spawns one goroutine per input channel to forward values.
// Closes output when all inputs are exhausted.
func fanIn(channels ...<-chan int) <-chan int {
    merged := make(chan int)
    var wg sync.WaitGroup

    forward := func(ch <-chan int) {
        defer wg.Done()
        for val := range ch {
            merged <- val
        }
    }

    wg.Add(len(channels))
    for _, ch := range channels {
        go forward(ch)
    }

    go func() {
        wg.Wait()
        close(merged) // safe to close only after all forwarders finish
    }()

    return merged
}
```

`...` (variadic) means fanIn accepts any number of channels.

---

## Key Components

| Component | Role |
|---|---|
| `input <-chan` | Single source of jobs |
| `worker(input)` | Returns its own `<-chan` result |
| `fanOut()` | Spawns workers, returns `[]<-chan` |
| `fanIn(channels...)` | Merges all worker channels into one |
| `merged <-chan` | Single stream main reads from |
| WaitGroup in fanIn | Closes merged only after all inputs drain |

---

## Lifecycle Step by Step

```
1. Create input channel, send jobs into it, close it
2. fanOut: spawn N workers, each reading from input, each with own output channel
3. fanIn: spawn N forwarder goroutines, one per worker channel
4. Each forwarder: range over its worker channel → send into merged channel
5. WaitGroup in fanIn: when all forwarders done → close merged
6. main: range over merged → collect all results
```

---

## The WaitGroup Lives Inside fanIn

This is different from Worker Pool where WaitGroup was in main.
Here fanIn owns the WaitGroup because fanIn is responsible for knowing when all its input channels are drained.

```
fanIn                           main
  wg.Add(len(channels))
  go forward(ch1) → wg.Done()
  go forward(ch2) → wg.Done()
  go forward(ch3) → wg.Done()
  go wg.Wait() → close(merged)
                                for val := range merged { ... }
```

main just ranges over merged — it doesn't need to know how many workers exist.

---

## Common Mistakes

1. **Closing merged inside a forwarder** — only one forwarder would close it; others would panic (send on closed channel). Always close merged in a separate goroutine after `wg.Wait()`
2. **Not using WaitGroup in fanIn** — merged channel never closes, main blocks forever
3. **Workers reading from the same channel incorrectly** — in fan-out, all workers share one input channel and compete for jobs (correct). Don't copy the channel.
4. **Returning a non-closed channel from worker** — fanIn's forwarder will block forever on `range`

---

## Fan-In / Fan-Out vs Pipeline

| | Fan-In/Fan-Out | Pipeline |
|---|---|---|
| Shape | Scatter → Gather | Linear stages A → B → C |
| Parallelism | Horizontal (same stage, many workers) | Vertical (different stages in sequence) |
| Next topic | This one | 05_pipeline |

They are often combined: a pipeline where one stage fans out and fans back in.

---

## Real-World Uses

- **Search**: fan out query to multiple search engines, fan in first result
- **Microservices**: fan out one request to multiple downstream services, merge responses
- **Data aggregation**: fan out to multiple databases, fan in results
- **Map-Reduce**: fan out = map phase, fan in = reduce phase
- **Image processing**: fan out one image to multiple filter goroutines, fan in processed tiles
