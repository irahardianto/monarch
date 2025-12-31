package main

import (
	"context"
	"fmt"
	"os"

	"github.com/monarch-dev/monarch/config"
	"github.com/monarch-dev/monarch/database"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()

	// Initialize DB if URL provided
	if cfg.DB != "" {
		pool, err := database.Connect(ctx, cfg.DB)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer pool.Close()
		fmt.Println("Connected to database")
	}

	fmt.Printf("Monarch Supervisor starting on port %d [%s]\n", cfg.Port, cfg.Env)
	
	// Future: Start HTTP server, Runner Manager, etc.
	
	return nil
}
