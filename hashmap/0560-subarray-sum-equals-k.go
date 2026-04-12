package main

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
