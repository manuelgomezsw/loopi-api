# Reporte de Discrepancias (Phase 3 — Field-by-Field)

**Generado**: 2026-03-26
**Aplicación**: Loopi
**Fuente de verdad**: Código Go (`internal/domain/entity/` + `internal/interface/router/`)

## Discrepancias README vs Código

### 🔴 CRÍTICO — Endpoints Faltantes en README

| Endpoint | En README | En Código | Veredicto |
|----------|-----------|-----------|-----------|
| `GET /health` | No | Sí | 🔸 CODE_ONLY |
| `GET /api/inventories/latest` | Sí | Sí | ✅✅ VERIFIED |
| `GET /api/inventories/in-progress` | No | Sí | 🔸 CODE_ONLY |
| `GET /api/inventories/suggested-schedule` | No | Sí | 🔸 CODE_ONLY |
| `GET /api/inventories/{id}/discrepancies` | No | Sí | 🔸 CODE_ONLY |
| `GET /api/inventories/{id}/summary` | No | Sí | 🔸 CODE_ONLY |
| `GET /api/admin/dashboard` | No | Sí | 🔸 CODE_ONLY |
| `GET /api/admin/inventories` | No | Sí | 🔸 CODE_ONLY |
| `GET /api/admin/inventories/active-count` | No | Sí | 🔸 CODE_ONLY |
| `POST /api/admin/inventories/initial` | No | Sí | 🔸 CODE_ONLY |
| `GET /api/admin/inventories/{id}` | No | Sí | 🔸 CODE_ONLY |
| `PUT /api/admin/inventories/{id}/details/{detailId}` | No | Sí | 🔸 CODE_ONLY |
| `GET /api/admin/measurement-units` | No | Sí | 🔸 CODE_ONLY |
| Todos los endpoints `/api/admin/items/*` (5) | No | Sí | 🔸 CODE_ONLY |
| Todos los endpoints `/api/admin/employees/*` (7) | No | Sí | 🔸 CODE_ONLY |
| Todos los endpoints `/api/admin/categories/*` (6) | No | Sí | 🔸 CODE_ONLY |
| Todos los endpoints `/api/admin/suppliers/*` (6) | No | Sí | 🔸 CODE_ONLY |

**Conclusión**: README documenta ~10 de 37 endpoints. Las specs generadas usan el código como fuente de verdad.

### 🔴 CRÍTICO — Versión de Go

| Atributo | README dice | go.mod dice | Veredicto |
|----------|-------------|-------------|-----------|
| Go version | "1.21+" | 1.24.0 | 🔸 CODE_ONLY (usar código) |

### 🟡 MEDIA — Fórmulas en docs vs código

| Fórmula | docs/ (histórico) | Código actual | Veredicto |
|---------|-------------------|---------------|-----------|
| `expected_at_end` | `suggested + stock_received - units_sold` | `suggested + stock_received - units_sold` (clamp >= 0) | ✅✅ VERIFIED |
| `expected_for_admin` | `suggested - shrinkage + stock_received - units_sold` | `suggested - shrinkage + stock_received - units_sold` (clamp >= 0) | ✅✅ VERIFIED |
| Discrepancia | `real != expected_at_end` | `real != expected_at_end` | ✅✅ VERIFIED |
| Sugerido siguiente | `real_anterior` (sin restar mermas) | Enricher: `suggested_value = real_value del anterior` | ✅✅ VERIFIED |

### 🟡 MEDIA — `inventory_issues` en docs vs código

| Elemento | Mencionado en docs | En código | Veredicto |
|----------|-------------------|-----------|-----------|
| Tabla `inventory_issues` | Sí (como feature existente) | ❌ ELIMINADA (migración 011) | 🔸 CODE_ONLY — docs desactualizados |
| Entidad `InventoryIssue` | Sí | ❌ No existe | Ignorar docs, usar código |

### 🟡 MEDIA — Turnos en docs vs código

| Elemento | Planificado en docs | Implementado en código | Veredicto |
|----------|--------------------|-----------------------|-----------|
| 4 turnos (morning_open/close, afternoon_open/close) | Sí (INTEGRACION_TURNOS_EN_REFACTOR.md) | ❌ No implementado. Código usa: opening/noon/closing | ⚠️ DOCS_ONLY (planificado, no implementado) |
| Migracion ENUM de schedule | Planificada | ❌ No ejecutada | ⚠️ DOCS_ONLY |

**Los docs de turnos son planes futuros, NO reflejan el estado actual del sistema.**

### 🟢 BAJA — Discrepancias menores

| Elemento | Docs | Código | Veredicto |
|----------|------|--------|-----------|
| Timezone | Implícita | `America/Bogota` hardcoded | 🔸 CODE_ONLY |
| Moneda | No mencionada | COP (sin decimales), `cost` como INT UNSIGNED | 🔸 CODE_ONLY |
| `bcrypt cost` | No mencionado | 12 en JWTManager / DefaultCost en admin | 🔸 CODE_ONLY |

## Validación de Campos de Entidades (vs ARCHITECTURE.md)

### Employee

| Campo | ARCHITECTURE.md | Código | Veredicto |
|-------|----------------|--------|-----------|
| id | Mencionado | `uint16` | ✅✅ VERIFIED |
| username | Mencionado | `string` UNIQUE | ✅✅ VERIFIED |
| role | ENUM employee/admin | `Role` ENUM | ✅✅ VERIFIED |
| Todos los demás campos | No documentados | Sí en código | 🔸 CODE_ONLY |

### InventoryDetail

| Campo | ARCHITECTURE.md | Código | Veredicto |
|-------|----------------|--------|-----------|
| suggested_value | Sí | `*uint16` | ✅✅ VERIFIED |
| real_value | Sí | `*uint16` | ✅✅ VERIFIED |
| stock_received | Sí | `*uint16` | ✅✅ VERIFIED |
| units_sold | Sí | `*uint16` | ✅✅ VERIFIED |
| shrinkage | Mencionado en docs | `*uint16` en código | ✅✅ VERIFIED |

## Items de Acción

| ID | Prioridad | Acción |
|----|-----------|--------|
| ACT-001 | 🔴 Alta | Actualizar README del backend con los 37 endpoints reales |
| ACT-002 | 🔴 Alta | Agregar tests unitarios (dominio) y de integración (servicios) |
| ACT-003 | 🟡 Media | Evaluar estado de `GetRecentDiscrepancies` y `GetDashboardStats` — ¿usan definición correcta? |
| ACT-004 | 🟡 Media | Decidir fecha de implementación de Feature Turnos (4 momentos) |
| ACT-005 | 🟢 Baja | Documentar API errors (códigos HTTP, mensajes) |
