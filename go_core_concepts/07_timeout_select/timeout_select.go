package main

import (
	"context"
	"fmt"
	"time"
)

// slowWork simulates an operation that takes a variable amount of time.
// Returns result on the channel when done.
func slowWork(duration time.Duration) <-chan string {
	out := make(chan string, 1)
	go func() {
		time.Sleep(duration)
		out <- fmt.Sprintf("work done after %v", duration)
	}()
	return out
}

// ---- Demo 1: time.After timeout ----
// Simplest timeout — race work channel against time.After channel.
// Whoever sends first wins.
func demoTimeAfter() {
	fmt.Println("=== Demo 1: time.After Timeout ===")

	// Case A: work finishes before timeout
	fmt.Println("\n[A] Work (1s) vs Timeout (3s):")
	select {
	case result := <-slowWork(1 * time.Second):
		fmt.Println("  ✓ Got result:", result)
	case <-time.After(3 * time.Second):
		fmt.Println("  ✗ Timed out")
	}

	// Case B: timeout fires before work finishes
	fmt.Println("\n[B] Work (5s) vs Timeout (2s):")
	select {
	case result := <-slowWork(5 * time.Second):
		fmt.Println("  ✓ Got result:", result)
	case <-time.After(2 * time.Second):
		fmt.Println("  ✗ Timed out — work was too slow")
	}
}

// ---- Demo 2: select with default (non-blocking) ----
// default runs immediately if no channel is ready.
// Used for polling without blocking.
func demoDefault() {
	fmt.Println("\n=== Demo 2: Select with Default (Non-Blocking) ===")

	ch := make(chan int, 1)

	// Check 1: channel is empty — default fires
	select {
	case val := <-ch:
		fmt.Println("  got:", val)
	default:
		fmt.Println("  nothing ready — default fired")
	}

	// Send a value, then check again
	ch <- 42

	// Check 2: channel has a value — case fires
	select {
	case val := <-ch:
		fmt.Println("  got:", val)
	default:
		fmt.Println("  nothing ready — default fired")
	}
}

// ---- Demo 3: for + select event loop ----
// Continuously handles multiple channels until quit signal.
// This is the standard long-running goroutine pattern in Go.
func demoEventLoop() {
	fmt.Println("\n=== Demo 3: For + Select Event Loop ===")

	jobs := make(chan int, 5)
	quit := make(chan struct{})

	// send 3 jobs then signal quit
	go func() {
		for i := 1; i <= 3; i++ {
			jobs <- i
			time.Sleep(200 * time.Millisecond)
		}
		quit <- struct{}{}
	}()

	for {
		select {
		case job := <-jobs:
			fmt.Printf("  processing job %d\n", job)

		case <-quit:
			fmt.Println("  quit signal received — stopping event loop")
			return

		case <-time.After(1 * time.Second):
			// idle timeout — no job arrived for 1 second
			fmt.Println("  idle timeout — no jobs for 1s")
			return
		}
	}
}

// ---- Demo 4: context.WithTimeout ----
// Production pattern. Context propagates through function calls.
// defer cancel() frees resources even if work finishes early.
func demoContext() {
	fmt.Println("\n=== Demo 4: context.WithTimeout ===")

	// Case A: work finishes within deadline
	fmt.Println("\n[A] Work (500ms) within deadline (2s):")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	select {
	case result := <-slowWork(500 * time.Millisecond):
		fmt.Println("  ✓ Got result:", result)
	case <-ctx.Done():
		fmt.Println("  ✗ Context cancelled:", ctx.Err())
	}

	// Case B: deadline expires before work finishes
	fmt.Println("\n[B] Work (5s) exceeds deadline (1s):")
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()

	select {
	case result := <-slowWork(5 * time.Second):
		fmt.Println("  ✓ Got result:", result)
	case <-ctx2.Done():
		// ctx.Err() tells you WHY it was cancelled:
		// context.DeadlineExceeded → timeout fired
		// context.Canceled        → cancel() was called manually
		fmt.Println("  ✗ Context cancelled:", ctx2.Err())
	}
}

func main() {
	demoTimeAfter()
	demoDefault()
	demoEventLoop()
	demoContext()
}
