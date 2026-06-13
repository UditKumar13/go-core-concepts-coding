# Pipeline Pattern in Go

## What is a Pipeline?

A Pipeline is a series of **stages connected by channels**, where each stage:
1. Receives values from an upstream channel
2. Transforms or processes those values
3. Sends results to a downstream channel

Data flows in one direction: source → stage1 → stage2 → stage3 → sink

---

## Mental Model

```
Source          Stage 1         Stage 2         Stage 3         Sink
(generate)   (transform)      (filter)        (format)       (collect)

numbers  →   [× 2 each]  →  [keep evens]  →  [add prefix]  →  print/store
chan int  →   chan int    →   chan int      →   chan string   →  main()
```

Each stage is a function that:
- Takes an input `<-chan`
- Returns an output `<-chan`
- Runs its own goroutine internally

---

## Why Pipelines?

Without pipelines, you process data in sequential batches:
```go
// batch style — must finish step1 for ALL items before starting step2
step1Results := doStep1(allData)     // wait...
step2Results := doStep2(step1Results) // wait...
step3Results := doStep3(step2Results) // wait...
```

With a pipeline, stages run **concurrently**:
```
item1: [stage1]→[stage2]→[stage3]
item2:          [stage1]→[stage2]→[stage3]
item3:                   [stage1]→[stage2]→[stage3]
```
While stage2 processes item1, stage1 is already processing item2. All stages stay busy.

---

## A Single Stage — The Template

Every stage follows this exact same shape:

```go
func stageName(input <-chan int) <-chan int {
    output := make(chan int)

    go func() {
        defer close(output)        // stage closes its own output when input is exhausted
        for val := range input {   // receive from upstream
            result := transform(val)
            output <- result       // send to downstream
        }
    }()

    return output                  // caller chains this into the next stage
}
```

This shape is the whole pattern. Learn this, and you can build any pipeline.

---

## Chaining Stages

Because each stage takes a channel and returns a channel, you chain them like this:

```go
// verbose style
s1 := generate(1, 2, 3, 4, 5)
s2 := double(s1)
s3 := filterEven(s2)
s4 := square(s3)

// inline style — reads like a pipeline
results := square(filterEven(double(generate(1, 2, 3, 4, 5))))
```

Both are identical. Inline reads like a data flow sentence.

---

## The Source Stage (Generator)

The first stage has no input channel — it creates data and puts it into a channel.

```go
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}
```

Takes plain values, returns a channel. This is the entry point of the pipeline.

---

## The Sink (Consumer)

The last stage has no output channel — it consumes the final channel and produces a result.

```go
// sink: collect all values from the final stage
total := 0
for val := range finalStage {
    total += val
}
```

Or it could print, write to a file, insert into a database, etc.

---

## Key Components

| Component | Role |
|---|---|
| Generator / Source | Creates data, returns first `<-chan` |
| Stage function | Takes `<-chan`, transforms, returns new `<-chan` |
| Internal goroutine | Runs inside each stage, does the actual work |
| `defer close(output)` | Signals downstream when stage is done |
| Sink / Consumer | Ranges over final channel, collects results |

---

## Lifecycle Step by Step

```
1. Generator sends values into chan1, closes it when done
2. Stage1 goroutine ranges chan1 → transforms → sends into chan2 → closes chan2
3. Stage2 goroutine ranges chan2 → transforms → sends into chan3 → closes chan3
4. Stage3 goroutine ranges chan3 → transforms → sends into chan4 → closes chan4
5. Sink (main) ranges chan4 → collects results
6. All goroutines exit naturally — no WaitGroup needed
```

No WaitGroup! The channel close signal cascades automatically stage by stage.

---

## Channel Close Cascade

This is what makes pipelines elegant:

```
generate closes chan1
    → stage1's range exits → stage1 closes chan2
        → stage2's range exits → stage2 closes chan3
            → stage3's range exits → stage3 closes chan4
                → main's range exits → program ends
```

One close at the source triggers a clean shutdown of the entire pipeline automatically.

---

## Pipeline vs Fan-Out/Fan-In

| | Pipeline | Fan-Out / Fan-In |
|---|---|---|
| Shape | Linear: A → B → C | Scatter: one → many → one |
| Parallelism | Between stages (concurrent stages) | Within a stage (parallel workers) |
| Data flow | Every item goes through every stage | Items distributed across parallel workers |
| Combine? | Yes — fan out inside one pipeline stage | Yes — each fan-out worker is a sub-pipeline |

**Combined**: A pipeline where one stage fans out to N goroutines and fans back in — this is the most powerful pattern and what production systems use.

---

## Pipeline vs Worker Pool

| | Worker Pool | Pipeline |
|---|---|---|
| Structure | One jobs channel, N identical workers | Multiple stages, each a different transform |
| Purpose | Bound concurrency on one task | Chain multiple different transformations |
| Stages | One | Many |
| Channel count | 2 (jobs + results) | N+1 (one per stage boundary) |

---

## Buffered vs Unbuffered Stage Channels

```go
output := make(chan int)    // unbuffered — stage blocks until downstream reads
output := make(chan int, N) // buffered  — stage can run N steps ahead of downstream
```

Unbuffered: stages stay tightly in sync. If stage2 is slow, stage1 slows too (backpressure).
Buffered: stages can run ahead. Smooths out speed differences between stages.

**Default**: start unbuffered. Add buffer only when you measure a bottleneck.

---

## Common Mistakes

1. **Not closing the output channel in a stage** — downstream range never exits, goroutine leaks
2. **Closing the channel from outside the stage** — only the writer (the goroutine) should close its own channel
3. **Forgetting `defer close(output)`** — if the stage panics, channel never closes, pipeline hangs
4. **Blocking the generator in main** — generator must run in a goroutine or use a buffered channel, else it deadlocks
5. **Sharing one channel across stages** — each stage boundary needs its own channel

---

## Real-World Uses

- **ETL** (Extract → Transform → Load): read CSV → clean data → insert to DB
- **Image processing**: decode → resize → compress → save
- **Log processing**: read lines → parse → filter errors → alert
- **Compiler**: tokenize → parse → type-check → generate code
- **HTTP middleware**: request → auth → rate-limit → handler → response
- **Video streaming**: decode frame → apply filter → encode → stream to client
