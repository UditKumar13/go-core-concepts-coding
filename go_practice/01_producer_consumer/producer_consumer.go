package main

import (
	"fmt"
	"sync"
)

// producer sends n jobs into jobs channel, then closes it.
func producer(jobs chan<- int, n int) {
	for i := 1; i <= n; i++ {
		jobs <- i
	}
	close(jobs)
}

// consumer reads jobs, doubles the value, sends result into results.
func consumer(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		result := job * 2
		results <- result
		fmt.Printf("Consumer %d processed job %d, result %d\n", id, job, result)
	}
}

func main() {
	const numJobs = 10
	const numWorkers = 3

	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go consumer(i, jobs, results, &wg)
	}

	producer(jobs, numJobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		fmt.Printf("Result: %d\n", result)
	}
}
