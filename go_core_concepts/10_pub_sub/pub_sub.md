# Pub/Sub Pattern in Go

## What is Pub/Sub?

Pub/Sub (Publish / Subscribe) is a messaging pattern where:

- **Publisher** sends messages to a **topic** (does not know who is listening)
- **Subscriber** listens to a **topic** (does not know who is sending)
- A **Broker** sits in the middle and routes messages

```
Publisher ──→ Broker ──→ Subscriber 1
    (topic)         └──→ Subscriber 2
                    └──→ Subscriber 3
```

---

## Components

### 1. Broker

The central hub. It holds a map of `topic → list of subscriber channels`.

```go
type Broker struct {
    mu          sync.RWMutex
    subscribers map[string][]chan string
}
```

- `subscribers` maps a topic name to all channels that subscribed to it
- `sync.RWMutex` protects the map for concurrent reads and writes

---

### 2. Subscribe

```go
func (b *Broker) Subscribe(topic string) <-chan string {
    ch := make(chan string, 10)
    b.subscribers[topic] = append(b.subscribers[topic], ch)
    return ch
}
```

- Creates a **buffered channel** (size 10) for the subscriber
- Appends it to the broker's list for that topic
- Returns the channel so the subscriber can read from it

---

### 3. Publish

```go
func (b *Broker) Publish(topic, message string) {
    for _, ch := range b.subscribers[topic] {
        ch <- message
    }
}
```

- Looks up all subscribers for the topic
- Sends the message to **every** subscriber's channel
- Publisher doesn't care how many subscribers exist

---

### 4. Close

```go
func (b *Broker) Close(topic string) {
    for _, ch := range b.subscribers[topic] {
        close(ch)
    }
    delete(b.subscribers, topic)
}
```

- Closes all subscriber channels for the topic
- This signals the `for range` loops in subscribers to stop
- Cleans up the map entry

---

## Flow Example

```
broker.Subscribe("news")  → ch1
broker.Subscribe("news")  → ch2
broker.Subscribe("sports") → ch3

broker.Publish("news", "Go 1.23 Released!")
  → sends to ch1 and ch2

broker.Publish("sports", "India won!")
  → sends to ch3 only

broker.Close("news")
  → closes ch1 and ch2 → subscriber goroutines exit

broker.Close("sports")
  → closes ch3 → subscriber goroutine exits
```

---

## Why Buffered Channels?

```go
ch := make(chan string, 10)
```

Without a buffer, `Publish` would **block** until the subscriber reads the message.
With a buffer of 10, the publisher can send up to 10 messages without waiting.

---

## Concurrency Safety

`sync.RWMutex` is used because:

| Operation | Lock type |
|-----------|-----------|
| Subscribe / Close (write to map) | `Lock()` (exclusive) |
| Publish (read from map) | `RLock()` (shared) |

Multiple publishers can publish at the same time (shared read lock), but subscribing/closing is exclusive.

---

## Real World Uses

| Use Case | Example |
|----------|---------|
| Event systems | UI button click → multiple handlers |
| Microservices | Order placed → notify inventory + billing |
| Logging | Log event → multiple log destinations |
| Chat apps | Message sent → all room members receive it |
