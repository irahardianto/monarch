# Repository Structure

## Current Structure
*   `docs/`: Documentation, requirements (PRD, Epics), and plans.
*   `.gemini/`: Gemini agent configuration and context.
*   `.serena/`: Serena agent memory and project state.

## Planned Structure (Monorepo Feature-Based)
*   `apps/`:
    *   `backend/`: Go application (Supervisor Engine).
        *   `cmd/monarch/main.go`: Application entry point.
        *   `config/`: Configuration feature (Env loading).
        *   `database/`: Database feature (Connection, Migrations).
        *   `runner/`: Execution Supervisor feature (Docker orchestration).
        *   `go.mod`: Go module definition.
    *   `frontend/`: Vue.js application (Mission Control).
        *   `src/`: Source code.
        *   `package.json`: Frontend dependencies.
*   `e2e/`: End-to-End tests.
*   `scripts/`: Shared build/deploy scripts.

## Key Conventions
*   **Feature-Based Packaging:** Code organized by business domain (`runner`, `project`), not technical layer.
*   **Monorepo:** Backend and Frontend coexist in `apps/`.
*   **Test Co-location:** Unit/Integration tests reside next to the code they test (`runner_test.go`).
