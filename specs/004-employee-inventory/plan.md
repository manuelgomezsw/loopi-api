# Plan de Implementación: Inventario del Empleado

**Branch**: `feature/employee-inventory` (migrado) | **Fecha**: 2026-05-14 | **Spec**: `specs/004-employee-inventory/spec.md`

## Resumen

Flujo completo de inventario para empleados operativos: crear inventario (con pre-población automática de ítems y valores sugeridos), registrar conteos físicos, registrar ventas/compras, consultar discrepancias y completar el inventario. La lógica de cálculo de valor esperado (`expected_at_end`) y detección de discrepancias vive en el dominio como funciones puras.

## Contexto Técnico (loopi-api — Bloqueado por Constitution)

| Categoría | Tecnología | Versión |
|-----------|------------|---------|
| Lenguaje | Go | 1.24+ |
| HTTP Router | go-chi/chi | v5.2.4 |
| Base de datos | MySQL 8.0 via go-sql-driver | v1.9.3 |
| Observabilidad | OpenTelemetry SDK | v1.43.0 |
| Logging | log/slog (stdlib) | stdlib |
| Testing | testing (stdlib) + fakes escritos a mano | stdlib |

## Verificación de Cumplimiento Constitution

- [x] **Clean Architecture**: Handler en `internal/interface/handler/employee/`, service en `internal/application/service/`, lógica de dominio en `internal/domain/inventory/`
- [x] **Repository Contracts**: Todas las interfaces en `internal/domain/repository/` antes de implementar en `internal/infrastructure/repository/`
- [x] **slog Only**: Logs con `logger.FromContext(ctx)` en servicio, no en handler ni repositorio
- [x] **Error Handling**: `apperrors.New(code, msg)` y `apperrors.ErrNotFound` para errores de dominio
- [x] **Sin Duplicación**: Lógica de expected/discrepancy en `domain/inventory/` — reutilizada por service y tests
- [x] **Directory Contract**: Todos los archivos en sus rutas correctas

## Estructura del Proyecto

### Documentación

```text
specs/004-employee-inventory/
├── spec.md      # especificación funcional
├── plan.md      # este archivo
└── tasks.md     # tareas de implementación
```

### Archivos Implementados

```text
internal/
├── domain/
│   ├── entity/
│   │   ├── inventory.go            # Inventory, InventoryDetail, InventoryType, Schedule, InventoryStatus
│   │   └── item.go                 # Item, ItemType, InventoryFrequency
│   ├── repository/
│   │   ├── inventory_repository.go # InventoryRepository, InventoryDetailRepository
│   │   └── item_repository.go      # ItemRepository.FindActiveByInventoryType
│   └── inventory/                  # LÓGICA DE DOMINIO PURA
│       ├── expected.go             # ExpectedAtEnd(d) → uint16
│       ├── discrepancy.go          # HasDiscrepancyFromExpectedEnd(d, expected)
│       ├── enrich.go               # Enricher: calcula stats de inventarios con subqueries
│       ├── expected_test.go        # tests de expected_at_end
│       ├── discrepancy_test.go     # tests de has_discrepancy
│       └── consistency_test.go     # cross-check: GetDiscrepancies == CompleteInventory count
├── infrastructure/
│   └── repository/
│       ├── mysql_inventory_repository.go        # ~615 líneas
│       └── mysql_inventory_detail_repository.go # ~487 líneas
├── application/
│   └── service/
│       ├── inventory_service.go                 # Lógica de negocio completa (~412 líneas)
│       ├── inventory_service_integration_test.go # Tests de GetDiscrepancies + CompleteInventory
│       └── employee_service.go                  # Facade que delega a InventoryService
└── interface/
    └── handler/
        └── employee/
            └── inventory_handler.go             # 10 endpoints (~350 líneas)
```

### Puntos de Integración

| Componente | Ruta | Tipo de cambio |
|------------|------|----------------|
| Router | `internal/interface/router/router.go` | 10 rutas bajo `/api/inventories` |
| ItemRepository | `internal/domain/repository/item_repository.go` | `FindActiveByInventoryType` |
| InventoryEnricher | `internal/domain/inventory/enrich.go` | Calcula `TotalItems`, `ItemsWithDiff` |

## Fases de Implementación

### Fase 0: Verificación y preparación

- [x] Confirmar migraciones 001, 003, 009, 010 aplicadas
- [x] Confirmar que la lógica de dominio es pura (sin dependencias de frameworks)

### Fase 1: Dominio — Entidades y Repositorios

- [x] `entity.Inventory` con métodos `IsCompleted()`, `IsInitial()`, `RequiresSalesAndPurchases()`, `RequiresPurchasesOnly()`
- [x] `entity.InventoryDetail` con método `IsComplete()` (verifica `real_value != nil` + ventas si aplica)
- [x] Interfaces `InventoryRepository` e `InventoryDetailRepository` en `internal/domain/repository/`
- [x] Lógica pura en `internal/domain/inventory/`:
  - `expected.go`: `ExpectedAtEnd(d)` — `suggested + stock_received - units_sold` (con guards para nil)
  - `discrepancy.go`: `HasDiscrepancyFromExpectedEnd(d, expected)` — `*real_value != expected`
  - `enrich.go`: `Enricher` que agrega stats de discrepancias a listas de inventarios

### Fase 2: Infraestructura — MySQL

- [x] `mysql_inventory_repository.go`: `Create`, `FindByID`, `FindByDateTypeAndSchedule`, `FindPreviousInventory`, `FindLatestCompleted`, `FindInProgressByEmployee`, `Complete`
- [x] `mysql_inventory_detail_repository.go`: `CreateBatch`, `FindByInventoryID`, `FindByInventoryIDWithItems`, `FindByInventoryAndItem`, `Update`
- [x] Migraciones aplicadas:
  - `migrations/001_initial_schema.up.sql` — tablas base
  - `migrations/003_inventory_frequency.up.sql` — `inventory_type`, UK fecha/tipo/turno
  - `migrations/009_initial_inventory_type.up.sql` — valor `initial`
  - `migrations/010_inventory_details_shrinkage.up.sql` — columna `shrinkage`

### Fase 3: Servicio de Aplicación

- [x] `CreateInventory`: validación → check duplicado → create → `prepopulateInventoryDetails` (sugerido = `real_value` anterior del mismo ítem) → log INFO → metric OTel
- [x] `GetSuggestedSchedule`: función pura basada en hora del día (opening/noon/closing)
- [x] `SaveInventoryDetail`: verifica inventario no completado → actualiza `real_value`
- [x] `SaveSalesAndPurchases`: verifica tipo de inventario → actualiza `stock_received`/`units_sold` → recalcula `suggested_value`
- [x] `GetDiscrepancies`: filtra detalles donde `ExpectedAtEnd(d) != real_value`
- [x] `CompleteInventory`: verifica que todos los ítems estén completos → cuenta discrepancias → marca completed → log INFO → metric OTel

### Fase 4: Handler e Interface

- [x] 10 endpoints en `employee/inventory_handler.go` — todos usan `response.Respond*`, ninguno loguea

### Fase 5: Tests

- [x] `internal/domain/inventory/expected_test.go` — `ExpectedAtEnd` con varios escenarios
- [x] `internal/domain/inventory/discrepancy_test.go` — `HasDiscrepancyFromExpectedEnd`
- [x] `internal/domain/inventory/consistency_test.go` — cross-check de consistencia
- [x] `internal/application/service/inventory_service_integration_test.go` — `GetDiscrepancies` y `CompleteInventory` con fakes

⚠️ **Sin tests para**: `CreateInventory`, `SaveInventoryDetail`, `SaveSalesAndPurchases`, `GetInventoryItems`

### Fase 6: Observabilidad OTel

- [x] Metric `loopi.inventory.movements` — `metric.Int64Counter` con atributos `action` (create/complete) y `type` (daily/weekly/monthly/initial)
- ⚠️ Sin OTel spans para operaciones individuales (gap)

## Complexity Tracking

| Excepción | Por qué necesaria | Alternativa descartada porque |
|-----------|-------------------|-------------------------------|
| Inventario `initial` no genera discrepancias | Es el inventario base de referencia — no tiene "anterior" contra el que comparar | Generar discrepancias falsas en el primer inventario confundiría a los usuarios |
| `RequiresPurchasesOnly` fuerza `units_sold=0` en weekly/monthly | El POS ya registra ventas — el inventario no debe duplicar ese dato | Permitir `units_sold` en semanales generaría doble conteo |
