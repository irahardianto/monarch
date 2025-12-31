# Core Foundation Implementation State (2025-12-31)

## Completed
- **Scaffold**: `apps/backend` module initialized, `config` package implemented. Added `cmd/monarch/main.go` and `Makefile`.
- **Database**: `sqlc` configured with `pgx/v5`, schema defined, code generated.
- **Runner**: `ReapZombies` implemented and integration tested. `Manager` implemented with `GetOrStart` and `IdleTimeout`.
- **API**: HTTP Server scaffolded with `slog` logging and panic recovery middleware (`apps/backend/api`).
- **Gates**: Configuration parser (`gates.yaml`) and stack auto-detection (Go, Node, Python) implemented (`apps/backend/gates`).
- **Project**: Registration service implemented with path validation, Git detection, and DB persistence (`apps/backend/project`).
- **Integration**: `main.go` wires `ProjectService`, `PostgresStore`, and `APIServer`.
- **Background Jobs**: `RunnerManager` (Monitor) and `Reaper` wired into `main.go` startup.

## Architecture Decisions
- **Monorepo**: Confirmed `apps/backend` structure.
- **Database**: Using `pgx/v5` pool and `sqlc` for type safety.
- **Docker**: Using `github.com/docker/docker` module. Interface-based design for testability.
- **Concurrency**: `sync.RWMutex` used for thread-safe map access in `Manager`.
- **API**: Standard `net/http` ServeMux with custom middleware chain. Feature-based handlers.

## Next Steps
- Implement MCP Server.
- Add "Planner" toolset.
