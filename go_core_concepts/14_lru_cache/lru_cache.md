# LRU Cache in Go

## What is LRU Cache?

**LRU = Least Recently Used**

A cache with a fixed capacity. When it's full and a new item comes in, it **evicts the item that was used the longest time ago**.

```
Think of it like a desk:
- You can only keep 3 books on your desk
- When a 4th book arrives, you remove the one you haven't touched in the longest time
```

---

## How it Works

```
Capacity = 3

Put(a) → [a]
Put(b) → [b, a]
Put(c) → [c, b, a]       ← full

Get(a) → [a, c, b]       ← 'a' moved to front (recently used)

Put(d) → Evict 'b'       ← 'b' is LRU (at the back)
         [d, a, c]
```

---

## Data Structures Used

LRU Cache needs **two data structures** working together:

| Structure | Purpose |
|-----------|---------|
| `doubly linked list` | Tracks order (MRU at front, LRU at back) |
| `hashmap` | O(1) lookup by key |

### Why both?

- **HashMap alone** — O(1) lookup, but no way to track usage order
- **LinkedList alone** — tracks order, but O(n) lookup
- **Both together** — O(1) lookup AND O(1) order tracking

---

## The Entry Struct

```go
type entry struct {
    key   string
    value int
}
```

Stores both key and value inside the list node. We need the key so that when we **evict** from the back of the list, we can also delete it from the map.

---

## The LRUCache Struct

```go
type LRUCache struct {
    capacity int
    list     *list.List              // doubly linked list
    items    map[string]*list.Element // key → list node
}
```

- `list` — maintains MRU → LRU order
- `items` — maps key to its node in the list (for O(1) access)

---

## Core Operations

### Get(key)

```go
func (c *LRUCache) Get(key string) (int, bool) {
    el, ok := c.items[key]   // O(1) lookup in map
    if !ok {
        return 0, false      // Not found
    }
    c.list.MoveToFront(el)   // Mark as recently used
    return el.Value.(*entry).value, true
}
```

**Steps:**
1. Look up key in hashmap → get the list node
2. Move that node to the **front** of the list (most recently used)
3. Return the value

---

### Put(key, value)

```go
func (c *LRUCache) Put(key string, value int) {
    // Case 1: Key exists → update and move to front
    if el, ok := c.items[key]; ok {
        c.list.MoveToFront(el)
        el.Value.(*entry).value = value
        return
    }

    // Case 2: Cache full → evict LRU (back of list)
    if c.list.Len() == c.capacity {
        back := c.list.Back()
        c.list.Remove(back)
        delete(c.items, back.Value.(*entry).key)
    }

    // Case 3: Insert new entry at front
    el := c.list.PushFront(&entry{key, value})
    c.items[key] = el
}
```

**Three cases:**

| Case | Action |
|------|--------|
| Key already exists | Update value, move to front |
| Cache full | Remove back of list + delete from map |
| New key | Push to front of list + add to map |

---

## Visual Step-by-Step

```
Capacity = 3

Put("a",1)   list: [a]        map: {a→node}
Put("b",2)   list: [b,a]      map: {a,b→nodes}
Put("c",3)   list: [c,b,a]    map: {a,b,c→nodes}

Get("a")     list: [a,c,b]    ← 'a' moved to front
             map: {a,b,c→nodes}

Put("d",4)   ← full! evict back = "b"
             list: [d,a,c]    map: {a,c,d→nodes}

Get("b")     → not found (evicted)
```

---

## Time Complexity

| Operation | Time |
|-----------|------|
| Get | O(1) |
| Put | O(1) |
| Evict | O(1) |

All operations are O(1) because:
- HashMap gives O(1) lookup
- LinkedList gives O(1) insert/remove/move when you already have the node pointer

---

## Real World Uses

| System | LRU Cache Use |
|--------|---|
| Web browsers | Cache recently visited pages |
| CPU processors | L1/L2/L3 cache for memory |
| Redis | Eviction policy for memory management |
| DNS resolvers | Cache recently resolved domains |
| Database query cache | Cache recently run query results |
