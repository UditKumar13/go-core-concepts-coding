package main

import (
	"errors"
	"fmt"
)

type RingBuffer struct {
	data  []int
	head  int // next read position
	tail  int // next write position
	size  int // current number of elements
	cap   int // max capacity
}

func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		data: make([]int, capacity),
		cap:  capacity,
	}
}

func (r *RingBuffer) Write(val int) error {
	if r.size == r.cap {
		return errors.New("buffer full")
	}
	r.data[r.tail] = val
	r.tail = (r.tail + 1) % r.cap
	r.size++
	return nil
}

func (r *RingBuffer) Read() (int, error) {
	if r.size == 0 {
		return 0, errors.New("buffer empty")
	}
	val := r.data[r.head]
	r.head = (r.head + 1) % r.cap
	r.size--
	return val, nil
}

func (r *RingBuffer) IsFull() bool  { return r.size == r.cap }
func (r *RingBuffer) IsEmpty() bool { return r.size == 0 }
func (r *RingBuffer) Size() int     { return r.size }

func (r *RingBuffer) Print() {
	fmt.Printf("Buffer (head=%d tail=%d size=%d/%d): ", r.head, r.tail, r.size, r.cap)
	if r.size == 0 {
		fmt.Println("[]")
		return
	}
	fmt.Print("[")
	for i := 0; i < r.size; i++ {
		idx := (r.head + i) % r.cap
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(r.data[idx])
	}
	fmt.Println("]")
}

func main() {
	fmt.Println("=== Ring Buffer (capacity=5) ===")
	rb := NewRingBuffer(5)

	fmt.Println("\n--- Writing 1,2,3,4,5 ---")
	for i := 1; i <= 5; i++ {
		if err := rb.Write(i); err != nil {
			fmt.Printf("Write(%d): %v\n", i, err)
		} else {
			fmt.Printf("Write(%d) OK\n", i)
		}
	}
	rb.Print()

	fmt.Println("\n--- Writing 6 (buffer full) ---")
	if err := rb.Write(6); err != nil {
		fmt.Println("Write(6):", err)
	}

	fmt.Println("\n--- Reading 3 values ---")
	for i := 0; i < 3; i++ {
		val, _ := rb.Read()
		fmt.Printf("Read() = %d\n", val)
	}
	rb.Print()

	fmt.Println("\n--- Writing 6,7,8 (wraps around) ---")
	for _, v := range []int{6, 7, 8} {
		rb.Write(v)
		fmt.Printf("Write(%d)\n", v)
	}
	rb.Print()

	fmt.Println("\n--- Reading all remaining ---")
	for !rb.IsEmpty() {
		val, _ := rb.Read()
		fmt.Printf("Read() = %d\n", val)
	}
	rb.Print()

	fmt.Println("\n--- Reading from empty buffer ---")
	_, err := rb.Read()
	fmt.Println("Read():", err)
}
