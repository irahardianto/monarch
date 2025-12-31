# Plan: Epic 2 - Project & Gate Configuration

**Status:** Implemented
**Epic:** 2 (Project & Gate Config)
**Stories:** 2.1, 2.2
**Date:** 2025-12-31

## **Overview**

This plan implements the Project Registration workflow and Gate Configuration system. It enables users to register local projects, validates their paths, and automatically detects the technology stack (Go, Node, Python) if a configuration file is missing. It also establishes the HTTP API layer to expose these capabilities.

## **Requirements Extraction**

### **Scope**
- **Stories:** 2.1 (Registration/Path Validation), 2.2 (Gate Parsing/Auto-detect).
- **Core Entities:** `Project` (DB), `GateConfig` (YAML/InMemory).
- **Infrastructure:** `net/http` (Router), `gopkg.in/yaml.v3`.

### **Gap Analysis**
- **Nouns:**
  - `Project Path`: Local filesystem path.
  - `Gates`: List of validation rules (Tier A/B/C).
  - `Stack Detection`: Heuristics (go.mod, package.json).
  - `API Server`: HTTP entry point.
- **Verbs:**
  - `Register`: Validate path -> Save to DB.
  - `Parse`: Read YAML -> Struct.
  - `Detect`: Scan files -> Generate Config.
  - `Route`: Map HTTP verbs to handlers.

### **Exclusions**
- **Git Operations:** "Diff Mode" logic is defined in Epic 5. We only check for `.git` existence here.
- **MCP Endpoints:** Deferred to Epic 4.
- **UI:** Deferred to Epic 6.

## **Knowledge Enrichment**

### **RAG Findings**
- **HTTP Server:** Use `net/http` with `http.NewServeMux()` for standard routing (Go 1.22+).
- **Middleware:** Use `slog` for structured request logging.
- **YAML:** Use `gopkg.in/yaml.v3` for parsing `gates.yaml`.
- **File Checks:** Use `os.Stat()` to verify paths and check for `.git`, `go.mod`, etc.

## **Implementation Plan**

### **Task 1: HTTP API Server Scaffold (Completed)**

**Files:**
- Create: `apps/backend/api/server.go`
- Create: `apps/backend/api/routes.go`
- Test: `apps/backend/api/server_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. `NewServer(cfg, db)` initializes HTTP server.
  2. Middleware stack includes Logging (slog) and Recovery.
  3. Returns 404 for unknown routes.
  4. `/health` endpoint returns 200 OK.

- **Test Coverage**
  - [Unit] `GET /health` returns 200.

**Step 1: Write failing test**
```go
// apps/backend/api/server_test.go
package api_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/monarch-dev/monarch/api"
    "github.com/stretchr/testify/assert"
)

func TestServer_Health(t *testing.T) {
    srv := api.NewServer(nil, nil) // Mock deps
    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()
    
    srv.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

**Step 3: Minimal Implementation**
```go
// apps/backend/api/server.go
package api

import (
    "net/http"
    "github.com/monarch-dev/monarch/config"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
    mux *http.ServeMux
    db  *pgxpool.Pool
    cfg *config.Config
}

func NewServer(cfg *config.Config, db *pgxpool.Pool) *Server {
    mux := http.NewServeMux()
    s := &Server{mux: mux, db: db, cfg: cfg}
    s.routes()
    return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.mux.ServeHTTP(w, r)
}
```

```go
// apps/backend/api/routes.go
package api

import "net/http"

func (s *Server) routes() {
    s.mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
}
```

---

### **Task 2: Gate Configuration Logic (Completed)**

**Files:**
- Create: `apps/backend/gates/types.go`
- Create: `apps/backend/gates/parser.go`
- Test: `apps/backend/gates/parser_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. `Parse(path)` reads `.monarch/gates.yaml`.
  2. If file missing, `DetectStack(path)` checks `go.mod`, `package.json`, `requirements.txt`.
  3. Returns `GateConfig` struct with correct environment type.

- **Functional Requirements**
  - Support `Tier A` (Standard) gate definitions.

- **Test Coverage**
  - [Unit] `Parse()` with valid YAML.
  - [Unit] `DetectStack()` with temp dir containing `go.mod`.

**Step 1: Write failing test**
```go
// apps/backend/gates/parser_test.go
package gates_test

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/monarch-dev/monarch/gates"
    "github.com/stretchr/testify/assert"
)

func TestDetectStack_Go(t *testing.T) {
    tmp := t.TempDir()
    os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module test"), 0644)
    
    cfg, err := gates.DetectStack(tmp)
    assert.NoError(t, err)
    assert.Equal(t, "go", cfg.Stack)
}
```

**Step 3: Minimal Implementation**
```go
// apps/backend/gates/types.go
package gates

type Config struct {
    Stack string   `yaml:"stack"`
    Gates []Gate   `yaml:"gates"`
}

type Gate struct {
    Name    string `yaml:"name"`
    Command string `yaml:"command"`
    Tier    string `yaml:"tier"` // A, B, C
}
```

```go
// apps/backend/gates/parser.go
package gates

import (
    "os"
    "path/filepath"
)

func DetectStack(root string) (*Config, error) {
    if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
        return &Config{Stack: "go"}, nil
    }
    // Add other checks
    return &Config{Stack: "unknown"}, nil
}
```

---

### **Task 3: Project Registration Feature (Completed)**

**Files:**
- Create: `apps/backend/project/service.go`
- Create: `apps/backend/project/handlers.go`
- Create: `apps/backend/project/store.go` (Interface)
- Create: `apps/backend/project/postgres.go` (Implementation)
- Test: `apps/backend/project/service_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. `Register(path)` validates path existence.
  2. Checks for `.git` directory.
  3. Calls `gates.Parser` to load config.
  4. Saves Project + Gate Config to DB.

- **Test Coverage**
  - [Unit] `Register()` fails on invalid path.
  - [Unit] `Register()` succeeds on valid path + mock DB.

**Step 1: Write failing test**
```go
// apps/backend/project/service_test.go
package project_test

import (
    "testing"
    "github.com/monarch-dev/monarch/project"
    "github.com/stretchr/testify/assert"
)

func TestRegister_InvalidPath(t *testing.T) {
    svc := project.NewService(nil, nil) // Mock store/gates
    _, err := svc.Register(context.Background(), "/invalid/path")
    assert.Error(t, err)
}
```

**Step 3: Minimal Implementation**
```go
// apps/backend/project/service.go
package project

import (
    "context"
    "errors"
    "os"
    "path/filepath"
    "github.com/monarch-dev/monarch/gates"
)

type Service struct {
    store Store
}

func (s *Service) Register(ctx context.Context, path string) (*Project, error) {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, errors.New("path does not exist")
    }
    
    // Check Git
    hasGit := false
    if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
        hasGit = true
    }

    // Detect Gates
    // ... logic calling gates.DetectStack ...

    // Save to DB
    // ... logic calling s.store.Create ...
    return &Project{Path: path, HasGit: hasGit}, nil
}
```

---

### **Task 4: API Wiring & Main Integration (Completed)**

**Files:**
- Modify: `apps/backend/cmd/monarch/main.go`
- Modify: `apps/backend/api/routes.go` (Add project routes)

**Requirements:**
- **Acceptance Criteria**
  1. `main.go` initializes `project.Service`.
  2. `api.Server` mounts project handlers.
  3. `POST /projects` calls `Register`.

**Step 3: Minimal Implementation**
```go
// apps/backend/cmd/monarch/main.go
// ...
projStore := project.NewPostgresStore(dbPool)
projSvc := project.NewService(projStore)
srv := api.NewServer(cfg, dbPool, projSvc)
// ...
```
