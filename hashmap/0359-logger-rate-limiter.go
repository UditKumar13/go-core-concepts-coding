package main

// Design a Logger class with one method:
// shouldPrintMessage(int timestamp, String message) → boolean
// Rule: A message can only be printed if it was not printed in the last 10 seconds.
// Timestamps arrive in non-decreasing (chronological) order.
// Example:
// shouldPrintMessage(1,  "foo") → true   (first time)
// shouldPrintMessage(2,  "bar") → true   (first time)
// shouldPrintMessage(3,  "foo") → false  (3 < 1+10 = 11)
// shouldPrintMessage(8,  "bar") → false  (8 < 2+10 = 12)
// shouldPrintMessage(10, "foo") → false  (10 < 11)
// shouldPrintMessage(11, "foo") → true   (11 >= 11 ✅)

type Logger struct {
	nextAllowed map[string]int
}

func Constructor() Logger {
	return Logger{nextAllowed: make(map[string]int)}
}

func (l *Logger) ShouldPrintMessage(timestamp int, message string) bool {
	if timestamp < l.nextAllowed[message] {
		// In Go, missing map key returns zero value (0), so this handles new messages too
		return false
	}
	l.nextAllowed[message] = timestamp + 10
	return true
}
