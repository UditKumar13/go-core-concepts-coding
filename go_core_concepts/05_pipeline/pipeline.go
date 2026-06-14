package main

import "fmt"

// --- Stage 0: Source ---
// generate is the entry point of the pipeline.
// Takes plain values, puts them into a channel, closes it, returns the channel.
func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// --- Stage 1: double ---
// Receives from upstream, multiplies each value by 2, sends downstream.
func double(input <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for val := range input {
			out <- val * 2
		}
	}()
	return out
}

// --- Stage 2: filterEven ---
// Receives from upstream, only passes through even numbers, drops odds.
func filterEven(input <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for val := range input {
			if val%2 == 0 {
				out <- val
			}
		}
	}()
	return out
}

// --- Stage 3: square ---
// Receives from upstream, squares each value, sends downstream.
func square(input <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for val := range input {
			out <- val * val
		}
	}()
	return out
}

func main() {
	// Chain stages together — data flows left to right.
	// generate → double → filterEven → square → sink (main)
	//
	// Each function returns a channel that the next stage reads from.
	// Closing cascades automatically: generate closes → double closes → filterEven closes → square closes → main exits.

	pipeline := square(
		filterEven(
			double(
				generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
			),
		),
	)

	// --- Sink: collect all results from the final stage ---
	fmt.Println("Input: 1..10")
	fmt.Println("Stages: double → filterEven → square")
	fmt.Println()

	total := 0
	for result := range pipeline {
		fmt.Printf("result: %d\n", result)
		total += result
	}

	fmt.Printf("\nTotal = %d\n", total)
}
