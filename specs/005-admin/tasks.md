---
description: "Tareas de implementación para el panel de administración (migrado)"
---

# Tareas: Panel de Administración

**Feature Branch**: `feature/admin` (migrado)
**Estado**: migrated — tareas de implementación completadas; tests y OTel pendientes
**Constitution**: `.specify/memory/constitution.md`

## Convenciones de rutas (loopi-api)

| Tipo de código | Ruta |
|---------------|------|
| Handlers admin | `internal/interface/handler/admin/<domain>_handler.go` |
| Servicios admin | `internal/application/service/admin_<domain>.go` |
| Facade admin | `internal/application/service/admin_service.go` |
| DTOs admin | `internal/application/service/admin_types.go` |
| Repos interfaces | `internal/domain/repository/<domain>_repository.go` |
| Repos impl | `internal/infrastructure/repository/mysql_<domain>_repository.go` |
| Tests | mismo paquete, `<file>_test.go` |

---

## Fase 1: Setup y Dominio

- [x] T001 Verificar migraciones 001, 005, 006, 007, 008, 012 aplicadas
- [x] T002 [P] Entidad `Category` con `display_order` en `internal/domain/entity/category.go`
- [x] T003 [P] Entidad `Supplier` en `internal/domain/entity/supplier.go`
- [x] T004 [P] Entidad `MeasurementUnit` en `internal/domain/entity/measurement_unit.go`
- [x] T005 Interfaz `CategoryRepository` con `UpdateDisplayOrders` (batch) en `internal/domain/repository/category_repository.go`
- [x] T006 Interfaz `SupplierRepository` en `internal/domain/repository/supplier_repository.go`
- [x] T007 Interfaz `MeasurementUnitRepository` en `internal/domain/repository/measurement_unit_repository.go`
- [x] T008 DTOs compartidos en `internal/application/service/admin_types.go`

**Checkpoint**: Interfaces definidas — implementaciones e implementaciones de servicio pueden proceder.

---

## Fase 2: Infraestructura MySQL

- [x] T009 [P] `mysql_category_repository.go` — CRUD + `UpdateDisplayOrders` (batch UPDATE)
- [x] T010 [P] `mysql_supplier_repository.go` — CRUD + `FindAllActive` + filtros
- [x] T011 [P] `mysql_measurement_unit_repository.go` — `FindAll`
- [x] T012 [P] Extender `mysql_employee_repository.go` — `Create`, `Update`, `UpdateStatus`, `ResetPassword`
- [x] T013 [P] Extender `mysql_item_repository.go` — `FindAllWithFilters`, `Create`, `Update`, `UpdateStatus`

---

## Fase 3: User Story 1 — Dashboard (P1) 🎯 MVP

**Goal**: Admin ve estadísticas del día en un solo endpoint.

### 🔴 TDD *(gap — pendiente)*

- [ ] T014 ⚠️ Escribir fake repos para `AdminDashboardService`
- [ ] T015 ⚠️ Escribir `TestGetDashboard_*` en `internal/application/service/admin_dashboard_test.go`
- [ ] T016 ⚠️ **Ejecutar `go test ./...` — confirmar FALLAN**

### 🟢 Implementación

- [x] T017 [US1] `AdminDashboardService.GetDashboard`: count inventarios del día, with/without discrepancies, pending
- [x] T018 [US1] `admin/dashboard_handler.GetDashboard` → `GET /api/admin/dashboard`

**Checkpoint**: Admin puede ver estadísticas del día.

---

## Fase 4: User Story 2 — Gestión de Inventarios Admin (P1)

### 🔴 TDD *(gap — pendiente)*

- [ ] T019 ⚠️ Escribir `TestListInventories_*`, `TestGetInventoryDetail_*` en `admin_inventory_test.go`
- [ ] T020 ⚠️ **Ejecutar `go test ./...` — confirmar FALLAN**

### 🟢 Implementación

- [x] T021 [US2] `AdminInventoryService.ListInventories`: paginación + filtros (fecha, tipo, empleado, hasDiscrepancies)
- [x] T022 [US2] `AdminInventoryService.GetInventoryDetail`: detalle con todos los ítems + responsable
- [x] T023 [US2] `AdminInventoryService.UpdateInventoryDetail`: editar cualquier campo de detalle
- [x] T024 [US2] `AdminInventoryService.CreateInitialInventory`: tipo `initial` para el día actual
- [x] T025 [US2] `AdminInventoryService.GetActiveInventoriesCount`
- [x] T026 [US2] Handlers correspondientes en `admin/inventory_handler.go`
- [x] T027 [US2] Registrar 5 rutas bajo `/api/admin/inventories` en `router.go`

---

## Fase 5: User Story 3 — Gestión de Ítems (P1)

### 🔴 TDD *(gap — pendiente)*

- [ ] T028 ⚠️ Escribir `TestCreateItem_WithAddToActiveInventories` — caso más complejo
- [ ] T029 ⚠️ **Ejecutar `go test ./...` — confirmar FALLAN**

### 🟢 Implementación

- [x] T030 [US3] `AdminItemService.ListItems`: paginación + filtros (`type`, `frequency`, `active`, `search`)
- [x] T031 [US3] `AdminItemService.CreateItem`: si `add_to_active_inventories=true` → buscar inventarios `in_progress` → `CreateBatch` en `inventory_details`
- [x] T032 [US3] `AdminItemService.UpdateItem`, `UpdateItemStatus`, `GetItem`
- [x] T033 [US3] `AdminItemService.ListMeasurementUnits`
- [x] T034 [US3] Handlers en `admin/item_handler.go`
- [x] T035 [US3] Registrar 5 rutas ítems + 1 ruta measurement-units en `router.go`

---

## Fase 6: User Story 4 — Gestión de Empleados (P1)

### 🔴 TDD *(gap — pendiente)*

- [ ] T036 ⚠️ Escribir `TestCreateEmployee_HashesPassword`, `TestResetEmployeePassword_*`
- [ ] T037 ⚠️ **Ejecutar `go test ./...` — confirmar FALLAN**

### 🟢 Implementación

- [x] T038 [US4] `AdminEmployeeService.CreateEmployee`: hashea password con bcrypt → guarda
- [x] T039 [US4] `AdminEmployeeService.UpdateEmployee`, `UpdateEmployeeStatus`, `GetEmployee`, `ListEmployees`, `ListAllActiveEmployees`
- [x] T040 [US4] `AdminEmployeeService.ResetEmployeePassword`: genera contraseña → hashea → guarda
- [x] T041 [US4] Handler en `admin/employee_handler.go`
- [x] T042 [US4] Registrar 7 rutas empleados en `router.go`

---

## Fase 7: User Stories 5 y 6 — Categorías y Proveedores (P2)

### 🟢 Implementación (sin tests — gap)

- [x] T043 [US5] `AdminCategoryService`: CRUD + `ReorderCategories` (batch update `display_order`)
- [x] T044 [US5] `admin/category_handler.go` + 6 rutas en `router.go`
- [x] T045 [US6] `AdminSupplierService`: CRUD + `ListAllActiveSuppliers`
- [x] T046 [US6] `admin/supplier_handler.go` + 6 rutas en `router.go`

---

## Fase 8: Facade y Middleware

- [x] T047 `AdminService` facade en `admin_service.go` — inyecta 6 sub-servicios, expone sus métodos
- [x] T048 `middleware.AdminOnly` aplicado a todo el grupo `/api/admin` en `router.go`

---

## Fase N: Verificación de Integración

- [x] TXXX Ejecutar `make build` — sin errores de compilación
- [x] TXXX Ejecutar `make test` — suite de dominio e inventario pasan
- [x] TXXX Smoke test manual con token admin: verificar que todos los endpoints responden correctamente
- [x] TXXX Smoke test con token employee: verificar HTTP 403 en todos los endpoints `/api/admin/*`

---

## Gaps Identificados

| Gap | Severidad | Acción recomendada |
|-----|-----------|-------------------|
| Sin tests para ningún admin service (7 servicios) | Crítica — 3.200 líneas sin cobertura | Crear feature `006-admin-tests` con `/speckit-specify` |
| Sin OTel spans en operaciones admin | Media — dificulta debugging en producción | Agregar en el próximo ciclo de features admin |
| Sin métricas de negocio en operaciones admin | Media — no hay visibilidad de volumen de cambios | Registrar counters en create/update/delete de entidades principales |
| `CreateEmployee` username duplicado sin error domain explícito | Baja — error de BD expuesto como HTTP 500 | Capturar error de uniqueness MySQL 1062 → `apperrors.ErrConflict` |

---

## Notas

- `AdminService` es puro facade — toda la lógica está en los 6 sub-servicios especializados
- `password_hash` NUNCA en responses ni en logs — verificado en `admin_employee.go`
- `ReorderCategories` hace batch update en una sola llamada a BD para evitar N queries
- El ítem se agrega a inventarios activos solo al crear (`add_to_active_inventories`) — las actualizaciones no tienen este comportamiento
