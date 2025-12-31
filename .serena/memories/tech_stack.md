# Tech Stack

## Backend
*   **Language:** Go (1.25+)
*   **Framework:** Standard Library + `net/http` (No heavy web framework)
*   **Database Driver:** `pgx/v5` (High performance PostgreSQL driver)
*   **ORM/Data Access:** `sqlc` (Type-safe SQL generator)
*   **Orchestration:** Docker SDK for Go
*   **Protocol:** Model Context Protocol (MCP) via SSE
*   **Logging:** `log/slog` (Structured JSON logging)

## Frontend
*   **Framework:** Vue.js 3 (Composition API)
*   **Language:** TypeScript
*   **Build Tool:** Vite
*   **UI Library:** Shadcn Vue (Tailwind CSS)
*   **State Management:** Pinia
*   **Real-time:** EventSource (Native SSE)

## Database
*   **Engine:** PostgreSQL 16
*   **Extensions:** `pgvector` (Vector embeddings for semantic search)
*   **Migrations:** `golang-migrate` or raw SQL managed by `sqlc` workflow
