# IOI AMMS Backend - Full Implementation Roadmap

## Current Status Overview

| Phase | Status | Description |
|-------|--------|-------------|
| Infrastructure | ✅ Done | Docker, PostgreSQL, MinIO |
| Auth | ✅ Done | JWT (mock user, needs DB integration) |
| Enterprise Config | ⏳ Pending | Environment-based config management |
| Core APIs | ⏳ Pending | Assets, Work Orders, Users CRUD |
| RBAC Enforcement | ⏳ Pending | Middleware-level role checks |
| Testing | ⏳ Pending | Unit & Integration tests |
| Production Readiness | ⏳ Pending | Logging, Monitoring, Healthchecks |

---

## 🔴 Critical Improvements (Enterprise-Grade)

### 1. Configuration Management
**Issue:** Secrets hardcoded or scattered across `os.Getenv()` calls.
**Solution:** Centralized `internal/config` package.

```
backend/internal/config/
├── config.go       # Struct + loader
└── config.yaml     # Optional file-based config
```

**Config Struct Example:**
```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Auth     AuthConfig
    MinIO    MinIOConfig
}
```

**Environment Variables (Production):**
- `JWT_SECRET` (required)
- `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`
- `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`

---

### 2. Structured Logging
**Issue:** Using `log.Fatal` which kills the app.
**Solution:** Use `slog` (Go 1.21+) or `zerolog`.

---

### 3. Error Handling
**Issue:** Inconsistent error responses.
**Solution:** Standardized error package with RFC 7807 Problem Details.

---

### 4. Database Layer
**Issue:** Global singleton, no transaction support.
**Solution:** Repository pattern with context-based transactions.

---

## 📋 Full Implementation Checklist

### Phase 1: Enterprise Foundation (Priority: High)
- [ ] **Config Package** - Centralized configuration loading
- [ ] **Environment Files** - `.env.example`, `.env.local`, `.env.production`
- [ ] **Structured Logging** - Replace `log` with `slog`
- [ ] **Graceful Shutdown** - Handle SIGTERM properly
- [ ] **CORS Middleware** - Allow frontend origins

### Phase 2: Database & Repository Layer
- [ ] **Repository Pattern** - `internal/repository/` for each entity
- [ ] **User Repository** - DB-backed user lookup for auth
- [ ] **Asset Repository** - CRUD operations
- [ ] **Work Order Repository** - CRUD operations
- [ ] **Migration Tool** - `golang-migrate` or `goose`

### Phase 3: Core API Modules (PRD Alignment)
- [ ] **Users API** - CRUD, password reset
- [ ] **Assets API** - List, Create, Update, Delete, Search
- [ ] **Locations API** - Hierarchical location management
- [ ] **Work Orders API** - Full lifecycle
- [ ] **Service Requests API** - Triage queue
- [ ] **Inventory API** - Parts, Stock, Wallets
- [ ] **Daily Logs API** - Bulk entry
- [ ] **Audit Logs** - Automatic change tracking

### Phase 4: RBAC & Security
- [ ] **Role-Based Middleware** - Enforce PRD Section 4 permissions
- [ ] **Tenant Isolation** - Multi-tenancy query scoping
- [ ] **Rate Limiting** - Prevent abuse
- [ ] **Input Validation** - Request validation middleware

### Phase 5: File Storage (MinIO)
- [ ] **Upload Service** - `internal/storage/minio.go`
- [ ] **Asset Attachments** - Photos, documents
- [ ] **Presigned URLs** - Secure file access

### Phase 6: Testing & Quality
- [ ] **Unit Tests** - Repository & service layer
- [ ] **Integration Tests** - API endpoint tests
- [ ] **Test Database** - Separate test container
- [ ] **CI Pipeline** - GitHub Actions / GitLab CI

### Phase 7: Production Readiness
- [ ] **Health Checks** - Liveness & readiness probes
- [ ] **Metrics** - Prometheus endpoint
- [ ] **Dockerfile (Prod)** - Multi-stage build, minimal image
- [ ] **Helm Chart** - Kubernetes deployment (if needed)
- [ ] **Secrets Management** - Vault / K8s secrets

---

## Recommended Folder Structure (Final)

```
backend/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/          # Configuration loading
│   ├── auth/            # JWT, password
│   ├── middleware/      # Auth, CORS, Logging
│   ├── database/        # Connection pool
│   ├── repository/      # Data access layer
│   │   ├── user.go
│   │   ├── asset.go
│   │   └── workorder.go
│   ├── service/         # Business logic layer
│   │   ├── user.go
│   │   ├── asset.go
│   │   └── workorder.go
│   ├── handler/         # HTTP handlers (controllers)
│   │   ├── auth.go
│   │   ├── asset.go
│   │   └── workorder.go
│   ├── routes/          # Route registration
│   ├── storage/         # MinIO integration
│   └── model/           # Shared data structs
├── migrations/          # SQL migrations
├── docs/                # API specs
├── Dockerfile
├── .env.example
└── go.mod
```

---

## Next Immediate Actions

1. **Implement Config Package** - Most critical for production deployments.
2. **Add `.env.example`** - Document all required environment variables.
3. **Create User Repository** - Connect auth to real database.
4. **Build Assets API** - First module per PRD.
