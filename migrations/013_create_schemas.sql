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

-- =============================================================
-- SETUP STEPS (run after this script):
-- =============================================================
-- 1. Dump existing data from the legacy schema:
--    mysqldump -u <user> -p loopi > loopi_backup.sql
--
-- 2. Restore into loopi_dev:
--    mysql -u <user> -p loopi_dev < loopi_backup.sql
--
-- 3. Restore into loopi_prod:
--    mysql -u <user> -p loopi_prod < loopi_backup.sql
--
-- 4. Update app.yaml: add APP_ENV=production, remove DB_NAME line.
--    Deploy to App Engine.
--
-- 5. Verify connectivity and data integrity in both schemas.
--
-- 6. Drop the legacy schema ONLY after verification is complete:
--    DROP SCHEMA loopi;
--
-- =============================================================
-- ROLLBACK PROCEDURE (if steps 3-4 fail before deployment):
-- =============================================================
--   DROP SCHEMA IF EXISTS loopi_prod;
--   DROP SCHEMA IF EXISTS loopi_dev;
--   (the original `loopi` schema is untouched and remains active)
-- =============================================================
