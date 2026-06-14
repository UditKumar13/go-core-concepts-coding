package main

import (
	"fmt"
	"sync"
	"time"
)

// Fork is a mutex with an id so we can order them (resource hierarchy).
type Fork struct {
	id int
	sync.Mutex
}

// Philosopher holds references to their left and right fork.
type Philosopher struct {
	id        int
	leftFork  *Fork
	rightFork *Fork
}

func (p *Philosopher) think() {
	fmt.Printf("Philosopher %d is thinking\n", p.id)
	time.Sleep(time.Millisecond * 200)
}

func (p *Philosopher) eat() {
	fmt.Printf("Philosopher %d is eating\n", p.id)
	time.Sleep(time.Millisecond * 200)
}

// dine runs the think → pick forks → eat → put forks cycle N times.
// Resource hierarchy: always lock the lower-numbered fork first.
// This breaks the circular wait → deadlock is impossible.
func (p *Philosopher) dine(meals int, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < meals; i++ {
		p.think()

		// Determine pickup order by fork id — lower id always first
		first, second := p.leftFork, p.rightFork
		if p.rightFork.id < p.leftFork.id {
			first, second = p.rightFork, p.leftFork
		}

		first.Lock()
		fmt.Printf("Philosopher %d picked up fork %d\n", p.id, first.id)

		second.Lock()
		fmt.Printf("Philosopher %d picked up fork %d — now eating (meal %d)\n", p.id, second.id, i+1)

		p.eat()

		second.Unlock()
		first.Unlock()
		fmt.Printf("Philosopher %d put down forks\n", p.id)

		// Small pause between meals — prevents one philosopher
		// from immediately re-grabbing forks and starving others
		time.Sleep(time.Millisecond * 50)
	}

	fmt.Printf("Philosopher %d has finished all meals\n", p.id)
}

func main() {
	const numPhilosophers = 5
	const mealsEach = 3

	// Create 5 forks numbered 1–5
	forks := make([]*Fork, numPhilosophers)
	for i := 0; i < numPhilosophers; i++ {
		forks[i] = &Fork{id: i + 1}
	}

	// Create 5 philosophers.
	// Philosopher i sits between fork[i] (left) and fork[(i+1)%5] (right).
	philosophers := make([]*Philosopher, numPhilosophers)
	for i := 0; i < numPhilosophers; i++ {
		philosophers[i] = &Philosopher{
			id:        i + 1,
			leftFork:  forks[i],
			rightFork: forks[(i+1)%numPhilosophers],
		}
	}

	fmt.Printf("=== Dining Philosophers (resource hierarchy, %d meals each) ===\n\n", mealsEach)

	// Verify the ordering for each philosopher so it's visible
	for _, p := range philosophers {
		first, second := p.leftFork, p.rightFork
		if p.rightFork.id < p.leftFork.id {
			first, second = p.rightFork, p.leftFork
		}
		fmt.Printf("Philosopher %d: picks fork%d first, fork%d second\n", p.id, first.id, second.id)
	}
	fmt.Println()

	var wg sync.WaitGroup

	for _, p := range philosophers {
		wg.Add(1)
		go p.dine(mealsEach, &wg)
	}

	wg.Wait()
	fmt.Println("\nAll philosophers have finished dining. No deadlock!")
}
