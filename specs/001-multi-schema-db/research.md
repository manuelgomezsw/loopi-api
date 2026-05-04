# Research: Multi-Schema Database Configuration

**Feature**: 001-multi-schema-db
**Date**: 2026-05-03

---

## Decision 1: Mechanism for schema selection

**Decision**: Add a new `APP_ENV` environment variable (`development` | `test` | `production`) that drives schema selection via a `resolveDBName()` mapping function in `pkg/config/config.go`. `DB_NAME` is no longer a direct env var — it is derived.

**Rationale**: `APP_ENV` prevents accidental misconfiguration (a developer cannot accidentally set `DB_NAME=loopi_prod` in their local `.env`). The mapping is deterministic and validated at startup. Explicit env vars for dev and prod are simpler to audit in `app.yaml` and `.env.example`. *(Decision revised from initial DB_NAME-only approach after checklist review — CHK008/CHK009.)*

**Mapping**:
- `development` → `loopi_dev`
- `test` → `loopi_dev`
- `production` → `loopi_prod`
- anything else → startup error with value in message

**Alternatives considered**:
- `DB_NAME` directly: Simpler but allows misconfiguration — rejected for safety reasons.
- Separate config file per environment: Overkill for a single variable change.

---

## Decision 2: Startup validation (fail-fast)

**Decision**: `config.Load()` calls `getEnvRequired("APP_ENV")` — returns `fmt.Errorf("required environment variable \"APP_ENV\" is not set")` if unset or empty. If `APP_ENV` is set but unknown, `resolveDBName()` returns a descriptive error naming the invalid value and the valid options.

**Rationale**: Two layers of validation — missing variable and invalid value — give operators actionable error messages. The error message names the variable and, for invalid values, lists the valid options.

**Alternatives considered**:
- Validate `DB_NAME` instead: Rejected per Decision 1 (DB_NAME is now derived).

---

## Decision 3: Schema creation / rename strategy

**Decision**: Create `loopi_dev` as a new schema. For `loopi_prod`, create it as a new schema and re-run all migrations against it (do not attempt to rename `loopi` — MySQL has no `RENAME DATABASE`). The existing `loopi` schema remains untouched until the team confirms the migration is complete.

**Rationale**: MySQL 8.0 does not support renaming a database directly. The safest path is: create `loopi_prod`, apply all migrations, update `app.yaml`, deploy, verify, then drop the old `loopi` schema when confident. This approach is zero-downtime and reversible.

**Alternatives considered**:
- Dump-and-restore `loopi` → `loopi_prod`: Valid but requires manual data migration. Acceptable for production cutover but out of scope for this spec (which focuses on configuration isolation).
- Keep `loopi` as prod and only create `loopi_dev`: Simpler short-term but diverges from the naming convention and creates confusion.

---

## Decision 4: Where new schemas live

**Decision**: Both schemas (`loopi_dev` and `loopi_prod`) are created in the **same existing Cloud SQL instance** — no new Cloud SQL instance needed.

**Rationale**: MySQL schemas are logical namespaces within a single instance. Using one instance minimizes cost and operational overhead. Access control is handled by the `DB_USER` already configured.

**Alternatives considered**:
- Separate Cloud SQL instances per environment: More isolation but significantly higher cost for a dev/test schema.

---

## Decision 5: Migration scripts location

**Decision**: Add a new SQL script `migrations/013_create_schemas.sql` (informational) that documents the `CREATE SCHEMA IF NOT EXISTS` commands for both schemas. Actual schema creation is done manually by the operator as a one-time setup step, not by the application at startup.

**Rationale**: The application should not auto-create its own database schema at startup — that is an ops concern. The script serves as documentation and reproducible setup instruction.

---

## No NEEDS CLARIFICATION items

All unknowns resolved through codebase analysis. Implementation is clear:
1. Remove `DB_NAME` default from config
2. Add `DB_NAME` validation in `config.Load()`
3. Update `app.yaml` → `DB_NAME=loopi_prod`
4. Update `.env.example` → `DB_NAME=loopi_dev`
5. Add migration script documenting schema creation
6. Write tests for config validation
