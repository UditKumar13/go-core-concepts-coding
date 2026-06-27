package main

import (
	"fmt"
	"sync"
)

type Broker struct {
	mu          sync.RWMutex
	subscribers map[string][]chan string
}

func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string][]chan string),
	}
}

func (b *Broker) Subscribe(topic string) <-chan string {
	ch := make(chan string, 10)
	b.mu.Lock()
	b.subscribers[topic] = append(b.subscribers[topic], ch)
	b.mu.Unlock()
	return ch
}

func (b *Broker) Publish(topic, message string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers[topic] {
		ch <- message
	}
}

// broadcast message to all subscribers of a topic

func (b *Broker) Close(topic string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subscribers[topic] {
		close(ch)
	}
	delete(b.subscribers, topic)

	// delete the topic entry from the subscribers map to free up memory
}

func main() {
	broker := NewBroker()

	ch1 := broker.Subscribe("news")
	ch2 := broker.Subscribe("news")
	ch3 := broker.Subscribe("sports")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range ch1 {
			fmt.Println("[Subscriber 1 - news]:", msg)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range ch2 {
			fmt.Println("[Subscriber 2 - news]:", msg)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range ch3 {
			fmt.Println("[Subscriber 3 - sports]:", msg)
		}
	}()

	broker.Publish("news", "Go 1.23 Released!")
	broker.Publish("news", "Generics improved in Go")
	broker.Publish("sports", "India won the match!")
	broker.Publish("sports", "Olympics 2026 schedule announced")

	broker.Close("news")
	broker.Close("sports")

	wg.Wait()
}
