---
description: "Tareas de implementación para el flujo de inventario del empleado (migrado)"
---

# Tareas: Inventario del Empleado

**Feature Branch**: `feature/employee-inventory` (migrado)
**Estado**: migrated — todas las tareas completadas excepto gaps marcados
**Constitution**: `.specify/memory/constitution.md`

## Convenciones de rutas (loopi-api)

| Tipo de código | Ruta |
|---------------|------|
| Handler | `internal/interface/handler/employee/inventory_handler.go` |
| Service | `internal/application/service/inventory_service.go` |
| Facade | `internal/application/service/employee_service.go` |
| Dominio puro | `internal/domain/inventory/` |
| Repos interfaces | `internal/domain/repository/inventory_repository.go` |
| Repos impl | `internal/infrastructure/repository/mysql_inventory_repository.go` |
| Tests | mismo paquete que el código, `*_test.go` |

---

## Fase 1: Setup y Verificación

- [x] T001 Verificar migraciones 001, 003, 009, 010 aplicadas en la BD
- [x] T002 Confirmar que no hay lógica de negocio en repositorios
- [x] T003 [P] Crear migración base `migrations/001_initial_schema.up.sql` — tablas `inventories`, `inventory_details`

---

## Fase 2: Dominio y Repositorios

**⚠️ CRÍTICO**: Esta fase DEBE completarse antes de Fases 3+.

- [x] T004 [P] Entidad `Inventory` con métodos `IsCompleted`, `IsInitial`, `RequiresSalesAndPurchases`, `RequiresPurchasesOnly` en `internal/domain/entity/inventory.go`
- [x] T005 [P] Entidad `InventoryDetail` con método `IsComplete` en `internal/domain/entity/inventory.go`
- [x] T006 [P] Funciones puras en `internal/domain/inventory/expected.go`: `ExpectedAtEnd(d)`
- [x] T007 [P] Funciones puras en `internal/domain/inventory/discrepancy.go`: `HasDiscrepancyFromExpectedEnd(d, expected)`
- [x] T008 Interfaces `InventoryRepository` e `InventoryDetailRepository` en `internal/domain/repository/inventory_repository.go`

**Checkpoint**: Interfaces definidas — implementación puede proceder.

---

## Fase 3: User Story 1 — Crear Inventario (P1) 🎯 MVP

**Goal**: Empleado crea inventario; sistema pre-puebla ítems con valores sugeridos.
**Test**: `go test ./internal/domain/inventory/... -run TestExpected`

### 🔴 TDD Fase 1: Tests de dominio

- [x] T009 [P] Tests de `ExpectedAtEnd` en `internal/domain/inventory/expected_test.go`
- [x] T010 [P] Tests de `HasDiscrepancyFromExpectedEnd` en `internal/domain/inventory/discrepancy_test.go`
- [x] T011 **Ejecutar `go test ./...` — confirmar FALLAN**

### 🟢 TDD Fase 2: Implementación de dominio

- [x] T012 [US1] Implementar `expected.go` y `discrepancy.go`
- [x] T013 **Ejecutar `go test ./...` — confirmar PASAN**

### 🔴 TDD Tests de servicio *(gap — pendiente)*

- [ ] T014 ⚠️ Escribir `TestCreateInventory_*` en `internal/application/service/inventory_service_test.go`
- [ ] T015 ⚠️ **Ejecutar `go test ./...` — confirmar FALLAN**

### 🟢 Implementación de servicio

- [x] T016 [US1] `InventoryService.CreateInventory`: validación → check duplicado → create → `prepopulateInventoryDetails` en `internal/application/service/inventory_service.go`
- [x] T017 [US1] `prepopulateInventoryDetails`: `FindActiveByInventoryType` → `FindPreviousInventory` → `CreateBatch` con `suggested_value = real_value_anterior`
- [x] T018 [US1] `InventoryService.GetSuggestedSchedule`: lógica por hora del día (06–11 opening, 11–16 noon, 16–22 closing)

### Interface Layer

- [x] T019 [US1] `inventory_handler.Create`: decode + validate → `service.CreateInventory` → HTTP 201
- [x] T020 [US1] `inventory_handler.GetSuggestedSchedule` → HTTP 200
- [x] T021 [US1] Registrar `POST /api/inventories` y `GET /api/inventories/suggested-schedule` en `router.go`

### Observabilidad OTel

- [x] T022 [US1] Metric `loopi.inventory.movements` con `action=create` y `type=<tipo>` en `inventory_service.go`

**Checkpoint**: Empleado puede crear inventario y ver ítems pre-poblados.

---

## Fase 4: User Story 3 — Contar y Guardar Ítems (P1)

**Goal**: Empleado registra conteo físico ítem por ítem.

### 🔴 TDD *(gap — pendiente)*

- [ ] T023 ⚠️ Escribir `TestSaveInventoryDetail_*` (inventario completo, ítem no encontrado, inventario no existe)
- [ ] T024 ⚠️ **Ejecutar `go test ./...` — confirmar FALLAN**

### 🟢 Implementación

- [x] T025 [US3] `InventoryService.SaveInventoryDetail`: verifica no completado → busca detalle → actualiza `real_value`
- [x] T026 [US3] `inventory_handler.SaveDetail` → HTTP 200 `{saved:true, suggested_value}`
- [x] T027 [US3] `inventory_handler.GetItems` → HTTP 200 con lista de ítems y su estado

**Checkpoint**: Empleado puede registrar conteos.

---

## Fase 5: User Story 4 — Ventas y Compras (P2)

### 🟢 Implementación (sin tests — gap)

- [x] T028 [US4] `InventoryService.SaveSalesAndPurchases`: verifica tipo → fuerza `units_sold=0` si `RequiresPurchasesOnly` → recalcula `suggested_value`
- [x] T029 [US4] `inventory_handler.SaveSales` → HTTP 200

---

## Fase 6: User Story 5 — Discrepancias y Completar (P1)

### 🔴 TDD

- [x] T030 [P] [US5] Tests de integración en `inventory_service_integration_test.go`:
  - `TestGetDiscrepancies_ReturnsOnlyDetailsWithRealNeExpected`
  - `TestCompleteInventory_DiscrepancyCountMatchesDomain`
  - `TestCompleteInventory_ConsistencyCheck`
- [x] T031 **Ejecutar `go test ./...` — confirmar FALLAN**

### 🟢 Implementación

- [x] T032 [US5] `InventoryService.GetDiscrepancies`: filtra con `HasDiscrepancyFromExpectedEnd`
- [x] T033 [US5] `InventoryService.CompleteInventory`: verifica todos completos → cuenta discrepancias → marca completed → log INFO + metric OTel
- [x] T034 **Ejecutar `go test ./...` — confirmar PASAN**

### 🔵 Refactor

- [x] T035 Extraer `Enricher` a `internal/domain/inventory/enrich.go` para reutilización en admin

### Interface Layer

- [x] T036 [US5] `inventory_handler.GetDiscrepancies`, `GetSummary`, `Complete` — HTTP 200/400
- [x] T037 [US5] Registrar rutas de discrepancias y complete en `router.go`

### Observabilidad OTel

- [x] T038 [US5] Metric `loopi.inventory.movements` con `action=complete`

**Checkpoint**: Flujo completo crear → contar → completar funciona.

---

## Fase N: Verificación de Integración

- [x] TXXX Ejecutar `make test` — suite completa pasa
- [x] TXXX Ejecutar `make test-coverage` — cobertura de dominio ≥ 80%
- [x] TXXX Ejecutar `go vet ./...`
- [x] TXXX Ejecutar `make build` — sin errores de compilación

---

## Gaps Identificados

| Gap | Severidad | Acción recomendada |
|-----|-----------|-------------------|
| Sin tests para `CreateInventory` | Alta — lógica compleja de pre-población | Crear feature de tests con `/speckit-specify` |
| Sin tests para `SaveInventoryDetail` y `SaveSalesAndPurchases` | Alta | Ídem |
| Sin OTel spans en operaciones individuales | Baja | Agregar al próximo feature que toque este servicio |

---

## Notas

- `EmployeeService` es un facade — no contiene lógica propia, solo delega a `InventoryService`
- `IsComplete()` en `InventoryDetail` es la única fuente de verdad sobre si un ítem está listo para completar
- El campo `shrinkage` existe en la BD pero no afecta `expected_at_end` — es informativo
