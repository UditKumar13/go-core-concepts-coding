package main

import (
	"fmt"
	"strings"
	"sync"
	"unicode"
)

type FileResult struct {
	Name      string
	WordCount int
	CharCount int
	Lines     int
}

func processFile(name, content string) FileResult {
	lines := strings.Count(content, "\n") + 1
	charCount := 0
	wordCount := 0
	inWord := false

	for _, ch := range content {
		if !unicode.IsSpace(ch) {
			charCount++
			if !inWord {
				wordCount++
				inWord = true
			}
		} else {
			inWord = false
		}
	}

	fmt.Printf("  Processed: %s\n", name)
	return FileResult{
		Name:      name,
		WordCount: wordCount,
		CharCount: charCount,
		Lines:     lines,
	}
}

func processFiles(files map[string]string) []FileResult {
	results := make([]FileResult, 0, len(files))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, content := range files {
		wg.Add(1)
		go func(n, c string) {
			defer wg.Done()
			result := processFile(n, c)
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(name, content)
	}

	wg.Wait()
	return results
}

func processFilesWithWorkerPool(files map[string]string, workerCount int) []FileResult {
	type job struct {
		name    string
		content string
	}

	jobs := make(chan job, len(files))
	results := make(chan FileResult, len(files))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := range jobs {
				fmt.Printf("  Worker %d processing: %s\n", workerID, j.name)
				results <- processFile(j.name, j.content)
			}
		}(i + 1)
	}

	// Send all jobs
	for name, content := range files {
		jobs <- job{name, content}
	}
	close(jobs)

	// Wait then close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var output []FileResult
	for r := range results {
		output = append(output, r)
	}
	return output
}

func printResults(results []FileResult) {
	fmt.Printf("  %-20s %6s %6s %6s\n", "File", "Words", "Chars", "Lines")
	fmt.Println("  " + strings.Repeat("-", 42))
	for _, r := range results {
		fmt.Printf("  %-20s %6d %6d %6d\n", r.Name, r.WordCount, r.CharCount, r.Lines)
	}
}

func main() {
	files := map[string]string{
		"readme.txt":    "Hello world\nThis is a readme file\nIt has three lines",
		"notes.txt":     "Go is fast\nGo is concurrent\nGo is fun",
		"log.txt":       "Error at line 42\nWarning at line 99\nInfo at line 200",
		"config.txt":    "host=localhost\nport=8080\ndebug=true",
		"changelog.txt": "v1.0 initial release\nv1.1 bug fixes\nv1.2 new features added",
	}

	fmt.Println("=== Example 1: Concurrent Processing (one goroutine per file) ===")
	results1 := processFiles(files)
	printResults(results1)

	fmt.Println()
	fmt.Println("=== Example 2: Worker Pool (2 workers for 5 files) ===")
	results2 := processFilesWithWorkerPool(files, 2)
	printResults(results2)
}
