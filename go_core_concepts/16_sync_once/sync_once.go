package main

import (
	"fmt"
	"sync"
)

// --- Example 1: Singleton DB Connection ---

type Database struct {
	name string
}

var (
	dbInstance *Database
	dbOnce     sync.Once
)

func GetDB() *Database {
	dbOnce.Do(func() {
		fmt.Println("  → Initializing DB connection (only once)...")
		dbInstance = &Database{name: "PostgreSQL"}
	})
	return dbInstance
}

// --- Example 2: Config loader ---

type Config struct {
	Env     string
	MaxConn int
}

var (
	cfg     *Config
	cfgOnce sync.Once
)

func GetConfig() *Config {
	cfgOnce.Do(func() {
		fmt.Println("  → Loading config from disk (only once)...")
		cfg = &Config{Env: "production", MaxConn: 100}
	})
	return cfg
}

// --- Example 3: Proving only one goroutine initializes ---

func main() {
	fmt.Println("=== Example 1: Singleton DB with sync.Once ===")
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			db := GetDB()
			fmt.Printf("  Goroutine %d got DB: %s\n", id, db.name)
		}(i)
	}
	wg.Wait()

	fmt.Println()
	fmt.Println("=== Example 2: Config Loader with sync.Once ===")

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c := GetConfig()
			fmt.Printf("  Goroutine %d got config: env=%s maxConn=%d\n", id, c.Env, c.MaxConn)
		}(i)
	}
	wg.Wait()

	fmt.Println()
	fmt.Println("=== Example 3: Proving Same Pointer Returned ===")
	db1 := GetDB()
	db2 := GetDB()
	db3 := GetDB()
	fmt.Printf("db1 pointer: %p\n", db1)
	fmt.Printf("db2 pointer: %p\n", db2)
	fmt.Printf("db3 pointer: %p\n", db3)
	fmt.Println("All same?", db1 == db2 && db2 == db3)
}
