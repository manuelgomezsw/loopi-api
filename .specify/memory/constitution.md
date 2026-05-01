# loopi-api Constitution

## Core Principles

### I. Clean Architecture Layering

The codebase is organized in strict layers: `interface/` → `application/` → `domain/` → `infrastructure/`. Dependencies flow inward only.

- **`internal/interface/handler/`** — HTTP binding only: decode request, call service, encode response. Zero business logic.
- **`internal/application/service/`** — All business logic lives here.
- **`internal/domain/`** — Entities, repository interfaces, pure domain functions. No framework dependencies.
- **`internal/infrastructure/`** — Repository implementations, auth, database drivers.

**Source**: Actual directory structure under `internal/`.
**Verification**: No import from `domain/` into `interface/`; no `infrastructure/` imports in `domain/`.

### II. Repository Contracts in Domain

Repository interfaces are declared in `internal/domain/repository/` and implemented in `internal/infrastructure/repository/`. Repositories do data access only — zero business logic inside them.

**Source**: `internal/domain/repository/*.go` (interfaces) vs `internal/infrastructure/repository/*.go` (implementations).
**Verification**: Repository structs only contain SELECT/INSERT/UPDATE/DELETE SQL calls.

### III. Structured Logging — slog Only

All logging uses `log/slog` from stdlib via the context pattern:

```go
log := logger.FromContext(ctx)
log.InfoContext(ctx, "event", "operation", "domain.Method", "entity_id", id)
```

No `fmt.Printf`, `log.Printf`, or third-party loggers (zap, logrus, zerolog).
Logs belong in **services only** — not in handlers, not in repositories.
Forbidden log fields: `password`, `password_hash`, `token`, `jwt`, `secret`, `authorization`.

**Source**: `pkg/logger/`, `CLAUDE.md` logging rules.
**Verification**: `grep -r "fmt.Printf\|log.Printf\|logrus\|zerolog" internal/` returns empty.

### IV. Unified Error & Response Handling

- Domain errors: `pkg/errors` — use `apperrors.ErrNotFound`, `apperrors.ErrConflict`, `apperrors.New(code, msg)`, `apperrors.Wrap(err, appErr)`.
- HTTP responses: `internal/interface/response` — use `RespondJSON`, `RespondError`, `RespondSuccess`.

No raw `http.Error()` calls or direct `json.NewEncoder().Encode()` in handlers.

**Source**: `pkg/errors/errors.go`, `internal/interface/response/response.go`.
**Verification**: All handlers use `response.Respond*` functions exclusively.

### V. Test-First (TDD — Non-Negotiable)

All new business logic must follow Red-Green-Refactor:
1. 🔴 Write failing test — run `go test ./...` and confirm failure.
2. 🟢 Write minimal code to pass — run and confirm pass.
3. 🔵 Refactor — run and confirm still passing.

**Test Framework**: stdlib `testing` package. Fakes/stubs via hand-written structs that implement repository interfaces (no mocking frameworks — gomock/testify/mock are not in go.mod).
**Test location**: same package as the code under test, `*_test.go` or `*_integration_test.go` files.
**Coverage requirement**: ≥ 80% for new code; 100% for core domain logic.

**Source**: `internal/application/service/*_integration_test.go`, `internal/domain/**/*_test.go`.
**Verification**: Git history shows test commit precedes implementation commit per feature.

### VI. Observability — OpenTelemetry + GCP

Non-trivial operations must be traceable via OTel spans (Google Cloud Trace). Business metrics are emitted via OTel metrics (Google Cloud Monitoring). Both are initialized in `cmd/api/main.go` and propagated through `context.Context`.

**Source**: `pkg/observability/`, `go.mod` OTel dependencies.
**Verification**: New services wrap critical paths with spans; new metrics registered in observability package.

### VII. GitFlow — Feature Branches from `develop`

Every feature starts from `develop`:
```bash
git checkout develop && git pull origin develop
git checkout -b feature/<descriptive-name>
```

Commit format: `[ADD]`, `[CHANGE]`, `[FIX]`, `[REMOVE]` + English description.
Never commit directly to `master` or `develop`.
Release merges to `master` only from `release/` or `hotfix/` branches via PR.

**Source**: `CLAUDE.md` GitFlow rules, `.github/workflows/`.
**Verification**: All PRs target `develop`; only release/hotfix PRs target `master`.

### VIII. No Code Duplication

Before creating any new type, function, or package, search for existing implementations in `pkg/`, `internal/domain/`, and `internal/application/service/`. Only create new code after confirming nothing reusable exists.

**Source**: DRY principle enforced across shared packages.
**Verification**: Code review checks for duplicate implementations.

---

## Tech Stack Anchoring

**SDD specs and plans must use this stack — no substitutions**:

| Category | Technology | Version | Non-replaceable |
|----------|------------|---------|-----------------|
| Language | Go | 1.24+ (go.mod: 1.25) | ✓ |
| HTTP Router | go-chi/chi | v5.2.4 | ✓ |
| Database | MySQL 8.0 via go-sql-driver | v1.9.3 | ✓ |
| Auth | golang-jwt/jwt | v5.3.1 | ✓ |
| OTel SDK | go.opentelemetry.io/otel | v1.43.0 | ✓ |
| Cloud Trace | opentelemetry-operations-go/trace | v1.32.0 | ✓ |
| Cloud Monitoring | opentelemetry-operations-go/metric | v0.56.0 | ✓ |
| Logging | log/slog (stdlib) | Go stdlib | ✓ |
| Testing | testing (stdlib) | Go stdlib | ✓ |
| Build | make + go build | — | ✓ |
| Deploy | Google App Engine Standard | — | ✓ |

---

## Directory Contract

**All new code must be placed at these paths**:

| Code Type | Standard Location | Naming Convention |
|-----------|-------------------|-------------------|
| HTTP Handler | `internal/interface/handler/<domain>/` | `snake_case.go`, exported type `Handler` |
| Service (business logic) | `internal/application/service/` | `<domain>_service.go` |
| Domain Entity | `internal/domain/entity/` | `snake_case.go`, PascalCase struct |
| Repository Interface | `internal/domain/repository/` | `<domain>_repository.go`, interface `<Domain>Repository` |
| Repository Implementation | `internal/infrastructure/repository/` | `mysql_<domain>_repository.go` |
| Domain Logic (pure functions) | `internal/domain/<domain>/` | `snake_case.go` |
| Shared Utilities | `pkg/<package>/` | `snake_case.go` |
| Config | `pkg/config/` | extend `config.go` |
| DB Migrations | `migrations/` | `NNNN_description.sql` |
| Tests | same package as production code | `<file>_test.go` or `<file>_integration_test.go` |

---

## Governance

### Constitution Priority

1. This constitution is the highest guidance for all SDD workflow in loopi-api.
2. All specs must conform to these principles.
3. All plans must use the anchored tech stack above.
4. All tasks must comply with the directory contract.

### Amendment Procedure

Modifying this constitution requires documenting the change reason and updating all dependent templates.

**Version**: 1.0.0 | **Created**: 2026-05-01 | **Source**: Brownfield Bootstrap
