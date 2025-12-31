# Core Foundation Implementation State (2025-12-31)

## Completed
- **Scaffold**: `apps/backend` module initialized, `config` package implemented. Added `cmd/monarch/main.go` and `Makefile`.
- **Database**: `sqlc` configured with `pgx/v5`, schema defined, code generated.
- **Runner**: `ReapZombies` implemented and integration tested. `Manager` implemented with `GetOrStart` and `IdleTimeout`.

## Architecture Decisions
- **Monorepo**: Confirmed `apps/backend` structure.
- **Database**: Using `pgx/v5` pool and `sqlc` for type safety.
- **Docker**: Using `github.com/docker/docker` module. Interface-based design for testability.
- **Concurrency**: `sync.RWMutex` used for thread-safe map access in `Manager`.

## Next Steps
- Implement HTTP API to expose these features.
- Integrate MCP server.
