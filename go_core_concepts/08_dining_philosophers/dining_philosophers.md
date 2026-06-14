# Dining Philosophers Problem in Go

## The Classic Problem

Five philosophers sit around a circular table. Between each pair of philosophers is one fork. To eat, a philosopher needs **both** the left and right fork. After eating, they put both forks down and think.

```
        Philosopher 1
       /             \
  Fork 5             Fork 1
     /                   \
Philosopher 5         Philosopher 2
     \                   /
  Fork 4             Fork 2
       \             /
        Philosopher 4 — Fork 3 — Philosopher 3
```

**The challenge**: How do philosophers share forks without:
- **Deadlock** — everyone waits forever (each holds one fork, waiting for the other)
- **Starvation** — someone never gets to eat
- **Race condition** — two philosophers grab the same fork simultaneously

---

## Why This Problem Matters

Dining Philosophers is a model for any system where:
- Multiple goroutines compete for multiple shared resources
- Each goroutine needs more than one resource simultaneously
- Resources are exclusive (only one holder at a time)

Real examples:
- Database transactions locking multiple rows
- Goroutines acquiring multiple mutexes
- Threads needing two network connections

---

## The Deadlock Scenario

If every philosopher picks up their LEFT fork first, then waits for their RIGHT fork:

```
t=0: P1 picks fork1 (left),  waiting for fork2
t=0: P2 picks fork2 (left),  waiting for fork3
t=0: P3 picks fork3 (left),  waiting for fork4
t=0: P4 picks fork4 (left),  waiting for fork5
t=0: P5 picks fork5 (left),  waiting for fork1

→ Everyone is waiting. Nobody can proceed. DEADLOCK.
```

This is a **circular wait** — the classic deadlock condition.

---

## The Four Conditions for Deadlock

All four must be true simultaneously for deadlock to occur:

| Condition | Meaning |
|---|---|
| Mutual Exclusion | A fork can only be held by one philosopher |
| Hold and Wait | Holding one fork while waiting for another |
| No Preemption | Cannot take a fork from another philosopher |
| Circular Wait | P1 waits for P2, P2 waits for P3, ... P5 waits for P1 |

Break **any one** of these and deadlock is impossible.

---

## Solutions

### Solution 1 — Resource Hierarchy (Break Circular Wait)

Number the forks 1–5. Every philosopher always picks up the **lower-numbered fork first**.

```
P1: picks fork1 then fork2  (1 < 2 ✓)
P2: picks fork2 then fork3  (2 < 3 ✓)
P3: picks fork3 then fork4  (3 < 4 ✓)
P4: picks fork4 then fork5  (4 < 5 ✓)
P5: picks fork1 then fork5  (1 < 5 ✓) ← P5 picks RIGHT fork first!
```

P5 breaks the circle. Now P1 and P5 both want fork1 — one gets it, the other waits. No circular wait → no deadlock.

**This is the simplest fix. We will implement this.**

---

### Solution 2 — Allow Only N-1 Philosophers at the Table

If only 4 of the 5 philosophers can sit at once, deadlock is impossible — at least one philosopher can always get both forks.

Use a semaphore (buffered channel) to limit how many sit simultaneously:

```go
seats := make(chan struct{}, 4) // only 4 allowed at table
```

---

### Solution 3 — Pick Up Both Forks Atomically

A philosopher either gets both forks at once or gets neither. Requires a central arbiter (mutex protecting the pick-up action).

More complex, lower concurrency.

---

## Mutex as a Fork in Go

Each fork is a `sync.Mutex`. Locking = picking up the fork. Unlocking = putting it down.

```go
type Fork struct {
    sync.Mutex
}

// pick up fork = lock the mutex
fork.Lock()

// put down fork = unlock the mutex
fork.Unlock()
```

Only one philosopher can lock (hold) a fork at a time — mutual exclusion guaranteed.

---

## Philosopher Lifecycle

```
for {
    think()          ← not holding any fork, safe, concurrent
    pick up fork A   ← lock forkA.Lock()
    pick up fork B   ← lock forkB.Lock()
    eat()            ← holds both forks
    put down fork B  ← forkB.Unlock()
    put down fork A  ← forkA.Unlock()
}
```

---

## Resource Hierarchy Implementation Plan

```go
type Philosopher struct {
    id        int
    leftFork  *Fork
    rightFork *Fork
}

func (p *Philosopher) dine(wg *sync.WaitGroup) {
    defer wg.Done()
    for i := 0; i < 3; i++ { // eat 3 times
        think(p.id)

        // Always lock lower-numbered fork first — breaks circular wait
        first, second := p.leftFork, p.rightFork
        if p.rightFork.id < p.leftFork.id {
            first, second = p.rightFork, p.leftFork
        }

        first.Lock()
        second.Lock()

        eat(p.id)

        second.Unlock()
        first.Unlock()
    }
}
```

---

## What We Will See in Output

Without the fix (naive left-first): program may hang forever (deadlock).

With resource hierarchy fix:
```
Philosopher 1 is thinking
Philosopher 3 is thinking
Philosopher 1 picked up forks, now eating
Philosopher 2 is thinking
Philosopher 1 finished eating, put down forks
Philosopher 3 picked up forks, now eating
...
All philosophers have finished dining
```

Philosophers take turns, nobody starves, program always terminates.

---

## Starvation vs Deadlock

| | Deadlock | Starvation |

| Deadlock guarentees starvation but not the other way around | 
|---|---|---|
| What happens | Everyone stops forever | One goroutine never gets CPU/resource |
| Cause | Circular wait | Unfair scheduling, one always loses |
| Resource hierarchy fixes | Deadlock | Not necessarily starvation |
| Go's mutex is | Fair? | Go's mutex is NOT strictly fair — starvation possible under high contention |

Resource hierarchy solves deadlock. Starvation needs additional fairness mechanisms (not covered here).

---

## Common Mistakes

1. **Both philosophers pick same-side fork first** — circular wait → deadlock
2. **Forgetting to Unlock on early return** — use `defer fork.Unlock()` carefully (can't defer in a loop safely — unlock explicitly)
3. **Using one global mutex for all forks** — solves deadlock but only one philosopher eats at a time — no concurrency
4. **Not checking fork IDs for ordering** — resource hierarchy only works if ordering is consistent

---

## Real-World Parallel

| Dining Problem | Real System |
|---|---|
| Philosopher | Goroutine / Thread |
| Fork | Mutex / Lock / Resource |
| Eating | Critical section (using both resources) |
| Thinking | Non-critical work |
| Deadlock | Database lock cycle, goroutine deadlock |
| Resource hierarchy | Lock ordering convention in codebases |
