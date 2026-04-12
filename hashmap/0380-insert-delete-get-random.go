// 🧠 Interview Thinking Flow
// Step 1 — The constraint that changes everything
// All three operations — insert, delete, getRandom — must be O(1). That's the puzzle.

// HashMap alone handles insert/delete/lookup in O(1), but getRandom is O(n) — you can't index into a hashmap.
// Array alone handles getRandom in O(1), but delete is O(n) — you'd have to shift elements.

// Step 2 — Combine both
// Use a HashMap + ArrayList together:

// ArrayList stores the actual values → supports O(1) random access
// HashMap maps value → index in ArrayList → supports O(1) lookup

// Step 3 — The delete trick
// Deleting from the middle of an ArrayList is O(n). So instead:

// Swap the target element with the last element
// Remove the last element → O(1)
// Update the swapped element's index in the HashMap

package main

import "math/rand"

// Java remove from map
// map.remove(key)

// go remove from map
// delete(map, key)

// Remove last from list
// java
// list.remove(list.size()-1)
// go
// slice = slice[:len(slice)-1]
type RandomizedSet struct {
	indexMap map[int]int
	list     []int
}

func NewRandomizedSet() RandomizedSet {
	return RandomizedSet{
		indexMap: make(map[int]int),
		list:     []int{},
	}
}

func (rs *RandomizedSet) Insert(val int) bool {
	if _, exists := rs.indexMap[val]; exists {
		return false
	}
	rs.list = append(rs.list, val)
	rs.indexMap[val] = len(rs.list) - 1
	return true
}

func (rs *RandomizedSet) Remove(val int) bool {
	idx, exists := rs.indexMap[val]
	if !exists {
		return false
	}
	lastVal := rs.list[len(rs.list)-1]
	rs.list[idx] = lastVal
	rs.indexMap[lastVal] = idx
	rs.list = rs.list[:len(rs.list)-1]
	delete(rs.indexMap, val)
	return true
}

func (rs *RandomizedSet) GetRandom() int {
	return rs.list[rand.Intn(len(rs.list))]
}
