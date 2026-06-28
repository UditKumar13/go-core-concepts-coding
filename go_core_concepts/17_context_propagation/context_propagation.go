package main

import (
	"context"
	"fmt"
	"time"
)

// --- Example 1: Passing values through context ---

type contextKey string

const userIDKey contextKey = "userID"
const requestIDKey contextKey = "requestID"

func handleRequest(ctx context.Context) {
	userID := ctx.Value(userIDKey)
	requestID := ctx.Value(requestIDKey)
	fmt.Printf("  handleRequest: userID=%v requestID=%v\n", userID, requestID)
	processData(ctx)
}

func processData(ctx context.Context) {
	userID := ctx.Value(userIDKey)
	fmt.Printf("  processData: userID=%v (propagated from parent)\n", userID)
	saveToDB(ctx)
}

func saveToDB(ctx context.Context) {
	userID := ctx.Value(userIDKey)
	fmt.Printf("  saveToDB: userID=%v (propagated through all layers)\n", userID)
}

// --- Example 2: Timeout propagation ---

func fetchData(ctx context.Context, name string) error {
	select {
	case <-ctx.Done():
		fmt.Printf("  [%s] cancelled: %v\n", name, ctx.Err())
		return ctx.Err()
	case <-time.After(2 * time.Second):
		fmt.Printf("  [%s] completed successfully\n", name)
		return nil
	}
}

func orchestrate(ctx context.Context) {
	// Child context with shorter timeout
	childCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	fmt.Println("  Starting child operation with 1s timeout...")
	if err := fetchData(childCtx, "DatabaseFetch"); err != nil {
		fmt.Println("  Child failed:", err)
	}
}

// --- Example 3: Cancellation propagation (parent cancels all children) ---

func worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("  Worker %d stopped: %v\n", id, ctx.Err())
			return
		case <-time.After(500 * time.Millisecond):
			fmt.Printf("  Worker %d doing work...\n", id)
		}
	}
}

func main() {
	// Example 1: Value propagation
	fmt.Println("=== Example 1: Value Propagation Through Layers ===")
	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, "user-42")
	ctx = context.WithValue(ctx, requestIDKey, "req-abc123")
	handleRequest(ctx)

	fmt.Println()

	// Example 2: Timeout propagation
	fmt.Println("=== Example 2: Timeout Propagation ===")
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer parentCancel()
	orchestrate(parentCtx)

	fmt.Println()

	// Example 3: Cancellation propagation
	fmt.Println("=== Example 3: Parent Cancel Stops All Children ===")
	rootCtx, rootCancel := context.WithCancel(context.Background())

	for i := 1; i <= 3; i++ {
		go worker(rootCtx, i)
	}

	time.Sleep(1200 * time.Millisecond)
	fmt.Println("  Parent cancelling context...")
	rootCancel()
	time.Sleep(200 * time.Millisecond) // let goroutines print stop message
	fmt.Println("  Done.")
}
