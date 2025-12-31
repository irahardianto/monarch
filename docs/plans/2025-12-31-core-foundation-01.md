# Plan: Epic 1 - Core Foundation & Supervisor Engine

**Status:** Completed
**Epic:** 1 (Core Foundation)
**Stories:** 1.1, 1.2, 1.3, 1.4
**Date:** 2025-12-31

## **Overview**

This plan implements the core runtime environment for Monarch. It establishes the Go binary scaffold, a **type-safe data access layer using `sqlc`**, and the Docker orchestration engine (Supervisor). It adheres to the **Monorepo Feature-Based** structure defined in the Technical Constitution.

## **Requirements Extraction**

### **Scope**
- **Stories:** 1.1 (Scaffold/DB), 1.2 (Zombie Reaper), 1.3 (Lifecycle Manager), 1.4 (Idle Timeout).
- **Core Entities:** `Project`, `Task`, `RunnerContainer`.
- **Infrastructure:** Go 1.25+, PostgreSQL 16 (pgvector), Docker Engine, **sqlc**.

### **Gap Analysis**
- **Nouns:**
  - `Monarch Binary`: `apps/backend/cmd/monarch`.
  - `Config`: `apps/backend/config` feature.
  - `DB Connection`: `apps/backend/database` feature (pgxpool).
  - `Data Access`: `apps/backend/database/sqlc` (Generated Code).
  - `Runner`: `apps/backend/runner` feature (Reaper, Lifecycle, Monitor).
- **Verbs:**
  - `Boot`: Initialize DB, migrate, reap zombies.
  - `Reap`: Kill/Remove Docker containers.
  - `GetOrStart`: Find existing or spawn new container.
  - `Monitor`: Check for inactivity and stop containers.
  - `Generate`: Create Go code from SQL (`sqlc generate`).

### **Exclusions**
- **MCP Server:** Deferred to Epic 4.
- **Dashboard UI:** Deferred to Epic 6.
- **Git Operations:** Deferred to Epic 2.

## **Knowledge Enrichment**

### **RAG Findings**
- **Docker SDK:** Use `github.com/docker/docker/client` with `filters.NewArgs(filters.Arg("label", "monarch.managed=true"))`.
- **Postgres:** Use `github.com/jackc/pgx/v5` for connection pooling.
- **sqlc:** Use `pgx/v5` driver output, `emit_interface: true` for mockability.
- **Logging:** Use `log/slog` (std lib) for structured JSON logging.
- **Structure:** `apps/backend` as module root, feature folders (`runner`, `config`) contain business logic + storage + handlers.

## **Implementation Plan**

### **Task 1: Project Scaffold & Go Module Init**

**Files:**
- Create: `apps/backend/go.mod`
- Create: `apps/backend/cmd/monarch/main.go`
- Create: `apps/backend/config/config.go`
- Create: `apps/backend/config/config_test.go`
- Create: `apps/backend/Makefile` (optional helper)

**Requirements:**
- **Acceptance Criteria**
  1. `go mod init github.com/monarch-dev/monarch` executed in `apps/backend`.
  2. Directory structure created (`apps/backend/config`).
  3. `config` feature loads `MONARCH_PORT` (default 9090) and `DATABASE_URL`.

- **Test Coverage**
  - [Unit] `Load()`: Validate defaults and env overrides.

**Step 1: Write failing test (Config)**
```go
// apps/backend/config/config_test.go
package config_test

import (
	"os"
	"testing"
	"github.com/monarch-dev/monarch/config" // Import path relative to module root
	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()
	cfg, err := config.Load()
	assert.NoError(t, err)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "development", cfg.Env)
}
```

**Step 3: Minimal Implementation**
```go
// apps/backend/config/config.go
package config

import (
    "os"
    "strconv"
)

type Config struct {
    Port int
    Env  string
    DB   string
}

func Load() (*Config, error) {
    portStr := os.Getenv("MONARCH_PORT")
    port := 9090
    if portStr != "" {
        var err error
        port, err = strconv.Atoi(portStr)
        if err != nil {
            return nil, err
        }
    }
    
    return &Config{
        Port: port,
        Env:  getEnv("MONARCH_ENV", "development"),
        DB:   os.Getenv("DATABASE_URL"),
    }, nil
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

---

### **Task 2: Database Connection, sqlc Setup & Migrations**

**Files:**
- Create: `apps/backend/sqlc.yaml`
- Create: `apps/backend/database/schema.sql`
- Create: `apps/backend/database/query.sql`
- Create: `apps/backend/database/database.go`
- Generate: `apps/backend/database/gen/models.go` (via sqlc)
- Test: `apps/backend/database/database_integration_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. `sqlc` configured with `pgx/v5` and `emit_interface: true`.
  2. `schema.sql` defines `projects` and `tasks` tables + `vector` extension.
  3. `query.sql` defines basic CRUD (GetProject, CreateProject).
  4. `sqlc generate` creates type-safe Go code.
  5. `pgxpool.Connect` establishes connection.

- **Test Coverage**
  - [Integration] `Connect()`: Verifies connection to Postgres (using Testcontainers).
  - [Integration] `Queries`: Test generated code against real DB.

**Step 1: Write sqlc config**
```yaml
# apps/backend/sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "database/query.sql"
    schema: "database/schema.sql"
    gen:
      go:
        package: "database"
        out: "database"
        sql_package: "pgx/v5"
        emit_interface: true
        emit_json_tags: true
```

**Step 2: Write Schema**
```sql
-- apps/backend/database/schema.sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    path TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Step 3: Write Query**
```sql
-- apps/backend/database/query.sql
-- name: CreateProject :one
INSERT INTO projects (path) VALUES ($1) RETURNING *;

-- name: GetProject :one
SELECT * FROM projects WHERE path = $1 LIMIT 1;
```

**Step 4: Minimal Implementation (DB Connection)**
```go
// apps/backend/database/database.go
package database

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, connString string) (*pgxpool.Pool, error) {
    return pgxpool.New(ctx, connString)
}
```

---

### **Task 3: Docker Client & Zombie Reaper**

**Files:**
- Create: `apps/backend/runner/docker_client.go`
- Create: `apps/backend/runner/reaper.go`
- Test: `apps/backend/runner/reaper_integration_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. `ReapZombies` lists containers with label `monarch.managed=true`.
  2. `ReapZombies` force-removes them.
  3. Returns count of removed containers.

- **Test Coverage**
  - [Integration] `ReapZombies()`: Spawn dummy container with label, run reap, verify gone.

**Step 1: Write failing test**
```go
// apps/backend/runner/reaper_integration_test.go
// Uses real Docker client or mock
```

**Step 3: Minimal Implementation**
```go
// apps/backend/runner/reaper.go
package runner

import (
    "context"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/filters"
    "github.com/docker/docker/client"
)

func ReapZombies(ctx context.Context, cli *client.Client) (int, error) {
    args := filters.NewArgs()
    args.Add("label", "monarch.managed=true")
    
    containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
        All: true,
        Filters: args,
    })
    if err != nil {
        return 0, err
    }
    
    count := 0
    for _, c := range containers {
        if err := cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true}); err == nil {
            count++
        }
    }
    return count, nil
}
```

---

### **Task 4: Container Lifecycle Manager (Core)**

**Files:**
- Create: `apps/backend/runner/manager.go`
- Test: `apps/backend/runner/manager_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. `GetOrStart(stack)` checks internal map.
  2. If missing, starts new container with label.
  3. Updates map with Container ID.

- **Test Coverage**
  - [Unit] `GetOrStart()`: Mock Docker client, verify `ContainerCreate` called if map empty.

**Step 3: Minimal Implementation**
```go
// apps/backend/runner/manager.go
type Manager struct {
    cli *client.Client
    // Map ProjectID -> Stack -> ContainerID
    runners map[string]map[string]string 
}

func (m *Manager) GetOrStart(ctx context.Context, projectID, stack string) (string, error) {
    // Logic to check map or start container
    return "container-id", nil
}
```

---

### **Task 5: Resource Hygiene (Idle Timeout)**

**Files:**
- Modify: `apps/backend/runner/manager.go`
- Test: `apps/backend/runner/monitor_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. Background goroutine checks last usage time.
  2. Stops container if idle > 5 mins.
  3. Removes from active map.

- **Test Coverage**
  - [Unit] `checkIdle()`: Advance time, verify stop called.

**Step 3: Minimal Implementation**
```go
// apps/backend/runner/manager.go
// Add lastUsed map and StartMonitor() function
```
