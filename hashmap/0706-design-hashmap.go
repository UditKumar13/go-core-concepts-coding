package main

const N = 1000

type pair struct{ key, val int }

type MyHashMap struct {
	buckets [N][]pair
}

func Constructor() MyHashMap { return MyHashMap{} }

func (m *MyHashMap) hash(key int) int { return key % N }

func (m *MyHashMap) Put(key int, value int) {
	h := m.hash(key)
	for i := range m.buckets[h] {
		if m.buckets[h][i].key == key {
			m.buckets[h][i].val = value
			return
		}
	}
	m.buckets[h] = append(m.buckets[h], pair{key, value})
}

func (m *MyHashMap) Get(key int) int {
	for _, p := range m.buckets[m.hash(key)] {
		if p.key == key {
			return p.val
		}
	}
	return -1
}

func (m *MyHashMap) Remove(key int) {
	h := m.hash(key)
	b := m.buckets[h]
	for i, p := range b {
		if p.key == key {
			m.buckets[h] = append(b[:i], b[i+1:]...)
			return
		}
	}
}
