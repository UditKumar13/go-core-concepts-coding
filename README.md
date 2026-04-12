# leetcode-go
DSA in Go • Java comparisons • built in public by a Java dev learning Go the hard way.
# 🐹 LeetCode in Go — With Java Comparisons

> A public learning repo by [Udit Kumar](https://github.com/UditKumar13) — SDE-2 @ Dell | M.Tech CS, IIIT Delhi  
> **Companion repo to:** [DSA-IN-JAVA](https://github.com/UditKumar13) · **Goal:** Go fluency for backend/systems roles

---

## 🧭 Why This Repo Exists

I already have a strong Java DSA foundation. This repo documents my journey of **re-learning the same patterns in Go** — with an explicit focus on:

- How Go **thinks differently** from Java
- Where Go is **cleaner, faster, or more idiomatic**
- Gotchas that **Java devs hit** when switching to Go

If you're a Java engineer learning Go, this repo is for you.

---

## 🗂️ Repo Structure

```
leetcode-go/
├── README.md
├── patterns/                    ← Read these FIRST before problems
│   ├── sliding-window.md
│   ├── binary-search.md
│   ├── two-pointers.md
│   ├── dfs-bfs.md
│   ├── dynamic-programming.md
│   └── java-vs-go-cheatsheet.md  ← Quick reference
├── arrays/
│   └── two-sum/
│       ├── solution.go
│       ├── solution_test.go
│       └── README.md
├── binary-search/
├── sliding-window/
├── trees/
├── graphs/
├── dynamic-programming/
└── tries/
```

Each problem folder has:
- `solution.go` — clean, idiomatic Go solution
- `solution_test.go` — table-driven tests
- `README.md` — problem + approach + **Java vs Go notes**

---

## ⚡ Java vs Go — Quick Reference for DSA

This is the cheatsheet I wish existed when I started.

### 1. Data Structures

| Need | Java | Go |
|---|---|---|
| Dynamic array | `ArrayList<Integer>` | `[]int` (slices) |
| Hash map | `HashMap<Integer, Integer>` | `map[int]int` |
| Hash set | `HashSet<Integer>` | `map[int]struct{}` |
| Stack | `Deque<Integer> stack = new ArrayDeque<>()` | `[]int` (slice as stack) |
| Queue | `Queue<Integer> q = new LinkedList<>()` | `[]int` or `container/list` |
| Min Heap | `PriorityQueue<Integer>` | `heap.Interface` (container/heap) |
| Max Heap | `PriorityQueue<Integer>(Collections.reverseOrder())` | `heap.Interface` (negate values) |

### 2. Syntax Patterns

**Iterating a slice/array:**
```go
// Go
for i, v := range nums {
    fmt.Println(i, v)
}
```
```java
// Java
for (int i = 0; i < nums.length; i++) {
    System.out.println(i + " " + nums[i]);
}
// or
for (int v : nums) { ... }
```

**Checking map membership:**
```go
// Go — idiomatic two-value assignment
if val, ok := myMap[key]; ok {
    // key exists, use val
}
```
```java
// Java
if (myMap.containsKey(key)) {
    int val = myMap.get(key);
}
```

**Sorting:**
```go
// Go
import "sort"
sort.Ints(nums)                          // ascending
sort.Sort(sort.Reverse(sort.IntSlice(nums))) // descending
sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] }) // custom
```
```java
// Java
Arrays.sort(nums);
Arrays.sort(nums, (a, b) -> b - a);     // descending
```

**Min / Max:**
```go
// Go — no built-in min/max before Go 1.21
// Go 1.21+
import "cmp"
m := min(a, b)  // built-in
m := max(a, b)

// Before 1.21 — write your own
func min(a, b int) int {
    if a < b { return a }
    return b
}
```
```java
// Java
int m = Math.min(a, b);
int m = Math.max(a, b);
```

**Integer overflow:**
```go
// Go — int is 64-bit on 64-bit systems, usually safe
// Use int64 explicitly when needed
var result int64 = int64(a) * int64(b)
```
```java
// Java — int is always 32-bit, be careful
long result = (long) a * b;
```

**String to int / int to string:**
```go
// Go
import "strconv"
n, err := strconv.Atoi("42")       // string → int
s := strconv.Itoa(42)              // int → string
```
```java
// Java
int n = Integer.parseInt("42");
String s = String.valueOf(42);
// or
String s = Integer.toString(42);
```

**Rune / char operations:**
```go
// Go — strings are UTF-8 bytes; use rune for chars
s := "hello"
for i, ch := range s {    // ch is rune (int32)
    fmt.Println(i, ch, string(ch))
}
freq := make([]int, 26)
freq[ch - 'a']++          // same pattern as Java
```
```java
// Java
char ch = s.charAt(i);
freq[ch - 'a']++;
```

### 3. Go-Specific Things Java Devs Must Know

| Concept | Java Mental Model | Go Reality |
|---|---|---|
| **`nil` vs `null`** | `null` causes NPE | `nil` slice/map is valid and usable |
| **Slices** | Array is fixed; ArrayList is dynamic | Slice is the default — backed by array, has `len` and `cap` |
| **No classes** | OOP with classes | Structs + methods + interfaces (implicit implementation) |
| **Error handling** | `try/catch` exceptions | Return `(value, error)` — check `err != nil` |
| **Pointers** | Mostly hidden | Explicit `*T` and `&val` — matters for tree/graph nodes |
| **Multiple return** | Return one value or use wrappers | `return val, err` natively |
| **`make` vs `new`** | `new ArrayList<>()` | `make([]int, n)` for slices/maps/channels; `new` rarely used |

---

## 📈 Progress Tracker

| Topic | Problems Solved | Pattern Doc | Status |
|---|---|---|---|
| Arrays & Strings | 0 | — | 🔲 Not Started |
| Binary Search | 0 | — | 🔲 Not Started |
| Sliding Window | 0 | — | 🔲 Not Started |
| Trees | 0 | — | 🔲 Not Started |
| Graphs | 0 | — | 🔲 Not Started |
| Dynamic Programming | 0 | — | 🔲 Not Started |
| Tries | 0 | — | 🔲 Not Started |

**Legend:** 🔲 Not Started · 🟡 In Progress · ✅ Completed

---

## 🧪 Running Solutions

```bash
# Run all tests
go test ./...

# Run tests for a specific problem
go test ./arrays/two-sum/

# Run with verbose output
go test -v ./arrays/two-sum/

# Check for race conditions
go test -race ./...

# Benchmark
go test -bench=. ./arrays/two-sum/
```

---

## 📐 Solution Template

Every solution in this repo follows this structure:

```go
// arrays/two-sum/solution.go
package twosum

// TwoSum returns indices of two numbers that add up to target.
// Time: O(n) | Space: O(n)
func TwoSum(nums []int, target int) []int {
    seen := make(map[int]int) // value → index
    for i, n := range nums {
        if j, ok := seen[target-n]; ok {
            return []int{j, i}
        }
        seen[n] = i
    }
    return nil
}
```

```go
// arrays/two-sum/solution_test.go
package twosum

import (
    "testing"
    "reflect"
)

func TestTwoSum(t *testing.T) {
    tests := []struct {
        name   string
        nums   []int
        target int
        want   []int
    }{
        {"basic", []int{2, 7, 11, 15}, 9, []int{0, 1}},
        {"duplicate values", []int{3, 3}, 6, []int{0, 1}},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got := TwoSum(tc.nums, tc.target)
            if !reflect.DeepEqual(got, tc.want) {
                t.Errorf("got %v, want %v", got, tc.want)
            }
        })
    }
}
```

---

## 🔗 Related Repos

| Repo | Description |
|---|---|
| [DSA-IN-JAVA](https://github.com/UditKumar13) | Original Java DSA solutions — the companion to this repo |
| [Go-Microservices-Project-using-Kubernetes](https://github.com/UditKumar13/Go-Microservices-Project-using-Kubernetes) | Production Go project |

---

## 📣 Follow the Journey

| Platform | What I post |
|---|---|
| [LinkedIn](https://linkedin.com/in/udit-kumar13) | Weekly problem breakdowns, Java→Go insights |
| [GitHub](https://github.com/UditKumar13) | All code, notes, pattern docs |
| Hashnode / dev.to | Long-form: "DSA Patterns That Actually Stick" series |

---

## 🤝 Contributing / Learning Together

Found a more idiomatic Go solution? Know a better Java comparison? Open a PR or raise an issue — learning is better in public.

---

> _"Loop: (Eat-Sleep-Code) -_-"_  
> — Udit Kumar
