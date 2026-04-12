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

func NewLogger() Logger {
	return Logger{nextAllowed: make(map[string]int)}
}

// (l *Logger) — Method Receiver in Go
// This is Go's way of attaching a method to a struct. It's the equivalent of this in Java.
func (l *Logger) ShouldPrintMessage(timestamp int, message string) bool {
	if timestamp < l.nextAllowed[message] {

		// in Java, this would throw an exception for missing key, but in Go, missing map key returns zero value (0), so this handles new messages too
		// in java, syntax will be l.nextAllowed.getOrDefault(message, 0) to handle missing keys, but in Go, missing map key returns zero value (0), so this handles new messages too
		// In Go, missing map key returns zero value (0), so this handles new messages too
		return false
	}
	l.nextAllowed[message] = timestamp + 10
	return true
}
