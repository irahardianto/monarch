# Core Foundation State

## Architecture
- **Language:** Go 1.25+
- **Database:** PostgreSQL + pgvector
- **Orchestration:** Docker SDK for Go (Sidecars)
- **API:** Chi Router + Standard Lib
- **Agent Protocol:** MCP (Model Context Protocol) via SSE

## Implemented Components
1.  **Project Management:**
    - Registration with git/stack detection.
    - Gate parsing (.monarch/gates.yaml).
    - Database schema for projects/tasks.

2.  **Runner Engine:**
    - `Executor`: Docker Exec wrapper for running commands.
    - `Manager`: Warm container lifecycle (Start, Get, Idle Timeout).
    - `Reaper`: Cleanup of orphaned containers on startup.
    - `Parsers`: Fail-closed parsing for `go test` and `eslint`.

3.  **Agent Interface (MCP):**
    - `MCPServer`: Core server instance using `go-sdk`.
    - `SSEHandler`: Transport layer bridging HTTP to MCP (via `go-sdk`).
    - `Planner`: Tools for querying state (`list_projects`, `search_past_tasks`).
    - `Builder`: Tools for executing tasks (`claim_task`, `submit_attempt`) with Circuit Breaker.
