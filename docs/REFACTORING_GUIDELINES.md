# Lineamientos de refactorización: Admin y optimización

Este documento define la estrategia de atomización de `AdminService`, optimizaciones de concurrencia y patrones de diseño aplicables al proyecto, con un enfoque en **lineamientos generales** para futuras features, no solo soluciones puntuales.

---

## 1. Atomización por dominio

### 1.1 Problema actual

`admin_service.go` (~1296 líneas) concentra todos los puntos de entrada del administrador en un único tipo, mezclando:

- **Dashboard** (agregador)
- **Inventario** (vista admin: listado, detalle, actualización de detalles, inventario inicial)
- **Ítems** (CRUD + agregar a inventarios activos)
- **Empleados** (CRUD + reset password, listado activos)
- **Categorías** (CRUD + reordenamiento)
- **Proveedores** (CRUD + listado activos)

Esto viola el principio de responsabilidad única y dificulta el mantenimiento y las pruebas.

### 1.2 Dominios ya definidos en el proyecto

En `internal/domain` existen:

- `entity` (inventory, item, category, supplier, employee)
- `inventory` (expected, discrepancy, enricher)
- `employee` (password)
- `repository` (interfaces por agregado)

La atomización debe alinearse con estos agregados y con la capa de aplicación (services/use cases).

### 1.3 Propuesta de servicios por dominio

| Servicio (use case) | Responsabilidad | Métodos actuales | Repos de los que depende |
|---------------------|-----------------|------------------|---------------------------|
| **AdminDashboardService** | Estadísticas y discrepancias recientes | `GetDashboard` | inventory, inventoryDetail + enricher |
| **AdminInventoryService** | Listado, detalle, edición de detalles, inventario inicial | `ListInventories`, `GetInventoryDetail`, `UpdateInventoryDetail`, `GetActiveInventoriesCount`, `CreateInitialInventory` + `countItemsWithDiscrepancy`, `getInventoryCounts` | inventory, inventoryDetail, employee, item + enricher |
| **AdminItemService** | CRUD ítems y agregar a inventarios activos | `ListItems`, `GetItem`, `CreateItem`, `UpdateItem`, `UpdateItemStatus`, `addItemToActiveInventories`, `shouldIncludeItem` | item, inventory, inventoryDetail, category, supplier |
| **AdminEmployeeService** | CRUD empleados, reset password, listado activos | `ListEmployees`, `GetEmployee`, `CreateEmployee`, `UpdateEmployee`, `UpdateEmployeeStatus`, `ResetEmployeePassword`, `ListAllActiveEmployees` + `hashPassword` | employee |
| **AdminCategoryService** | CRUD categorías y reordenamiento | `ListCategories`, `GetCategory`, `CreateCategory`, `UpdateCategory`, `UpdateCategoryStatus`, `ReorderCategories` | category |
| **AdminSupplierService** | CRUD proveedores y listado activos | `ListSuppliers`, `GetSupplier`, `CreateSupplier`, `UpdateSupplier`, `UpdateSupplierStatus`, `ListAllActiveSuppliers` | supplier |

### 1.4 Rol de AdminService: fachada (orchestrator)

- **AdminService** debe quedar como **fachada de solo delegación**: sin lógica de negocio, solo compone los servicios de dominio y expone la misma API que hoy usa `AdminHandler`.
- El handler sigue inyectando un único `AdminService`; internamente este tiene los 6 servicios (o interfaces) y cada método delega al servicio correspondiente.
- Ventajas: el handler no cambia de firma; la migración puede ser incremental (mover métodos de a grupos); cada servicio puede testearse y evolucionar por separado.

### 1.5 Estructura de paquetes sugerida

```
internal/application/
  service/
    admin/
      admin.go              # AdminService struct + NewAdminService (inyección de los 6)
      dashboard.go          # AdminDashboardService
      inventory.go          # AdminInventoryService
      item.go               # AdminItemService
      employee.go           # AdminEmployeeService
      category.go           # AdminCategoryService
      supplier.go           # AdminSupplierService
      types.go              # DTOs/filtros compartidos (DashboardData, InventoryFilter, etc.)
```

Alternativa más plana (si se prefiere menos carpetas):

- Mantener `service/admin_service.go` como fachada delgado (~100–150 líneas) que solo delega.
- Crear `service/admin_dashboard.go`, `service/admin_inventory.go`, etc., en el mismo paquete `service`, con tipos `AdminDashboardService`, `AdminInventoryService`, etc. AdminService los embebe o tiene como campos.

### 1.6 Interfaces por servicio (testabilidad)

Cada subservicio debe exponer una **interfaz** en el mismo paquete (o en un paquete `port` si se quiere hexagonal estricto), para que la fachada y los tests dependan de abstracciones:

```go
// Ejemplo: internal/application/service/admin/dashboard.go
type DashboardService interface {
    GetDashboard(ctx context.Context, days int) (*DashboardData, error)
}

type AdminDashboardService struct { ... }
func (s *AdminDashboardService) GetDashboard(...) { ... }
```

AdminService recibiría `DashboardService`, `InventoryAdminService`, etc., en lugar de structs concretos.

### 1.7 Regla para nuevas features de admin

- **Criterio**: si la feature pertenece claramente a un agregado (inventory, item, employee, category, supplier), su lógica vive en el servicio de ese dominio (AdminXxxService), no en un “AdminService” monolítico.
- **Dashboard / reportes**: si agrega datos de varios dominios, crear o extender un use case de “dashboard” o “reportes” que orqueste lecturas sobre los servicios de dominio, sin duplicar lógica de negocio.

---

## 2. Optimización con goroutines y patrones

### 2.1 Principio general

- **Llamadas independientes**: si dos o más llamadas a repos (o a servicios) no dependen entre sí, ejecutarlas en paralelo con `errgroup` (o goroutines + canal de errores).
- **Bucles que hacen I/O**: si un loop hace una llamada por iteración (p. ej. N consultas), valorar: (1) batch/preload en el repo, o (2) paralelismo acotado (worker pool / semáforo) para no disparar demasiadas goroutines ni saturar DB.
- **Context**: propagar siempre `context.Context` y respetar cancelación en todas las llamadas paralelas.

### 2.2 GetDashboard (línea ~79)

**Situación actual:**

- Cuatro llamadas secuenciales: `CountInventoriesByDate`, `CountInProgress`, `FindCompletedInventoriesByDate`, `FindRecentDetailsWithInventory`.
- Las **cuatro son independientes** entre sí.
- Luego: loop sobre `completedToday` llamando `countItemsWithDiscrepancy(inv)` por inventario (N consultas secuenciales).
- Luego: agrupación por inventario y enriquecimiento + filtro de discrepancias (CPU + posible I/O según Enricher).

**Propuesta:**

1. **Fase 1 – Paralelizar las 4 lecturas iniciales** con `golang.org/x/sync/errgroup`:

   ```go
   var todayCount int
   var pending int
   var completedToday []*entity.Inventory
   var recentDetails []*entity.InventoryDetail
   g, ctx := errgroup.WithContext(ctx)
   g.Go(func() error { var err error; todayCount, err = s.inventoryRepo.CountInventoriesByDate(ctx, today); return err })
   g.Go(func() error { var err error; pending, err = s.inventoryRepo.CountInProgress(ctx); return err })
   g.Go(func() error { var err error; completedToday, err = s.inventoryRepo.FindCompletedInventoriesByDate(ctx, today); return err })
   g.Go(func() error { var err error; recentDetails, err = s.inventoryDetailRepo.FindRecentDetailsWithInventory(ctx, days, 100); return err })
   if err := g.Wait(); err != nil { return nil, err }
   ```

2. **Fase 2 – Conteo withDiscrepancies**: el loop `for _, inv := range completedToday { countItemsWithDiscrepancy(ctx, inv) }` hace 1 query por inventario. Opciones:
   - **A)** Introducir en el repositorio un método batch, p. ej. `CountItemsWithDiscrepancyByInventoryIDs(ctx, ids []uint32) (map[uint32]int, error)`, que en una o pocas queries devuelva los conteos para todos los IDs (evita N round-trips).
   - **B)** Si se mantiene el loop: paralelizar con errgroup y un **máximo de concurrencia** (semáforo o pool de workers) para no abrir demasiadas conexiones (p. ej. límite 5–10). Misma idea que en 2.4.

3. **Fase 3 – Procesamiento de `invDetails`**: el doble loop (por inventario, luego por detalle) con `Enrich` y reglas de dominio puede mantenerse secuencial si Enricher es rápido; si Enricher hace I/O o el volumen crece, aplicar un **worker pool** que procese grupos de inventarios en paralelo (con límite de goroutines) y agregue resultados en un slice con mutex o canal.

### 2.3 GetInventoryDetail (línea ~325)

**Situación actual:**

- `FindByIDWithEmployee(inventoryID)` y `FindByInventoryIDWithItems(inventoryID)` son **independientes** (mismo ID pero dos consultas distintas).

**Propuesta:**

- Ejecutar ambas en paralelo con `errgroup`; si alguna falla, devolver error; si ambas ok, seguir con `Enrich` y construcción de `InventoryDetailView`.

```go
var inventory *entity.Inventory
var details []*entity.InventoryDetail
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { var err error; inventory, err = s.inventoryRepo.FindByIDWithEmployee(ctx, inventoryID); return err })
g.Go(func() error { var err error; details, err = s.inventoryDetailRepo.FindByInventoryIDWithItems(ctx, inventoryID); return err })
if err := g.Wait(); err != nil { return nil, ... }
if inventory == nil { return nil, nil }
// Enrich + build view...
```

### 2.4 UpdateInventoryDetail (línea ~385)

**Situación actual:**

- `FindByID(inventoryID)` y `FindByID(detailID)`; la validación de “detail pertenece al inventory” se hace después.

**Propuesta:**

- Ambas lecturas son independientes; ejecutarlas en paralelo con `errgroup`. Luego en código secuencial validar `detail.InventoryID == inventoryID` y regla de shrinkage. No hay más I/O en el flujo, por lo que con paralelizar las dos lecturas es suficiente.

### 2.5 ListInventories: ciclo con getInventoryCounts

**Situación actual:**

- Se obtienen inventarios paginados y luego, por cada uno, `getInventoryCounts(ctx, inv)` → N llamadas a `FindByInventoryID` + Enrich.

**Propuesta (lineamiento):**

- **Patrón “batch load”**: evitar 1 query por ítem de la lista. Opciones:
  - Nuevo método en repositorio, p. ej. `FindDetailsGroupedByInventoryIDs(ctx, ids []uint32) (map[uint32][]*entity.InventoryDetail, error)`, y en el servicio un solo Enrich/count por inventario sobre ese mapa. Así se pasa de N queries a 1 (o 2 si se separa por límites de SQL).
- Si no se puede batch aún: paralelizar el loop con **semáforo** (p. ej. `sem := make(chan struct{}, 10)` y goroutines que adquieran/liberen) y `errgroup`, y construir la lista de ítems de forma thread-safe (mutex o canal de resultados). Límite bajo (5–10) para no saturar DB.

### 2.6 Regla para ciclos que hacen I/O

- **Preferencia 1**: Diseñar en el repositorio una operación **batch** que devuelva en una o pocas queries los datos necesarios para todo el conjunto (por IDs, por rango, etc.).
- **Preferencia 2**: Si el loop es inevitable, usar **paralelismo acotado**: worker pool o semáforo, con `errgroup` para errores y `context` para cancelación. Evitar “una goroutine por ítem” sin límite.
- Documentar en ARCHITECTURE o en este doc que “los servicios no deben hacer N queries en loop sin batch o sin límite de concurrencia”.

---

## 3. Resumen de lineamientos para el proyecto

### 3.1 Servicios y dominios

- **Un servicio (use case) por agregado/dominio** en la capa de aplicación; fachadas (p. ej. Admin) solo orquestan y delegan.
- **Interfaces por servicio** para inyección y tests.
- **Nuevas features de admin**: asignar a un AdminXxxService existente o crear uno nuevo si aparece un nuevo agregado; no añadir lógica al “AdminService” monolítico.

### 3.2 Concurrencia

- **Llamadas independientes**: paralelizar con `errgroup` (o equivalente) y propagar `context`.
- **Bucles con I/O**: preferir batch en repo; si no, paralelismo acotado (semáforo/worker pool) y manejo seguro de errores y cancelación.
- **No** lanzar un número unbounded de goroutines por ítem sin límite explícito.

### 3.3 Repositorios

- Si un servicio repite un patrón “for each id, get X”, considerar un método **FindByIDs** o **batch** en el repositorio correspondiente y usarlo desde el servicio.

### 3.4 Orden sugerido de refactor

1. **Fase 1 – Atomización**: Extraer los 6 servicios (dashboard, inventory, item, employee, category, supplier) y tipos asociados; convertir AdminService en fachada que los inyecta y delega. Sin cambiar comportamiento ni optimizar aún.
2. **Fase 2 – Paralelismo “fácil”**: Aplicar errgroup en GetDashboard (4 lecturas), GetInventoryDetail (2 lecturas), UpdateInventoryDetail (2 lecturas).
3. **Fase 3 – Batch y loops**: Introducir métodos batch donde hoy hay N queries en loop (dashboard counts por inventarios, listado de inventarios con conteos); opcionalmente worker pool en procesamiento de discrepancias recientes si hace falta.
4. **Fase 4 – Handler**: Si se desea, el handler podría inyectar servicios individuales para rutas concretas; no es obligatorio si la fachada AdminService se mantiene estable.

---

## 4. Referencias rápidas

| Archivo / método | Acción principal |
|------------------|------------------|
| `GetDashboard` | Paralelizar 4 lecturas; batch o pool para counts y para invDetails |
| `GetInventoryDetail` | Paralelizar FindByIDWithEmployee + FindByInventoryIDWithItems |
| `UpdateInventoryDetail` | Paralelizar FindByID(inv) + FindByID(detail) |
| `ListInventories` | Batch load de details por inventory IDs en lugar de getInventoryCounts por inv |
| `admin_service.go` | Dividir en AdminDashboard, AdminInventory, AdminItem, AdminEmployee, AdminCategory, AdminSupplier; AdminService = fachada |

Este documento debe tratarse como la referencia para refactorizar `admin_service.go` y para aplicar los mismos criterios en futuras features del proyecto.
