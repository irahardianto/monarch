package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"time"

	"github.com/docker/docker/client"
	"github.com/monarch-dev/monarch/api"
	"github.com/monarch-dev/monarch/config"
	"github.com/monarch-dev/monarch/database"
	"github.com/monarch-dev/monarch/project"
	"github.com/monarch-dev/monarch/runner"
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

	// Initialize DB
	if cfg.DB == "" {
		return fmt.Errorf("database URL is required")
	}

	pool, err := database.Connect(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()
	fmt.Println("Connected to database")

	// Initialize Docker Client
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerCli.Close()

	// 1. Reaper: Clean up zombies
	cleaned, err := runner.ReapZombies(ctx, dockerCli)
	if err != nil {
		// Log but don't fail, as it might be permission issue or transient
		fmt.Printf("Warning: failed to reap zombies: %v\n", err)
	} else if cleaned > 0 {
		fmt.Printf("Reaped %d zombie containers\n", cleaned)
	}

	// 2. Runner Manager: Start monitoring
	runMgr := runner.NewManager(dockerCli)
	// Check idle every minute, timeout after 5 minutes
	runMgr.StartMonitor(ctx, 1*time.Minute, 5*time.Minute)

	// Initialize Services
	projStore := project.NewPostgresStore(pool)
	projSvc := project.NewService(projStore)

	// Initialize Server
	srv := api.NewServer(cfg, pool, projSvc)

	fmt.Printf("Monarch Supervisor starting on port %d [%s]\n", cfg.Port, cfg.Env)

	return http.ListenAndServe(":"+strconv.Itoa(cfg.Port), srv)
}
