package main

import (
	"fmt"
	"sync"
)

// ---- Approach 1: Mutex guarded map ----
// Every read AND write acquires the same exclusive lock.
// Simple but reads block each other.

type MutexMap struct {
	mu sync.Mutex
	m  map[string]int
}

func NewMutexMap() *MutexMap {
	return &MutexMap{m: make(map[string]int)}
}

func (s *MutexMap) Set(key string, val int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = val
}

func (s *MutexMap) Get(key string) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.m[key]
	return val, ok
}

// ---- Approach 2: RWMutex guarded map ----
// Reads use RLock — many goroutines can read simultaneously.
// Writes use Lock  — exclusive, blocks all readers and writers.
// Best when reads >> writes.

type RWMap struct {
	mu sync.RWMutex
	m  map[string]int
}

func NewRWMap() *RWMap {
	return &RWMap{m: make(map[string]int)}
}

func (s *RWMap) Set(key string, val int) {
	s.mu.Lock() // exclusive write lock
	defer s.mu.Unlock()
	s.m[key] = val
}

func (s *RWMap) Get(key string) (int, bool) {
	s.mu.RLock() // shared read lock — multiple goroutines can hold this at once
	defer s.mu.RUnlock()
	val, ok := s.m[key]
	return val, ok
}

// Increment is a compound operation: read → compute → write.
// Must hold the write lock for the entire operation, not just each step.
func (s *RWMap) Increment(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key]++ // read + write inside one lock — safe
}

// ---- Approach 3: sync.Map ----
// Built into Go stdlib. No struct needed.
// Best for write-once / read-many or disjoint key access patterns.

func demonstrateSyncMap() {
	var m sync.Map

	m.Store("hits", 100)
	m.Store("errors", 5)
	m.Store("latency", 42)

	// LoadOrStore: returns existing value if key present, stores and returns new value if not
	val, loaded := m.LoadOrStore("hits", 999)
	fmt.Printf("  LoadOrStore hits → val=%v, was already loaded=%v\n", val, loaded)

	val, loaded = m.LoadOrStore("new_key", 777)
	fmt.Printf("  LoadOrStore new_key → val=%v, was already loaded=%v\n", val, loaded)

	// Range: iterate all key-value pairs
	fmt.Println("  All keys in sync.Map:")
	m.Range(func(key, value any) bool {
		fmt.Printf("    %s = %v\n", key, value)
		return true // return false to stop early
	})
}

func main() {

	// ── Demo 1: MutexMap under concurrent load ──────────────────────────────
	fmt.Println("=== Approach 1: Mutex Map ===")
	mm := NewMutexMap()
	var wg sync.WaitGroup

	// 5 writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			mm.Set(key, i*10)
		}(i)
	}

	// 5 readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			val, ok := mm.Get(key)
			if ok {
				fmt.Printf("  MutexMap got %s = %d\n", key, val)
			}
		}(i)
	}

	wg.Wait()

	// ── Demo 2: RWMap with concurrent reads + increments ────────────────────
	fmt.Println("\n=== Approach 2: RWMutex Map ===")
	rw := NewRWMap()
	rw.Set("counter", 0)

	// 10 goroutines all incrementing the same key — compound op, safe
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rw.Increment("counter")
		}()
	}
	wg.Wait()

	val, _ := rw.Get("counter")
	fmt.Printf("  RWMap counter after 10 increments = %d (expected 10)\n", val)

	// ── Demo 3: sync.Map ────────────────────────────────────────────────────
	fmt.Println("\n=== Approach 3: sync.Map ===")
	demonstrateSyncMap()
}
