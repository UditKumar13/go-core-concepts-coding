package main

import (
	"fmt"
)

func evenGenerator(n int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 0; i <= n; i += 2 {
			ch <- i
		}
	}()
	return ch
}

func oddGenerator(n int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 1; i <= n; i += 2 {
			ch <- i
		}
	}()
	return ch
}

// merge interleaves two sorted int channels into one sorted channel
func merge(evens, odds <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		e, eOpen := <-evens
		o, oOpen := <-odds
		for eOpen || oOpen {
			switch {
			case !eOpen:
				out <- o
				o, oOpen = <-odds
			case !oOpen:
				out <- e
				e, eOpen = <-evens
			case e <= o:
				out <- e
				e, eOpen = <-evens
			default:
				out <- o
				o, oOpen = <-odds
			}
		}
	}()
	return out
}

func main() {
	limit := 10

	evens := evenGenerator(limit)
	odds := oddGenerator(limit)

	for v := range merge(evens, odds) {
		fmt.Println(v)
	}
}
