package main

import (
	"fmt"
	"sync"
	"time"
)

// worker pulls jobs from jobs channel.
// Before processing each job it must receive a token from the shared limiter.
// This is the gate — no token = wait. One token = proceed.
func worker(id int, jobs <-chan int, results chan<- int, limiter <-chan time.Time, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		<-limiter // block here until a token arrives from the ticker

		start := time.Now()
		result := job * 10
		fmt.Printf("Worker %d | job %d | result %d | waited until %s\n",
			id, job, result, start.Format("15:04:05.000"))
		results <- result
	}
}

func main() {
	const numJobs = 8
	const numWorkers = 4
	const rate = 2 // jobs per second

	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	// --- Rate Limiter setup ---
	// interval = 1s / rate → one token every 500ms → 2 jobs per second
	interval := time.Second / time.Duration(rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop() // always stop the ticker — prevents goroutine leak

	fmt.Printf("Rate: %d jobs/sec | Interval: %v | Workers: %d | Jobs: %d\n\n",
		rate, interval, numWorkers, numJobs)

	var wg sync.WaitGroup

	// --- Spin up worker pool, all sharing the same limiter (ticker.C) ---
	// ticker.C is a single channel — only ONE worker gets each token.
	// This means the total rate across ALL workers = rate of the ticker.
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, results, ticker.C, &wg)
	}

	// --- Dispatch all jobs ---
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	// --- Close results after all workers finish ---
	go func() {
		wg.Wait()
		close(results)
	}()

	// --- Collect results ---
	total := 0
	for result := range results {
		total += result
	}

	fmt.Printf("\nAll done. Total = %d\n", total)
}
