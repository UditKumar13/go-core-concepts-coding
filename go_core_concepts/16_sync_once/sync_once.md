# sync.Once in Go

## What is sync.Once?

`sync.Once` guarantees that a function runs **exactly once**, no matter how many goroutines call it simultaneously.

```go
var once sync.Once

once.Do(func() {
    fmt.Println("This runs only once")
})
```

Even if 100 goroutines call `once.Do(...)` at the same time, the function inside executes **exactly one time**.

---

## The Problem It Solves

### Without sync.Once (race condition):

```go
var db *Database

func GetDB() *Database {
    if db == nil {                     // ← Multiple goroutines can pass this check
        db = &Database{name: "Postgres"} // ← Multiple goroutines create DB!
    }
    return db
}
```

If 5 goroutines call `GetDB()` at the same time, all 5 might see `db == nil` and create 5 separate DB connections. That's a bug.

### With sync.Once (safe):

```go
var once sync.Once
var db *Database

func GetDB() *Database {
    once.Do(func() {
        db = &Database{name: "Postgres"} // ← Only 1 goroutine runs this
    })
    return dbp[-5]
}
```

Guaranteed: only one DB connection, no matter how many goroutines call `GetDB()`.

---

## Proof from Output

```
=== Example 3: Proving Same Pointer Returned ===
db1 pointer: 0xc000184000
db2 pointer: 0xc000184000
db3 pointer: 0xc000184000
All same? true
```

All three calls return the exact same pointer — same object in memory.

---

## How It Works Internally

`sync.Once` uses an atomic flag + mutex under the hood:

```
First call:
  → flag = 0 (not done)
  → acquire mutex
  → run the function
  → set flag = 1 (done)
  → release mutex

All subsequent calls:
  → flag = 1 (done)
  → skip the function immediately
```

---

## Common Use Cases

| Use Case | Why sync.Once? |
|----------|---|
| DB connection | Create once, reuse everywhere |
| Config loading | Load file once, not on every request |
| Logger setup | Initialize logger one time |
| Cache warm-up | Populate cache once at startup |
| Singleton pattern | Any resource that must be created once |

---

## sync.Once vs init()

| | `sync.Once` | `init()` |
|---|---|---|
| When it runs | On first call (lazy) | At program startup (eager) |
| Can be conditional | Yes | No |
| Thread safe | Yes | Yes |
| Use case | Lazy singletons | Always-needed setup |

Use `sync.Once` when initialization is **expensive** and you only want it if it's actually needed.
