# Concurrent File Processor in Go

## What is it?

Instead of processing files one by one (sequentially), we process **multiple files at the same time** using goroutines — then collect all results.

```
Sequential:   file1 → file2 → file3 → file4 → file5   (slow)
Concurrent:   file1 ┐
              file2 ├→ all at once → collect results   (fast)
              file3 ┘
```

---

## Two Approaches

| Approach | How | When to use |
|----------|-----|-------------|
| One goroutine per file | Each file gets its own goroutine | Small number of files |
| Worker pool | Fixed N workers process all files | Large number of files |

---

## Approach 1: One Goroutine Per File

```go
for name, content := range files {
    wg.Add(1)
    go func(n, c string) {
        defer wg.Done()
        result := processFile(n, c)
        mu.Lock()
        results = append(results, result)
        mu.Unlock()
    }(name, content)
}
wg.Wait()
```

### Key pieces:

| Piece | Purpose |
|-------|---------|
| `wg.Add(1)` | Track one more goroutine |
| `go func(...)` | Launch goroutine per file |
| `mu.Lock()` | Protect shared `results` slice |
| `wg.Wait()` | Block until all goroutines finish |

### Why pass `name, content` as args?

```go
// BAD: closure captures loop variable by reference
go func() {
    processFile(name, content)  // 'name' changes every iteration!
}()

// GOOD: pass as argument, copies the value
go func(n, c string) {
    processFile(n, c)  // 'n' is a fixed copy
}(name, content)
```

---

## Approach 2: Worker Pool

```go
jobs    := make(chan job, len(files))   // buffered channel of work
results := make(chan FileResult, len(files))

// Start N workers
for i := 0; i < workerCount; i++ {
    go func(workerID int) {
        for j := range jobs {      // pull jobs until channel closes
            results <- processFile(j.name, j.content)
        }
    }(i + 1)
}

// Send all jobs
for name, content := range files {
    jobs <- job{name, content}
}
close(jobs)   // signals workers: no more work coming
```

### Flow with 2 workers, 5 files:

```
jobs channel: [file1, file2, file3, file4, file5]

Worker 1 picks file1 → processes → sends result
Worker 2 picks file2 → processes → sends result
Worker 1 picks file3 → processes → sends result   (worker 1 is free again)
Worker 2 picks file4 → processes → sends result
Worker 1 picks file5 → processes → sends result
```

Workers keep pulling from `jobs` until it's empty and closed.

---

## Why Use a Worker Pool?

**Problem with one goroutine per file:**

```
10,000 files → 10,000 goroutines launched at once
             → too much memory + CPU overhead
```

**Solution with worker pool:**

```
10,000 files → only 10 workers running at any time
             → controlled, predictable resource usage
```

---

## The FileResult Struct

```go
type FileResult struct {
    Name      string
    WordCount int
    CharCount int
    Lines     int
}
```

Each processed file returns its stats. All results are collected into a slice after all goroutines finish.

---

## Mutex Protects Shared State

```go
var mu sync.Mutex

mu.Lock()
results = append(results, result)  // only one goroutine appends at a time
mu.Unlock()
```

Without the mutex, two goroutines could append at the same time → corrupted slice (race condition).

---

## Sequential vs Concurrent Speed

```
5 files, each takes 1 second to process:

Sequential:  1+1+1+1+1 = 5 seconds
Concurrent:  max(1,1,1,1,1) = 1 second   (all run at same time)
```

Concurrent is ~5x faster here. The more files, the bigger the gain.
