package main

import (
	"fmt"
	"sync"
	"time"
)

// worker reads from the shared input channel, does work, sends to its OWN output channel.
// Every worker gets its own channel — this is the key difference from Worker Pool.
func worker(id int, input <-chan int) <-chan int {
	output := make(chan int)

	go func() {
		defer close(output) // worker owns its channel — it closes it when done
		for job := range input {
			time.Sleep(time.Millisecond * 100) // simulate work
			result := job * job
			fmt.Printf("Worker %d | job %d → result %d\n", id, job, result)
			output <- result
		}
	}()

	return output // caller gets a read-only handle to this worker's results
}

// fanOut spawns numWorkers workers, all reading from the same input channel.
// Returns a slice of output channels — one per worker.
func fanOut(input <-chan int, numWorkers int) []<-chan int {
	channels := make([]<-chan int, numWorkers)
	for i := 0; i < numWorkers; i++ {
		channels[i] = worker(i+1, input)
	}
	return channels
}

// fanIn merges all worker output channels into a single channel.
// Spawns one forwarder goroutine per input channel.
// Closes merged only after every forwarder finishes — safe, no panic.
func fanIn(channels ...<-chan int) <-chan int {
	merged := make(chan int)
	var wg sync.WaitGroup

	// forward drains one channel into merged
	forward := func(ch <-chan int) {
		defer wg.Done()
		for val := range ch {
			merged <- val
		}
	}

	wg.Add(len(channels))
	for _, ch := range channels {
		go forward(ch)
	}

	// closer: wait for all forwarders, then close merged so main's range exits
	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

func main() {
	const numJobs    = 9
	const numWorkers = 3

	// --- Step 1: create and fill the input channel ---
	input := make(chan int, numJobs)
	for j := 1; j <= numJobs; j++ {
		input <- j
	}
	close(input) // no more jobs — workers' range loops will exit when drained

	fmt.Printf("Jobs: %d | Workers: %d\n\n", numJobs, numWorkers)

	// --- Step 2: Fan-Out — spread input across N workers ---
	workerChannels := fanOut(input, numWorkers)

	// --- Step 3: Fan-In — merge all worker channels into one ---
	merged := fanIn(workerChannels...)

	// --- Step 4: collect results from the single merged channel ---
	total := 0
	for result := range merged {
		total += result
	}

	fmt.Printf("\nAll done. Sum of squares 1..%d = %d\n", numJobs, total)
}
