# Concurrent Safe Map in Go

## The Problem — Why Plain Maps Are Dangerous

Go's built-in `map` is **not safe for concurrent use**. If two goroutines read and write a map at the same time, the program will crash with:

```
fatal error: concurrent map read and map write
```

This is not a compile error — it crashes at runtime, often in production.

```go
// DANGEROUS — do not do this
m := make(map[string]int)

go func() { m["a"] = 1 }()  // writer goroutine
go func() { fmt.Println(m["a"]) }()  // reader goroutine
// → fatal error: concurrent map read and map write
```

---

## Why Does This Happen?

A map in Go is a complex data structure (hash table with buckets). When two goroutines touch it simultaneously:
- One goroutine may be resizing the map while another reads
- Bucket pointers get corrupted mid-operation
- Go's race detector catches this even before the crash

Go **intentionally** made maps non-concurrent to keep them fast for single-goroutine use. You explicitly opt into concurrency safety.

---

## Three Solutions in Go

| Approach | Tool | Best For |
|---|---|---|
| 1. Mutex guard | `sync.Mutex` | General use, full control |
| 2. RWMutex guard | `sync.RWMutex` | Read-heavy workloads |
| 3. sync.Map | `sync.Map` (stdlib) | Mostly-read, mostly-disjoint keys |

---

## Solution 1 — sync.Mutex

Wrap the map in a struct with a `Mutex`. Lock before every read or write, unlock after.

```go
type SafeMap struct {
    mu sync.Mutex
    m  map[string]int
}

func (s *SafeMap) Set(key string, val int) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.m[key] = val
}

func (s *SafeMap) Get(key string) (int, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    val, ok := s.m[key]
    return val, ok
}
```

**Problem**: Even reads block each other. If 100 goroutines want to read and none are writing, they still queue up one at a time.

---

## Solution 2 — sync.RWMutex (Better for Reads)

`RWMutex` has two lock modes:
- `RLock()` / `RUnlock()` — shared read lock. Multiple readers allowed simultaneously.
- `Lock()` / `Unlock()` — exclusive write lock. Blocks everyone else.

```go
type SafeMap struct {
    mu sync.RWMutex
    m  map[string]int
}

func (s *SafeMap) Set(key string, val int) {
    s.mu.Lock()         // exclusive — no readers or writers
    defer s.mu.Unlock()
    s.m[key] = val
}

func (s *SafeMap) Get(key string) (int, bool) {
    s.mu.RLock()        // shared — many readers can hold this at once
    defer s.mu.RUnlock()
    val, ok := s.m[key]
    return val, ok
}
```

**Rule**: Use `RWMutex` when reads >> writes. If writes are frequent, RWMutex is no better than Mutex.

---

## Solution 3 — sync.Map (stdlib)

Go's standard library provides `sync.Map` — a concurrent map built-in.

```go
var m sync.Map

// Store
m.Store("key", 42)

// Load
val, ok := m.Load("key")

// Delete
m.Delete("key")

// LoadOrStore — atomic: load if exists, store if not
actual, loaded := m.LoadOrStore("key", 99)

// Range — iterate over all key-value pairs
m.Range(func(key, value any) bool {
    fmt.Println(key, value)
    return true // return false to stop iteration
})
```

**No lock needed** — sync.Map handles it internally.

**When to use sync.Map**:
- Keys are written once, read many times
- Different goroutines work on disjoint sets of keys (no key contention)
- You don't need custom logic around the lock

**When NOT to use sync.Map**:
- You need `len(m)` — sync.Map has no Len() method
- You need to do compound operations atomically (read-then-write)
- Writes are as frequent as reads — sync.Map is slower than RWMutex in this case

---

## Mutex vs RWMutex vs sync.Map

| | Mutex | RWMutex | sync.Map |
|---|---|---|---|
| Concurrent reads | No (serialized) | Yes (parallel) | Yes |
| Concurrent writes | No | No | Yes (different keys) |
| Compound ops | Easy (just hold lock) | Easy | Hard (no lock to hold) |
| len() support | Yes | Yes | No |
| Simplicity | Simple | Simple | Different API |
| Read-heavy perf | OK | Great | Great |
| Write-heavy perf | OK | Same as Mutex | Worse |

---

## The Lock — Unlock Contract

**Always use `defer` to unlock.** If you forget unlock, every other goroutine hangs forever.

```go
// WRONG — if function panics, lock is never released
s.mu.Lock()
s.m[key] = val
s.mu.Unlock()

// CORRECT — defer guarantees unlock even on panic
s.mu.Lock()
defer s.mu.Unlock()
s.m[key] = val
```

---

## Compound Operations — The Critical Concept

A compound operation is **read + decide + write**. You must hold the lock for the entire operation, not just each step individually.

```go
// WRONG — race condition between RUnlock and Lock
s.mu.RLock()
val, ok := s.m[key]   // read
s.mu.RUnlock()
if !ok {
    s.mu.Lock()
    s.m[key] = 1      // write — but another goroutine may have written between!
    s.mu.Unlock()
}

// CORRECT — hold write lock for the entire check-then-set
s.mu.Lock()
defer s.mu.Unlock()
if _, ok := s.m[key]; !ok {
    s.m[key] = 1
}
```

sync.Map solves this specific case with `LoadOrStore`, but for custom logic you need Mutex.

---

## Race Detector

Go ships with a built-in race detector. Run it to catch concurrent map access bugs:

```bash
go run -race main.go
go test -race ./...
```

Output when a race is detected:
```
WARNING: DATA RACE
Write at 0x... by goroutine 7:
  main.main.func1()
Read at 0x... by goroutine 8:
  main.main.func2()
```

Always run with `-race` during development and in CI.

---

## Lifecycle of a SafeMap Operation

```
Goroutine A (writer)          Goroutine B (reader)
    mu.Lock()                     mu.RLock() ← blocks, A holds write lock
    m["x"] = 10
    mu.Unlock()
                                  mu.RLock() ← now succeeds
                                  val := m["x"]  → 10
                                  mu.RUnlock()
```

---

## Common Mistakes

1. **Using plain map across goroutines** — runtime panic, often intermittent
2. **Forgetting defer Unlock** — deadlock if function returns early or panics
3. **Locking in caller instead of inside the safe type** — leaks the lock, easy to misuse
4. **Using RLock for writes** — corrupts the map silently (two "readers" both write)
5. **Calling Lock inside a Lock** — deadlock (Go's Mutex is not reentrant)
6. **Reading map length on sync.Map** — no Len() method, must maintain a separate counter

---

## Real-World Uses

- **Cache**: storing computed results shared across goroutines
- **Session store**: HTTP server storing user sessions (many readers, occasional writes)
- **Registry**: service discovery, plugin registry (write once at startup, read constantly)
- **Counters / metrics**: tracking request counts per endpoint
- **Rate limiting per key**: track last request time per user ID
