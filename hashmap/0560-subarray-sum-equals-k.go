package main

// Go's maps return 0 for missing keys by default
// — so no need for getOrDefault. Clean and concise.

// Seed the map with {0: 1} to handle subarrays starting from index 0.
// TC: O(n) | SC: O(n)

func subarraySum(nums []int, k int) int {
	prefixCount := map[int]int{0: 1} // seed empty prefix

	currentSum := 0
	count := 0

	for _, num := range nums {
		currentSum += num

		// How many times has (currentSum - k) appeared?
		count += prefixCount[currentSum-k]

		// Record this prefix sum
		prefixCount[currentSum]++
	}

	return count
}
