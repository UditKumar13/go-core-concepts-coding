# Ring Buffer in Go

## What is a Ring Buffer?

A ring buffer (also called circular buffer) is a **fixed-size array that wraps around** — when you reach the end, writing continues from the beginning.

```
capacity = 5

[ 1 | 2 | 3 | 4 | 5 ]
  ↑                 ↑
 head              tail

Read 1,2,3 → head moves right
Write 6,7,8 → tail wraps to front

[ 6 | 7 | 8 | 4 | 5 ]
              ↑
             head
```

---

## Key Pointers

| Pointer | Purpose |
|---------|---------|
| `head` | Next position to **read** from |
| `tail` | Next position to **write** to |
| `size` | Current number of elements |

Both `head` and `tail` advance by 1 after each operation and **wrap around** using modulo:

```go
r.tail = (r.tail + 1) % r.cap  // wraps 5 → 0
r.head = (r.head + 1) % r.cap  // wraps 5 → 0
```

---

## Write Operation

```go
func (r *RingBuffer) Write(val int) error {
    if r.size == r.cap {
        return errors.New("buffer full")  // no overwrite
    }
    r.data[r.tail] = val
    r.tail = (r.tail + 1) % r.cap  // advance tail (wrap if needed)
    r.size++
    return nil
}
```

---

## Read Operation

```go
func (r *RingBuffer) Read() (int, error) {
    if r.size == 0 {
        return 0, errors.New("buffer empty")
    }
    val := r.data[r.head]
    r.head = (r.head + 1) % r.cap  // advance head (wrap if needed)
    r.size--
    return val, nil
}
```

---

## Step-by-Step Trace (from output)

```
Write 1,2,3,4,5:
  data: [1, 2, 3, 4, 5]  head=0  tail=0  size=5

Read 1,2,3:
  data: [1, 2, 3, 4, 5]  head=3  tail=0  size=2
  (data still in array, head just moved past it)

Write 6,7,8 (wraps!):
  tail was 0 → writes at index 0,1,2
  data: [6, 7, 8, 4, 5]  head=3  tail=3  size=5
  reads: [4, 5, 6, 7, 8] (starts from head=3, wraps around)
```

---

## Why head=tail=3 when Full AND Empty?

| State | head | tail | size |
|-------|------|------|------|
| Full | 3 | 3 | 5 |
| Empty | 3 | 3 | 0 |

We use `size` to distinguish — not head/tail position. Both states look the same positionally but `size` tells the truth.

---

## Time Complexity

| Operation | Time |
|-----------|------|
| Write | O(1) |
| Read | O(1) |
| Space | O(n) fixed |

No shifting, no resizing — just index arithmetic.

---

## Ring Buffer vs Regular Queue

| | Ring Buffer | Slice Queue |
|---|---|---|
| Size | Fixed | Grows dynamically |
| Memory | Reuses same array | Allocates more memory |
| Speed | O(1) always | O(1) amortized |
| Overflow | Returns error | Unlimited |

---

## Real World Uses

| System | Use Case |
|--------|---------|
| OS kernel | I/O buffers, pipe buffers |
| Audio/video | Streaming media buffers |
| Network drivers | Packet buffers |
| Log systems | Fixed-size rolling logs |
| Producer-Consumer | Decouple fast producer from slow consumer |
