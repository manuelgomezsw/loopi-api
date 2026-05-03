# Configuration Requirements Checklist: Multi-Schema Database Configuration

**Purpose**: Author self-check — validates requirements completeness and clarity before starting implementation
**Created**: 2026-05-03
**Resolved**: 2026-05-03
**Feature**: [spec.md](../spec.md) | [plan.md](../plan.md)

---

## Requirement Completeness

- [x] CHK001 — `APP_ENV` es la variable correcta para FR-004. El plan y la spec están alineados. [Resolved]
- [x] CHK002 — No aplica: `APP_ENV` solo acepta `development`, `test`, `production`. Valores fuera del set causan error en `resolveDBName()`. [Resolved — N/A]
- [x] CHK003 — Las credenciales (`DB_USER`, `DB_PASSWORD`) son compartidas entre `loopi_dev` y `loopi_prod`. Documentado en FR-005 y Assumptions. [Resolved]
- [x] CHK004 — Ambos schemas se inicializan con dump completo de `loopi`. Nuevo requisito FR-007 agregado a la spec. [Resolved]
- [x] CHK005 — El schema `loopi` se elimina tras verificar que los dos nuevos funcionan. Nuevo requisito FR-008 agregado. [Resolved]

---

## Requirement Clarity

- [x] CHK006 — FR-004 actualizado: el error debe incluir el nombre de la variable y la lista de valores válidos. [Resolved]
- [x] CHK007 — FR-006 actualizado: se requiere script SQL (`migrations/013_create_schemas.sql`) + sección en README. Tarea T012 agregada. [Resolved]
- [x] CHK008 — `APP_ENV` es el mecanismo (Option B). Spec y plan alineados. [Resolved 2026-05-03]
- [x] CHK009 — Acceptance scenarios de US1 son consistentes con la implementación planificada. [Resolved 2026-05-03]

---

## Requirement Consistency

- [x] CHK010 — SC-003 está cubierto por el mecanismo de FR-001/FR-003: `APP_ENV=production` → `loopi_prod` hace imposible acceder a `loopi_dev` desde prod. [Resolved]
- [x] CHK011 — Credenciales compartidas confirmadas. Assumptions actualizado. [Resolved]
- [x] CHK012 — Valores válidos confirmados: `development`, `test`, `production`. FR-001 actualizado con la lista explícita. [Resolved]

---

## Acceptance Criteria Quality

- [x] CHK013 — SC-004 aceptado por diseño: el cambio es mínimo (leer una variable extra). Sin impacto medible en tiempo de arranque. [Resolved — accepted by design]
- [x] CHK014 — Resuelto por CHK008/CHK009. SC-002 es válido tal como está. [Resolved]
- [x] CHK015 — `IF NOT EXISTS` en el script SQL es suficiente garantía de idempotencia para US2 scenario 2. [Resolved]

---

## Scenario Coverage

- [x] CHK016 — Rollback documentado en `migrations/013_create_schemas.sql` como comentario y en README. Nuevo requisito FR-009 agregado. [Resolved]
- [x] CHK017 — Solo ejecución local por ahora. No hay CI/CD configurado. Documentado en Assumptions. [Resolved — out of scope]
- [x] CHK018 — Fuera de scope: MySQL maneja DDL con locks; ejecución concurrente de migraciones en dev no requiere requisito. [Resolved — out of scope]

---

## Dependencies & Assumptions

- [x] CHK019 — Cloud SQL ya provisionado con el schema `loopi` activo. Confirmado. [Resolved]
- [x] CHK020 — Acceso DBA disponible para ejecutar `CREATE SCHEMA`. Confirmado. [Resolved]
- [x] CHK021 — Nombres `loopi_dev` y `loopi_prod` confirmados. [Resolved]

---

## Notes

- Todos los ítems resueltos el 2026-05-03.
- Cambios derivados del checklist: FR-004, FR-005, FR-006 actualizados; FR-007, FR-008, FR-009 nuevos; Assumptions actualizados; T012 (README) y T011 (runbook completo) actualizados en tasks.md.
- ✅ **Checklist completo — listo para implementación.**
