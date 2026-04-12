package main

import "fmt"

// Approach: HashMap (Single Pass)
// - For each num, check if (target - num) already exists in map
// - If yes → found our pair, return indices
// - If no  → store num with its index and move on
//
// TC: O(n) — single pass through the array
// SC: O(n) — storing up to n elements in the map
//
// Key Takeaway:
// Whenever you're searching for a specific value inside a loop,
// ask: "Can I trade space for time with a HashMap?"
// That's the O(n²) → O(n) upgrade pattern.

func twoSum(nums []int, target int) []int {
	seen := make(map[int]int)

	for i, num := range nums {
		complement := target - num
		if j, ok := seen[complement]; ok {
			return []int{j, i}
		}
		seen[num] = i
	}

	return nil
}

func main() {
	nums := []int{2, 7, 11, 15}
	target := 9
	fmt.Println(twoSum(nums, target)) // Output: [0 1]
}
