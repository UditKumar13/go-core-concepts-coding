package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type RetryConfig struct {
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	Jitter      bool
}

func DefaultConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 5,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
		Jitter:     true,
	}
}

func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	delay := cfg.BaseDelay

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			fmt.Printf("Attempt %d: SUCCESS\n", attempt)
			return nil
		}

		fmt.Printf("Attempt %d: FAILED — %v\n", attempt, lastErr)

		if attempt == cfg.MaxRetries {
			break
		}

		// Apply jitter to avoid thundering herd
		wait := delay
		if cfg.Jitter {
			jitter := time.Duration(rand.Int63n(int64(delay) / 2))
			wait = delay + jitter
		}

		// Cap at MaxDelay
		if wait > cfg.MaxDelay {
			wait = cfg.MaxDelay
		}

		fmt.Printf("Waiting %v before retry...\n", wait)

		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(wait):
		}

		// Exponential growth: delay = delay * multiplier
		delay = time.Duration(float64(delay) * cfg.Multiplier)
	}

	return fmt.Errorf("all %d attempts failed: %w", cfg.MaxRetries, lastErr)
}

// --- Simulate real-world scenarios ---

var attempt1 int

// Succeeds on 3rd try
func unstableAPI() error {
	attempt1++
	if attempt1 < 3 {
		return errors.New("connection timeout")
	}
	return nil
}

var attempt2 int

// Always fails (to show max retries exhausted)
func alwaysFailingAPI() error {
	attempt2++
	return fmt.Errorf("server unavailable (attempt %d)", attempt2)
}

func main() {
	// Example 1: Unstable API that eventually succeeds
	fmt.Println("=== Example 1: Unstable API (succeeds on 3rd try) ===")
	cfg := DefaultConfig()
	cfg.MaxRetries = 5
	cfg.BaseDelay = 500 * time.Millisecond

	ctx := context.Background()
	if err := Retry(ctx, cfg, unstableAPI); err != nil {
		fmt.Println("Final Error:", err)
	}

	fmt.Println()

	// Example 2: API that always fails
	fmt.Println("=== Example 2: Always Failing API ===")
	cfg2 := DefaultConfig()
	cfg2.MaxRetries = 3
	cfg2.BaseDelay = 200 * time.Millisecond

	if err := Retry(ctx, cfg2, alwaysFailingAPI); err != nil {
		fmt.Println("Final Error:", err)
	}

	fmt.Println()

	// Example 3: Context timeout cancels retry early
	fmt.Println("=== Example 3: Context Timeout Cancels Retry ===")
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	callCount := 0
	if err := Retry(ctxTimeout, DefaultConfig(), func() error {
		callCount++
		return fmt.Errorf("always failing (call %d)", callCount)
	}); err != nil {
		fmt.Println("Final Error:", err)
	}
}
