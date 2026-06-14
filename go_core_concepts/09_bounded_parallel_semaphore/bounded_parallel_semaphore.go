package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ---- The Semaphore ----
// A buffered channel of empty structs.
// Sending = acquire (blocks when full).
// Receiving = release (frees one slot).

type Semaphore chan struct{}

func NewSemaphore(n int) Semaphore {
	return make(chan struct{}, n)
}

func (s Semaphore) Acquire() {
	s <- struct{}{} // blocks if N slots already taken
}

func (s Semaphore) Release() {
	<-s // frees one slot, unblocks one waiting goroutine
}

// ---- Task processing ----

func processTask(id int, duration time.Duration, active *atomic.Int32) string {
	active.Add(1)
	defer active.Add(-1)

	fmt.Printf("  Task %02d started  | active goroutines: %d\n", id, active.Load())
	time.Sleep(duration) // simulate work
	result := fmt.Sprintf("task-%02d-done", id)
	fmt.Printf("  Task %02d finished | active goroutines: %d\n", id, active.Load())
	return result
}

func main() {
	const totalTasks    = 12
	const maxConcurrent = 3 // semaphore size — at most 3 tasks run at once

	sem := NewSemaphore(maxConcurrent)
	var wg sync.WaitGroup
	results := make([]string, totalTasks)

	// active tracks how many goroutines are inside processTask right now.
	// Used only for printing — proves the semaphore is working.
	var active atomic.Int32

	fmt.Printf("=== Bounded Parallelism: %d tasks, max %d concurrent ===\n\n", totalTasks, maxConcurrent)

	for i := 0; i < totalTasks; i++ {
		// Acquire BEFORE spawning — this is the critical rule.
		// If we acquired inside the goroutine, the goroutine is already
		// running by then — concurrency is not actually bounded.
		sem.Acquire()
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			defer sem.Release() // release slot when goroutine exits — always runs

			duration := time.Duration(100+id*20) * time.Millisecond
			results[id] = processTask(id+1, duration, &active)
		}(i)
	}

	wg.Wait()

	fmt.Printf("\n=== All tasks complete. Results ===\n")
	for _, r := range results {
		fmt.Println(" ", r)
	}
}
