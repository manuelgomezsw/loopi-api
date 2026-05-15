---
description: "Task list template for loopi-api feature implementation"
---

# Tasks: [FEATURE NAME]

**Feature Branch**: `feature/[descriptive-name]`
**Input**: `specs/[feature-name]/spec.md` + `specs/[feature-name]/plan.md`
**Prerequisites**: plan.md (required), spec.md (required)
**Constitution**: `.specify/memory/constitution.md`

## Format: `[ID] [P?] [US?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[USN]**: User story this task belongs to (US1, US2, …)
- Include exact file paths relative to project root

## Path Conventions (loopi-api — Constitution Locked)

⚠️ **All new files must follow these paths**:

| Code Type | Path |
|-----------|------|
| HTTP Handler | `internal/interface/handler/<domain>/<file>.go` |
| Service | `internal/application/service/<domain>_service.go` |
| Domain Entity | `internal/domain/entity/<entity>.go` |
| Repository Interface | `internal/domain/repository/<domain>_repository.go` |
| Repository Implementation | `internal/infrastructure/repository/mysql_<domain>_repository.go` |
| Domain Logic | `internal/domain/<domain>/<file>.go` |
| Shared Utility | `pkg/<package>/<file>.go` |
| DB Migration | `migrations/NNNN_<description>.sql` |
| Tests | Same package as production code, `<file>_test.go` |

<!--
  ============================================================================
  IMPORTANT: The tasks below are SAMPLE TASKS for illustration purposes only.

  The /speckit-tasks command MUST replace these with actual tasks based on:
  - User stories from spec.md (with their priorities P1, P2, P3…)
  - Plan phases from plan.md
  - loopi-api directory contract from constitution

  TDD is MANDATORY: tests must be written and FAIL before implementation.
  DO NOT keep these sample tasks in the generated tasks.md file.
  ============================================================================
-->

---

## Phase 1: Setup & Verification

**Purpose**: Confirm prerequisites and look for reusable code.

- [ ] T001 Verify working branch is `feature/[name]` branched from `develop`
- [ ] T002 Search for existing code to reuse: `grep -r "[Entity/Method]" internal/ pkg/`
- [ ] T003 Confirm constitution compliance checks from `plan.md` pass
- [ ] T004 [P] Create DB migration `migrations/NNNN_[description].sql` if schema changes needed

---

## Phase 2: Domain & Repository Layer

**Purpose**: Define contracts before implementation.

**⚠️ CRITICAL**: This phase MUST complete before Phases 3–4.

- [ ] T005 [P] Add/update entity in `internal/domain/entity/<entity>.go`
- [ ] T006 [P] Add method to repository interface in `internal/domain/repository/<domain>_repository.go`
- [ ] T007 Add pure domain functions in `internal/domain/<domain>/<file>.go` if needed

**Checkpoint**: Repository interfaces defined — implementation and service can now proceed.

---

## Phase 3: User Story 1 - [Title] (Priority: P1) 🎯 MVP

**Goal**: [Brief description of what this story delivers]
**Independent Test**: `go test ./internal/application/service/... -run TestXxx`

### 🔴 TDD Phase 1: Write Failing Tests

- [ ] T008 [P] [US1] Write fake repo implementing `<Domain>Repository` interface
- [ ] T009 [P] [US1] Write test `TestXxx` in `internal/application/service/<domain>_service_test.go`
- [ ] T010 **Run `go test ./...`, confirm tests FAIL** ← 🔴 Must fail first

### 🟢 TDD Phase 2: Minimal Implementation

- [ ] T011 [US1] Implement repository method in `internal/infrastructure/repository/mysql_<domain>_repository.go`
- [ ] T012 [US1] Implement service method in `internal/application/service/<domain>_service.go`
  - Use `logger.FromContext(ctx)` — log INFO on success, WARN on domain errors, ERROR on DB failures
  - Use `pkg/errors` for all error returns
- [ ] T013 **Run `go test ./...`, confirm tests PASS** ← 🟢 Must pass

### 🔵 TDD Phase 3: Refactor

- [ ] T014 Refactor code, improve readability
- [ ] T015 **Run `go test ./...`, confirm still passing** ← 🔵 Must not break

### Interface Layer

- [ ] T016 [US1] Implement handler method in `internal/interface/handler/<domain>/<handler>.go`
  - Use `response.RespondJSON` / `response.RespondError` / `response.RespondSuccess` only
  - No logging in handlers
- [ ] T017 [US1] Register route in `internal/interface/router/router.go`

### Observabilidad OTel *(si la operación es no-trivial)*

- [ ] T018 [US1] Add OTel span in service: `tracer.Start(ctx, "domain.Method")` + `defer span.End()`
- [ ] T019 [US1] Register business metric in `pkg/observability/metrics.go` if applicable

**Checkpoint**: User Story 1 fully functional — `make test-coverage` shows ≥ 80% for new code.

---

## Phase 4: User Story 2 - [Title] (Priority: P2)

**Goal**: [Brief description of what this story delivers]
**Independent Test**: `go test ./internal/application/service/... -run TestYyy`

### 🔴 TDD Phase 1: Write Failing Tests

- [ ] T018 [P] [US2] Write test `TestYyy` in `internal/application/service/<domain>_service_test.go`
- [ ] T019 **Run `go test ./...`, confirm tests FAIL** ← 🔴 Must fail first

### 🟢 TDD Phase 2: Minimal Implementation

- [ ] T020 [US2] Implement repository method in `internal/infrastructure/repository/mysql_<domain>_repository.go`
- [ ] T021 [US2] Implement service method in `internal/application/service/<domain>_service.go`
- [ ] T022 **Run `go test ./...`, confirm tests PASS** ← 🟢 Must pass

### 🔵 TDD Phase 3: Refactor

- [ ] T023 Refactor and run `go test ./...` to confirm still passing

### Interface Layer

- [ ] T024 [US2] Implement handler in `internal/interface/handler/<domain>/`
- [ ] T025 [US2] Register route in `internal/interface/router/router.go`

### Observabilidad OTel *(si la operación es no-trivial)*

- [ ] T026 [US2] Add OTel span in service: `tracer.Start(ctx, "domain.Method")` + `defer span.End()`

**Checkpoint**: User Stories 1 AND 2 both functional independently.

---

[Add more user story phases following the same TDD pattern]

---

## Phase N: Integration Verification

- [ ] TXXX Run full test suite: `make test`
- [ ] TXXX Run test with coverage: `make test-coverage` — confirm ≥ 80%
- [ ] TXXX Run vet: `go vet ./...`
- [ ] TXXX Run linter: `make lint` (if golangci-lint installed)
- [ ] TXXX Build: `make build` — confirm no compile errors
- [ ] TXXX Manual smoke test via Postman or `curl` if new HTTP endpoints added

---

## Execution Order & Parallel Opportunities

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Domain)**: After Phase 1 — BLOCKS Phases 3+
- **Phase 3+ (User Stories)**: After Phase 2 — can proceed in priority order
- **Phase N (Verification)**: After all user story phases complete

### TDD Enforcement

Within each user story, strict order:
1. Write tests → confirm FAIL 🔴
2. Implement → confirm PASS 🟢
3. Refactor → confirm still passing 🔵

### Parallel Opportunities

- `[P]` tasks within a phase can run in parallel
- Domain entity + repository interface tasks (T005, T006) are parallel
- Test writing tasks for independent stories are parallel once Phase 2 completes

---

## Notes

- Commit after each TDD phase (FAIL, PASS, REFACTOR) for clean git history
- `[P]` = different files, no shared state
- Never log `password`, `password_hash`, `token`, `jwt`, `secret`, or `authorization`
- Handlers: no logging — services capture all events
- Repositories: no business logic — SELECT/INSERT/UPDATE/DELETE only
