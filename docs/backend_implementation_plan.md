# Backend Implementation Plan

## User Review Required
> [!IMPORTANT]
> **Stack Recommendation:** The PRD explicitly mentions `Single-Container Deployment (Go Binary)`. Therefore, I strongly recommend **Go** over .NET. It aligns with the "Disruption through Simplicity" philosophy and the containerization strategy.
> **Database:** PostgreSQL 16+ as requested.

## Proposed Architecture
- **Language:** Go 1.22+
- **Framework:** `Chi` (Lightweight, idiomatic router) or Standard Library + Middleware.
- **Database driver:** `pgx` (High performance).
- **ORM/Query:** `sqlc` (Type-safe SQL) or `GORM` (if rapid prototyping is preferred, but `sqlc` fits "simplicity/performance" better). *Recommendation: raw SQL or sqlc for performance.*
- **Containerization:** Docker + OrbStack.

## Proposed Changes

### Project Structure (New Files)
#### [NEW] [backend/go.mod](backend/go.mod)
Initialize Go module.

#### [NEW] [backend/cmd/server/main.go](backend/cmd/server/main.go)
Entry point. Sets up DB connection and Router.

#### [NEW] [backend/internal/server/server.go](backend/internal/server/server.go)
Server struct and startup logic.

#### [NEW] [backend/internal/database/database.go](backend/internal/database/database.go)
Database connection logic using `pgx`.

#### [NEW] [docker-compose.yml](docker-compose.yml)
Define `app` (backend) service with container name `IOI_AMMS` and `db` (Postgres) services.

#### [NEW] [backend/Dockerfile](backend/Dockerfile)
Multi-stage build for Go binary.

### Authentication & RBAC (Pending)
- **JWT:** Use `golang-jwt/jwt/v5`.
- **Flow:** Login -> Access Token (15m) + Refresh Token (7d).
- **RBAC:** Middleware to enforce roles based on PRD Section 4.

#### [NEW] [backend/internal/auth/auth.go](backend/internal/auth/auth.go)
JWT generation and validation logic.

#### [NEW] [backend/internal/middleware/auth.go](backend/internal/middleware/auth.go)
HTTP Middleware for extracting user claims and enforcing roles.

## Verification Plan

### Automated Tests
- Run `go test ./...` (Unit tests).
- `curl` health check endpoint.

### Manual Verification
1. Run `docker-compose up -d`.
2. Connect to DB via TablePlus or CLI to verify connection.
3. Access `http://localhost:8080/health` (or configured port) to see if backend is running.
