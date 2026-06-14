package main

import (
	"fmt"
	"sync"
	"time"
)

// worker is one unit in the pool.
// It pulls jobs from the shared jobs channel until it is closed,
// simulates work with a sleep, then sends the result out.
func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done() // called when range exits (jobs channel closed + drained)

	for job := range jobs { // blocks waiting for next job; exits when channel closes
		fmt.Printf("Worker %d starting  job %d\n", id, job)
		time.Sleep(time.Millisecond * 100) // simulate work
		result := job * job                // actual computation: square the job number
		results <- result
		fmt.Printf("Worker %d finished job %d → result %d\n", id, job, result)
	}
}

func main() {
	const numJobs = 10
	const numWorkers = 3

	jobs := make(chan int, numJobs)    // buffered: dispatcher fills without blocking
	results := make(chan int, numJobs) // buffered: workers write without blocking

	var wg sync.WaitGroup

	// --- Step 1: spin up the fixed worker pool ---
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1) // MUST be before go, not inside the goroutine
		go worker(i, jobs, results, &wg)
	}

	// --- Step 2: dispatch all jobs ---
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs) // signals workers: no more jobs coming — their range loop will exit

	// --- Step 3: close results once ALL workers are done ---
	// This runs in a separate goroutine because wg.Wait() blocks.
	// If we called wg.Wait() here in main before ranging results,
	// we would deadlock: main waits for workers, workers can't finish
	// because results is full and nobody is draining it.
	go func() {
		wg.Wait()
		close(results)
	}()

	// --- Step 4: collect all results ---
	total := 0
	for result := range results {
		total += result
	}

	fmt.Printf("\nAll jobs done. Sum of squares 1..%d = %d\n", numJobs, total)
}
