# Plan de Implementación: Panel de Administración

**Branch**: `feature/admin` (migrado) | **Fecha**: 2026-05-14 | **Spec**: `specs/005-admin/spec.md`

## Resumen

Panel de administración completo que cubre 6 sub-dominios: dashboard (stats del día), gestión de inventarios (lista/detalle/edición/inventario inicial), ítems (CRUD + activar en inventarios activos), empleados (CRUD + reset password), categorías (CRUD + reorden) y proveedores (CRUD). Implementado con el patrón Facade (`AdminService`) que delega a servicios especializados por sub-dominio.

## Contexto Técnico (loopi-api — Bloqueado por Constitution)

| Categoría | Tecnología | Versión |
|-----------|------------|---------|
| Lenguaje | Go | 1.24+ |
| HTTP Router | go-chi/chi | v5.2.4 |
| Base de datos | MySQL 8.0 via go-sql-driver | v1.9.3 |
| Hashing | golang.org/x/crypto/bcrypt | cost=12 |
| Observabilidad | OpenTelemetry SDK | v1.43.0 |
| Logging | log/slog (stdlib) | stdlib |
| Testing | testing (stdlib) + fakes | stdlib |

## Verificación de Cumplimiento Constitution

- [x] **Clean Architecture**: 6 handlers en `internal/interface/handler/admin/`, 7 servicios en `internal/application/service/admin_*.go`
- [x] **Repository Contracts**: Todas las interfaces en `internal/domain/repository/` antes de implementar
- [x] **slog Only**: Logs en servicios únicamente
- [x] **Error Handling**: `apperrors.*` para errores de dominio, `response.Respond*` en handlers
- [x] **AdminOnly Middleware**: Todos los endpoints protegidos con `middleware.AdminOnly`
- ⚠️ **Sin tests**: Ningún admin service tiene tests — gap crítico

## Estructura del Proyecto

### Documentación

```text
specs/005-admin/
├── spec.md      # especificación funcional
├── plan.md      # este archivo
└── tasks.md     # tareas de implementación
```

### Archivos Implementados

```text
internal/
├── domain/
│   ├── entity/
│   │   ├── category.go
│   │   ├── supplier.go
│   │   └── measurement_unit.go
│   └── repository/
│       ├── category_repository.go
│       ├── supplier_repository.go
│       └── measurement_unit_repository.go
├── infrastructure/
│   └── repository/
│       ├── mysql_category_repository.go        # ~218 líneas
│       ├── mysql_supplier_repository.go        # ~246 líneas
│       ├── mysql_measurement_unit_repository.go # ~66 líneas
│       ├── mysql_employee_repository.go        # ~273 líneas
│       └── mysql_item_repository.go            # ~364 líneas
├── application/
│   └── service/
│       ├── admin_service.go                    # Facade — delega a 6 sub-servicios
│       ├── admin_types.go                      # DTOs compartidos (requests/responses)
│       ├── admin_dashboard.go                  # AdminDashboardService (~100 líneas)
│       ├── admin_inventory.go                  # AdminInventoryService (~334 líneas)
│       ├── admin_item.go                       # AdminItemService (~262 líneas)
│       ├── admin_employee.go                   # AdminEmployeeService (~224 líneas)
│       ├── admin_category.go                   # AdminCategoryService (~139 líneas)
│       └── admin_supplier.go                   # AdminSupplierService (~157 líneas)
└── interface/
    └── handler/
        └── admin/
            ├── dashboard_handler.go
            ├── inventory_handler.go
            ├── item_handler.go
            ├── employee_handler.go
            ├── category_handler.go
            └── supplier_handler.go
```

### Puntos de Integración

| Componente | Ruta | Tipo de cambio |
|------------|------|----------------|
| Router | `internal/interface/router/router.go` | 30 rutas bajo `/api/admin` + `middleware.AdminOnly` |
| Enricher | `internal/domain/inventory/enrich.go` | Reutilizado para stats de discrepancias |
| HashPassword | `internal/infrastructure/auth/password.go` | Reutilizado en `CreateEmployee` y `ResetEmployeePassword` |

## Fases de Implementación

### Fase 0: Verificación

- [x] Confirmar migraciones 001, 005, 006, 007, 008, 012 aplicadas
- [x] Confirmar `middleware.AdminOnly` implementado antes de registrar rutas

### Fase 1: Dominio — Nuevas Entidades y Repositorios

- [x] Entidades: `Category` (con `display_order`), `Supplier`, `MeasurementUnit` en `internal/domain/entity/`
- [x] Interfaces en `internal/domain/repository/`:
  - `CategoryRepository`: `FindAll`, `FindByID`, `Create`, `Update`, `UpdateStatus`, `UpdateDisplayOrders`
  - `SupplierRepository`: `FindAllWithFilters`, `FindAllActive`, `FindByID`, `Create`, `Update`, `UpdateStatus`
  - `MeasurementUnitRepository`: `FindAll`

### Fase 2: Infraestructura — Nuevos Repositorios MySQL

- [x] `mysql_category_repository.go` — incluye `UpdateDisplayOrders` (batch update)
- [x] `mysql_supplier_repository.go`
- [x] `mysql_measurement_unit_repository.go`
- [x] Extensión de `mysql_employee_repository.go` — `Create`, `Update`, `UpdateStatus`, `ResetPassword`
- [x] Extensión de `mysql_item_repository.go` — `FindAllWithFilters`, `Create`, `Update`, `UpdateStatus`
- [x] Migraciones:
  - `migrations/005_categories.up.sql` — tabla `categories`
  - `migrations/006_suppliers.up.sql` — tabla `suppliers`
  - `migrations/007_items_add_category_supplier_cost.up.sql` — relaciones en `items`
  - `migrations/008_document_type_nuip.up.sql` — campos de documento en `employees`
  - `migrations/012_measurement_units.up.sql` — tabla `measurement_units`

### Fase 3: AdminDashboardService

- [x] `GetDashboard(ctx, days)`: cuenta inventarios del día actual, separa con/sin discrepancias, cuenta pendientes (`in_progress`)
- [x] `admin_dashboard_handler.GetDashboard` → `GET /api/admin/dashboard`

### Fase 4: AdminInventoryService

- [x] `ListInventories`: paginación + filtros (fecha, tipo, empleado, hasDiscrepancies) usando `FindAllWithFilters`
- [x] `GetInventoryDetail`: detalle completo con todos los `InventoryDetail` + datos del empleado responsable
- [x] `UpdateInventoryDetail`: permite editar `suggested_value`, `real_value`, `stock_received`, `units_sold`, `shrinkage`
- [x] `CreateInitialInventory`: crea inventario tipo `initial` para el día actual
- [x] `GetActiveInventoriesCount`: usa `CountInProgress`
- [x] Handlers correspondientes en `admin/inventory_handler.go`

### Fase 5: AdminItemService

- [x] `ListItems`: paginación + filtros (`type`, `frequency`, `active`, `search`)
- [x] `GetItem`, `CreateItem`, `UpdateItem`, `UpdateItemStatus`
- [x] `CreateItem` con `add_to_active_inventories=true`: busca inventarios `in_progress` → agrega ítem con `CreateBatch`
- [x] `ListMeasurementUnits`: lista todas las unidades activas
- [x] Handlers en `admin/item_handler.go`

### Fase 6: AdminEmployeeService, AdminCategoryService, AdminSupplierService

- [x] `AdminEmployeeService`: CRUD + `ResetEmployeePassword` (genera contraseña temporal → bcrypt → guarda)
- [x] `AdminCategoryService`: CRUD + `ReorderCategories` (batch update de `display_order`)
- [x] `AdminSupplierService`: CRUD + `ListAllActiveSuppliers`
- [x] Handlers correspondientes (employee, category, supplier)
- [x] 30 rutas registradas en `router.go` bajo `/api/admin` con `middleware.AdminOnly`

### Fase 7: Tests

⚠️ **Sin tests para ningún admin service** — todos los gaps están en esta fase.

- [ ] Tests para `AdminDashboardService.GetDashboard`
- [ ] Tests para `AdminInventoryService.ListInventories`, `GetInventoryDetail`
- [ ] Tests para `AdminItemService.CreateItem` (especialmente con `add_to_active_inventories=true`)
- [ ] Tests para `AdminEmployeeService.CreateEmployee`, `ResetEmployeePassword`
- [ ] Tests para `AdminCategoryService.ReorderCategories`

### Fase 8: Observabilidad OTel

⚠️ **Sin OTel spans ni métricas en admin services** — gap identificado.

- [ ] Agregar span `tracer.Start(ctx, "admin.<domain>.<Method>")` en operaciones críticas
- [ ] Registrar metrics para operaciones de escritura (create/update/delete)

## Complexity Tracking

| Excepción | Por qué necesaria | Alternativa descartada porque |
|-----------|-------------------|-------------------------------|
| Facade pattern (`AdminService`) | Centraliza la inyección de dependencias en el router — evita 7 parámetros separados en `router.New` | Constructor del router con 7 servicios separados aumenta acoplamiento y hace el código difícil de mantener |
| `UpdateInventoryDetail` acepta todos los campos como `*uint16` | Admin necesita poder corregir cualquier campo de un detalle, incluyendo ventas y mermas que el empleado no puede editar | Endpoints separados por campo serían excesivos (6+ endpoints para un solo detalle) |
