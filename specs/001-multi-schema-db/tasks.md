---
description: "Task list for multi-schema database configuration"
---

# Tasks: Multi-Schema Database Configuration

**Feature Branch**: `feature/001-multi-schema-db`
**Input**: `specs/001-multi-schema-db/spec.md` + `specs/001-multi-schema-db/plan.md`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅
**Constitution**: `.specify/memory/constitution.md`

## Format: `[ID] [P?] [US?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[USN]**: User story this task belongs to
- File paths are relative to project root

---

## Phase 1: Setup & Verification

**Purpose**: Confirm branch, find reusable code, validate no conflicts.

- [x] T001 Verify current branch is `feature/001-multi-schema-db` branched from `develop`; if in worktree branch `claude/xxx`, run `git checkout develop && git pull origin develop && git checkout -b feature/001-multi-schema-db`
- [x] T002 Search for existing `APP_ENV`, `getEnvRequired`, `resolveDBName` in codebase: `grep -r "APP_ENV\|getEnvRequired\|resolveDBName" pkg/ internal/` — confirm none exist before adding
- [x] T003 Confirm `pkg/config/config.go` currently uses `getEnv("DB_NAME", "loopi")` as the only schema config — note line number for edit

---

## Phase 2: User Story 1 — APP_ENV-driven schema selection (Priority: P1) 🎯 MVP

**Goal**: Application reads `APP_ENV` at startup, maps it to the correct MySQL schema, and fails with a descriptive error if `APP_ENV` is missing or unrecognized.

**Independent Test**: `go test ./pkg/config/... -v`

### 🔴 TDD: Write failing tests first

- [x] T004 [US1] Create `pkg/config/config_test.go` with failing tests for `resolveDBName()`:
  - `TestResolveDBName_Development` → expects `"loopi_dev"`, no error
  - `TestResolveDBName_Test` → expects `"loopi_dev"`, no error
  - `TestResolveDBName_Production` → expects `"loopi_prod"`, no error
  - `TestResolveDBName_Unknown` → expects error containing the invalid value and valid options
  - Run `go test ./pkg/config/...` and confirm compilation failure (function not yet defined)

- [x] T005 [US1] Add failing tests for `Load()` in `pkg/config/config_test.go`:
  - `TestLoad_MissingAppEnv` → unset `APP_ENV`, expects error containing `"APP_ENV"`
  - `TestLoad_InvalidAppEnv` → `APP_ENV=staging`, expects error containing `"staging"` and valid options
  - `TestLoad_DevelopmentEnv` → `APP_ENV=development`, expects `DatabaseConfig.Name == "loopi_dev"`
  - `TestLoad_ProductionEnv` → `APP_ENV=production`, expects `DatabaseConfig.Name == "loopi_prod"`
  - Run `go test ./pkg/config/...` and confirm tests compile but fail (Load still uses DB_NAME)

### 🟢 Implementation

- [x] T006 [US1] Add `getEnvRequired(key string) (string, error)` to `pkg/config/config.go`:
  - Returns `fmt.Errorf("required environment variable %q is not set", key)` if `os.LookupEnv` returns empty or not found
  - Place after existing `getEnv` helper

- [x] T007 [US1] Add `resolveDBName(appEnv string) (string, error)` to `pkg/config/config.go`:
  - `"development"` → `"loopi_dev"`, nil
  - `"test"` → `"loopi_dev"`, nil
  - `"production"` → `"loopi_prod"`, nil
  - default → `fmt.Errorf("unknown APP_ENV value %q: must be one of: development, test, production", appEnv)`

- [x] T008 [US1] Update `Load()` in `pkg/config/config.go`:
  - Read `appEnv, err := getEnvRequired("APP_ENV")` — propagate error on failure
  - Call `dbName, err := resolveDBName(appEnv)` — propagate error on failure
  - Set `DatabaseConfig.Name: dbName` (replaces `getEnv("DB_NAME", "loopi")`)
  - Remove `DB_NAME` from `DatabaseConfig` env var loading

- [x] T009 [US1] Run `go test ./pkg/config/... -v` — confirm all T004/T005 tests pass

### 🔵 Refactor

- [x] T010 [US1] Review `pkg/config/config.go` for clarity: ensure `resolveDBName` and `getEnvRequired` are placed logically near `getEnv`; confirm no dead code left from removed `DB_NAME` default; run `go test ./pkg/config/...` again to confirm still passing

---

## Phase 3: User Story 2 — Schema setup documentation (Priority: P2)

**Goal**: Both schemas (`loopi_dev`, `loopi_prod`) can be reproducibly created on Cloud SQL; environment configs are updated to use `APP_ENV`.

**Independent Test**: Manual review of `migrations/013_create_schemas.sql` and verification that `app.yaml` + `.env.example` reference `APP_ENV`.

- [x] T011 [US2] Create `migrations/013_create_schemas.sql` with the full operator runbook as comments:
  ```sql
  -- =============================================================
  -- One-time operator setup: run as DBA user on Cloud SQL instance
  -- IDEMPOTENT: safe to run multiple times (IF NOT EXISTS)
  -- =============================================================

  CREATE SCHEMA IF NOT EXISTS loopi_dev
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

  CREATE SCHEMA IF NOT EXISTS loopi_prod
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

  -- SETUP STEPS (run after this script):
  -- 1. Dump existing data:
  --    mysqldump -u <user> -p loopi > loopi_backup.sql
  -- 2. Restore into loopi_dev:
  --    mysql -u <user> -p loopi_dev < loopi_backup.sql
  -- 3. Restore into loopi_prod:
  --    mysql -u <user> -p loopi_prod < loopi_backup.sql
  -- 4. Update app.yaml: APP_ENV=production (remove DB_NAME) and deploy
  -- 5. Verify connectivity and data in both schemas
  -- 6. Drop legacy schema ONLY after verification:
  --    DROP SCHEMA loopi;

  -- ROLLBACK PROCEDURE (if steps 3-4 fail before deployment):
  --   DROP SCHEMA IF EXISTS loopi_prod;
  --   DROP SCHEMA IF EXISTS loopi_dev;
  --   (original `loopi` schema is untouched and remains active)
  ```

- [x] T012 [P] [US2] Add a `## Database Setup` section to `README.md` explaining:
  - The two-schema model: `loopi_dev` (development/test) and `loopi_prod` (production)
  - How to set `APP_ENV` locally (`development`) and in production (`production`)
  - Reference to `migrations/013_create_schemas.sql` for Cloud SQL setup
  - The rollback procedure if initialization fails

- [x] T013 [P] [US2] Update `app.yaml`:
  - Add `APP_ENV: "production"` to `env_variables`
  - Remove `DB_NAME: "loopi"` from `env_variables` (now derived from APP_ENV)

- [x] T014 [P] [US2] Update `.env.example`:
  - Add `APP_ENV=development   # development | test | production` near the Database section
  - Remove `DB_NAME=loopi` line (now derived from APP_ENV)
  - Add comment: `# APP_ENV=development → loopi_dev | APP_ENV=production → loopi_prod`

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: Full regression check and build validation.

- [x] T015 Run full test suite: `go test ./...` — confirm zero failures and no regressions in existing tests
- [x] T016 Run build: `go build ./cmd/api/...` — confirm clean compilation
- [x] T017 [P] Verify `app.yaml` does not contain any `DB_NAME` reference: `grep "DB_NAME" app.yaml` must return empty
- [x] T018 [P] Verify `.env.example` does not contain `DB_NAME`: `grep "DB_NAME" .env.example` must return empty
- [x] T019 Update `specs/001-multi-schema-db/data-model.md` — change "Configuration Model Change" section to reflect final implementation (`resolveDBName` mapping `APP_ENV` to schema name)

---

## Dependency Graph

```
T001 → T002 → T003
                ↓
              T004 → T005 (tests, can both be written before any impl)
                          ↓
              T006 → T007 → T008 → T009 → T010
                                              ↓
              T011 ──┐                        ↓
              T012 ──┼──────────────────── T015 → T016 → T017 → T018 → T019
              T013 ──┤
              T014 ──┘
```

T011–T014 son independientes de la Fase 2 y pueden correr en paralelo desde T003.

---

## Parallel Execution Opportunities

| Parallel Group | Tasks | Condition |
|----------------|-------|-----------|
| Test authoring | T004, T005 | Can both be written before any impl code exists |
| Implementation helpers | T006, T007 | No dependency between each other; both needed before T008 |
| US2 config updates | T011, T012, T013, T014 | Fully independent of US1 implementation |
| Final checks | T017, T018 | Can run in parallel after T016 |

---

## Implementation Strategy

**MVP** (User Story 1 only): Complete T001–T010. This delivers the core behaviour — `APP_ENV` drives schema selection with fail-fast validation.

**Full delivery**: Add T011–T019 to complete migration script, README documentation, environment config updates, and regression checks.

**Suggested commit points**:
- After T010: `[ADD] config: APP_ENV-driven schema selection with resolveDBName`
- After T014: `[CHANGE] config: update app.yaml and .env.example for multi-schema setup`
- After T019: `[ADD] migrations: 013 create loopi_dev and loopi_prod schemas with runbook`
