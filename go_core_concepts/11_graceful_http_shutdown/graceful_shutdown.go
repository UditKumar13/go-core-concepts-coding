package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request received, processing...")
	time.Sleep(3 * time.Second) // simulate slow work
	fmt.Fprintln(w, "Response done!")
	fmt.Println("Request handled.")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Channel to listen for OS signals (Ctrl+C or SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		fmt.Println("Server running on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("Server error:", err)
		}
	}()

	// Block until a signal is received
	<-quit
	fmt.Println("\nShutdown signal received. Gracefully shutting down...")

	// Give in-flight requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Forced shutdown:", err)
	} else {
		fmt.Println("Server shut down cleanly.")
	}
}
