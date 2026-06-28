package main

import (
	"container/list"
	"fmt"
)

type entry struct {
	key   string
	value int
}

type LRUCache struct {
	capacity int
	list     *list.List
	items    map[string]*list.Element
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		list:     list.New(),
		items:    make(map[string]*list.Element),
	}
}

func (c *LRUCache) Get(key string) (int, bool) {
	el, ok := c.items[key]
	if !ok {
		return 0, false
	}
	// Move to front — most recently used
	c.list.MoveToFront(el)
	return el.Value.(*entry).value, true
}

func (c *LRUCache) Put(key string, value int) {
	// If key exists, update value and move to front
	if el, ok := c.items[key]; ok {
		c.list.MoveToFront(el)
		el.Value.(*entry).value = value
		return
	}

	// Evict least recently used if at capacity
	if c.list.Len() == c.capacity {
		back := c.list.Back()
		if back != nil {
			c.list.Remove(back)
			delete(c.items, back.Value.(*entry).key)
			fmt.Printf("Evicted: %q\n", back.Value.(*entry).key)
		}
	}

	// Insert new entry at front
	el := c.list.PushFront(&entry{key, value})
	c.items[key] = el
}

func (c *LRUCache) Print() {
	fmt.Print("Cache (MRU → LRU): ")
	for el := c.list.Front(); el != nil; el = el.Next() {
		e := el.Value.(*entry)
		fmt.Printf("[%s:%d] ", e.key, e.value)
	}
	fmt.Println()
}

func main() {
	fmt.Println("=== LRU Cache (capacity=3) ===")
	cache := NewLRUCache(3)

	cache.Put("a", 1)
	cache.Put("b", 2)
	cache.Put("c", 3)
	cache.Print()

	fmt.Println("\nGet 'a' (moves it to front):")
	val, ok := cache.Get("a")
	fmt.Printf("Get('a') = %d, found=%v\n", val, ok)
	cache.Print()

	fmt.Println("\nPut 'd' → cache full, evict LRU:")
	cache.Put("d", 4)
	cache.Print()

	fmt.Println("\nGet 'b' → was evicted:")
	val, ok = cache.Get("b")
	fmt.Printf("Get('b') = %d, found=%v\n", val, ok)

	fmt.Println("\nUpdate existing key 'c':")
	cache.Put("c", 99)
	cache.Print()
}
