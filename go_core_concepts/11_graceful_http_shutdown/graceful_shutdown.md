# Graceful HTTP Shutdown in Go

## What is Graceful Shutdown?

When you stop a server by pressing `Ctrl+C`, two things can happen:

| Type | What happens |
|------|---|
| **Abrupt shutdown** | Server dies immediately, in-flight requests are dropped |
| **Graceful shutdown** | Server stops accepting new requests, waits for active ones to finish, then exits |

Graceful shutdown avoids cutting off users mid-request.

---

## Flow

```
1. Server starts on :8080
2. Ctrl+C pressed  →  OS sends SIGINT signal
3. Signal received →  Stop accepting new requests
4. Wait up to 10s  →  Let in-flight requests finish
5. Exit cleanly
```

---

## Key Components

### 1. OS Signal Channel

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
```

- `SIGINT`  — Ctrl+C from terminal
- `SIGTERM` — kill command from system/container (Docker, Kubernetes)
- `signal.Notify` wires these OS signals into the `quit` channel

---

### 2. Server in a Goroutine

```go
go func() {
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        fmt.Println("Server error:", err)
    }
}()
```

Server runs in the background so `main()` can block on the signal.
`http.ErrServerClosed` is expected when Shutdown is called — we ignore it.

---

### 3. Block Until Signal

```go
<-quit
```

`main()` is blocked here waiting for `Ctrl+C` or `SIGTERM`.
Once a signal arrives, execution continues to the shutdown logic.

---

### 4. Graceful Shutdown with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

server.Shutdown(ctx)
```

- `server.Shutdown(ctx)` stops accepting new requests immediately
- Waits for active requests to finish
- If they don't finish in **10 seconds**, the context times out and forces exit

---

## Timeline Example

```
T=0s   Ctrl+C pressed
T=0s   Server stops accepting new requests
T=0s   3 requests still in-flight (each takes 3s)
T=3s   All 3 requests complete
T=3s   server.Shutdown() returns nil (clean exit)
T=3s   Program exits
```

If requests took longer than 10s:
```
T=0s   Ctrl+C pressed
T=10s  Context timeout → Shutdown() returns error
T=10s  Program exits (requests were cut off after grace period)
```

---

## Why Not Just `os.Exit()`?

```go
// BAD - kills everything instantly
<-quit
os.Exit(0)

// GOOD - waits for in-flight requests
<-quit
server.Shutdown(ctx)
```

`os.Exit()` is an abrupt shutdown — like pulling the power plug.
`server.Shutdown()` is a graceful shutdown — like flipping the off switch properly.

---

## Real World Use Cases

| Scenario | Why graceful shutdown matters |
|----------|---|
| Kubernetes pod restart | Pod gets SIGTERM — must finish requests before dying |
| Deployment rollout | Old server must drain before new one takes over |
| Docker `docker stop` | Sends SIGTERM — graceful handler prevents dropped requests |
| Payment processing | A payment mid-flight must not be interrupted |
