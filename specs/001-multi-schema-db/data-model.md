# Data Model: Multi-Schema Database Configuration

**Feature**: 001-multi-schema-db
**Date**: 2026-05-03

---

## Schema Topology

No new entities or tables are introduced. This feature changes the **database schema name** used per environment.

| Schema Name  | Environment(s) | Created by          |
|--------------|----------------|---------------------|
| `loopi_dev`  | development, test | Manual setup (see migration script) |
| `loopi_prod` | production     | Manual setup (see migration script) |
| `loopi`      | (legacy)       | Existing — kept until cutover confirmed |

---

## No New Entities

This feature does not add, remove, or modify any domain entities or tables. All existing tables and their schemas remain unchanged — only the database/schema **name** changes per environment.

---

## Configuration Model Change

### Before

```
DatabaseConfig.Name = getEnv("DB_NAME", "loopi")   // same schema for all environments
```

### After (implemented)

```go
appEnv, err := getEnvRequired("APP_ENV")   // fails if APP_ENV unset/empty
dbName, err := resolveDBName(appEnv)       // maps env → schema name
DatabaseConfig.Name = dbName
```

`resolveDBName` mapping:

| `APP_ENV` value | `DatabaseConfig.Name` |
|-----------------|-----------------------|
| `development`   | `loopi_dev`           |
| `test`          | `loopi_dev`           |
| `production`    | `loopi_prod`          |
| anything else   | error with valid options listed |

`DB_NAME` is no longer an environment variable — schema name is always derived from `APP_ENV`.

---

## Migration Script Reference

File: `migrations/013_create_schemas.sql`

Purpose: Documents the one-time Cloud SQL setup commands for operators. Not executed by the application at startup.

```sql
-- Run once on Cloud SQL instance as DBA user
CREATE SCHEMA IF NOT EXISTS loopi_dev
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

CREATE SCHEMA IF NOT EXISTS loopi_prod
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

-- After creating loopi_prod, apply all existing migrations against it:
-- USE loopi_prod; source migrations/001_initial_schema.up.sql; ...
```

---

## Files Changed (Summary)

| File | Change |
|------|--------|
| `pkg/config/config.go` | Remove `DB_NAME` default; add `getEnvRequired` helper with error on missing |
| `app.yaml` | Change `DB_NAME: "loopi"` → `DB_NAME: "loopi_prod"` |
| `.env.example` | Change `DB_NAME=loopi` → `DB_NAME=loopi_dev` |
| `migrations/013_create_schemas.sql` | New — operator setup script |
| `pkg/config/config_test.go` | New — unit tests for `getEnvRequired` and `Load()` validation |
