package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// Simulating API calls
func fetchUserData(id int) error {
	time.Sleep(2 * time.Second)
	if id == 2 {
		return fmt.Errorf("failed to fetch user %d", id)
	}
	fmt.Printf("Fetched user %d\n", id)
	return nil
}

func fetchPostData(id int) error {
	time.Sleep(1 * time.Second)
	fmt.Printf("Fetched post %d\n", id)
	return nil
}

func fetchCommentData(id int) error {
	time.Sleep(1 * time.Second)
	if id == 1 {
		return fmt.Errorf("failed to fetch comment %d", id)
	}
	fmt.Printf("Fetched comment %d\n", id)
	return nil
}

// Example 1: Using errgroup without context
func exampleWithoutContext() {
	fmt.Println("=== Example 1: Without Context ===")
	var eg errgroup.Group

	eg.Go(func() error {
		return fetchUserData(1)
	})

	eg.Go(func() error {
		return fetchPostData(1)
	})

	eg.Go(func() error {
		return fetchCommentData(2)
	})

	if err := eg.Wait(); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("All tasks completed successfully!")
	}
	fmt.Println()
}

// Example 2: Using errgroup with context and cancellation
func exampleWithContext() {
	fmt.Println("=== Example 2: With Context (Error Cancels Others) ===")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return fetchUserData(2) // This will fail
	})

	eg.Go(func() error {
		select {
		case <-egCtx.Done():
			fmt.Println("Post fetch cancelled due to error in user fetch")
			return egCtx.Err()
		case <-time.After(2 * time.Second):
			return fetchPostData(1)
		}
	})

	eg.Go(func() error {
		select {
		case <-egCtx.Done():
			fmt.Println("Comment fetch cancelled due to error")
			return egCtx.Err()
		case <-time.After(2 * time.Second):
			return fetchCommentData(1)
		}
	})

	if err := eg.Wait(); err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println()
}

// Example 3: Collecting results from goroutines
func exampleWithResults() {
	fmt.Println("=== Example 3: Collecting Results ===")
	var eg errgroup.Group
	results := make([]string, 0)
	var mu sync.Mutex

	for i := 1; i <= 3; i++ {
		id := i
		eg.Go(func() error {
			time.Sleep(time.Duration(id) * time.Second)
			result := fmt.Sprintf("Task %d completed", id)
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
			fmt.Println(result)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("All results:", results)
	fmt.Println()
}

// Example 4: Limiting concurrency
func exampleWithLimit() {
	fmt.Println("=== Example 4: Limiting Concurrency (Max 2 at a time) ===")
	var eg errgroup.Group
	eg.SetLimit(2) // Only 2 goroutines run concurrently

	for i := 1; i <= 5; i++ {
		id := i
		eg.Go(func() error {
			fmt.Printf("Starting task %d\n", id)
			time.Sleep(1 * time.Second)
			fmt.Printf("Finished task %d\n", id)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println()
}

func main() {
	exampleWithoutContext()
	exampleWithContext()
	exampleWithResults()
	exampleWithLimit()
}
