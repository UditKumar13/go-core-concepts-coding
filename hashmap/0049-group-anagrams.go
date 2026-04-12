package main

import "sort"

func groupAnagrams(strs []string) [][]string {
	m := make(map[string][]string)
	for _, s := range strs {
		b := []byte(s)
		sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
		key := string(b)
		m[key] = append(m[key], s)
	}
	result := make([][]string, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

// TC: O(n · k log k) | SC: O(n · k)
// difference here with java is :
// The Go version uses sort.Slice on a byte slice (since Go strings are immutable); Java uses toCharArray() + Arrays.sort().
//  Both then convert back to a string for the map key.

// In Go, reading a missing key from a map doesn't panic —
//  it returns the zero value for that type. For []string, the zero value is nil.
// append(nil, s) is perfectly valid — Go treats nil as an empty slice
//  and returns a new slice with s as the first element.
// m["aet"]        →  nil          (key doesn't exist yet)
// append(nil, "eat")  →  ["eat"]
// m["aet"] = ["eat"]

// in java
// Without computeIfAbsent, you'd write:
// if (!map.containsKey(key)) {
//     map.put(key, new ArrayList<>());
// }
// map.get(key).add(s);

// // computeIfAbsent collapses this to one line:
// map.computeIfAbsent(key, k -> new ArrayList<>()).add(s);
