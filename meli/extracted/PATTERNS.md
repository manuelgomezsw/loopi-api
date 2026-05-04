# Patrones del Proyecto — Loopi

**Generado**: 2026-03-26
**Repo size**: Mediano (~15k LOC estimado)
**Max patrones**: 20

---

## Categoría: HTTP/API

### 1. Respuesta JSON Centralizada

**Category**: HTTP/API

**Evidence**: Usado en:
- `internal/interface/response/response.go`
- Todos los handlers (auth, employee, admin/*.go)

**Example**:
```go
// Respuesta exitosa
response.RespondJSON(w, http.StatusOK, data)

// Respuesta de error
response.RespondError(w, http.StatusBadRequest, "mensaje de error")
```

**When to use**: Para toda respuesta HTTP del API. Centraliza el formato JSON y el Content-Type.

---

### 2. Manejo de Errores con AppError

**Category**: HTTP/API / Error Handling

**Evidence**: Usado en:
- `pkg/errors/errors.go`
- `internal/application/service/inventory_service.go`
- `internal/application/service/admin_employee.go`
- `internal/infrastructure/repository/mysql_employee_repository.go`

**Example**:
```go
// Definición de errores tipados
var ErrNotFound = &AppError{Code: "NOT_FOUND", HTTPStatus: 404}
var ErrConflict = &AppError{Code: "CONFLICT", HTTPStatus: 409}

// Uso en servicio
if employee == nil {
    return nil, ErrNotFound
}
```

**When to use**: Para errores de negocio con código HTTP específico. Permite al handler mapear errores sin lógica condicional.

---

### 3. Paginación Consistente

**Category**: HTTP/API

**Evidence**: Usado en:
- `internal/application/service/admin_types.go`
- `internal/application/service/admin_inventory.go`
- `internal/application/service/admin_item.go`
- `internal/application/service/admin_employee.go`
- `internal/application/service/admin_supplier.go`

**Example**:
```go
// Filtro estándar de paginación
type InventoryFilter struct {
    Page     int
    PageSize int // default 20, max 100
    // ... campos específicos del recurso
}

// Resultado paginado
type InventoryListResult struct {
    Items      []*Inventory
    TotalCount int
    Page       int
    PageSize   int
}
```

**When to use**: Para cualquier endpoint de listado con soporte de filtros y paginación.

---

## Categoría: Error Handling

### 4. Context Propagation en Todos los Repositorios

**Category**: Error Handling / Database

**Evidence**: Usado en:
- Todas las interfaces en `internal/domain/repository/`
- Todas las implementaciones en `internal/infrastructure/repository/`

**Example**:
```go
// Interfaz del repositorio
type InventoryRepository interface {
    FindByID(ctx context.Context, id uint32) (*entity.Inventory, error)
    Create(ctx context.Context, inv *entity.Inventory) error
    // ...
}
```

**When to use**: Todos los métodos de repositorio y servicio deben aceptar `context.Context` como primer parámetro para permitir cancelación y deadlines.

---

### 5. Queries en Paralelo con errgroup

**Category**: Error Handling / Database

**Evidence**: Usado en:
- `internal/application/service/admin_dashboard.go`
- `internal/application/service/admin_inventory.go`

**Example**:
```go
import "golang.org/x/sync/errgroup"

eg, ctx := errgroup.WithContext(ctx)

var result1 int
var result2 []*Entity

eg.Go(func() error {
    var err error
    result1, err = repo.CountToday(ctx)
    return err
})
eg.Go(func() error {
    var err error
    result2, err = repo.FindDetails(ctx, id)
    return err
})

if err := eg.Wait(); err != nil {
    return nil, err
}
```

**When to use**: Cuando se necesitan múltiples queries independientes en un mismo handler/servicio. Reduce la latencia total al ejecutarlas en paralelo.

---

## Categoría: Database

### 6. Facade de Servicio Admin

**Category**: Database / Architecture

**Evidence**: Usado en:
- `internal/application/service/admin_service.go`
- `internal/application/service/admin_types.go`

**Example**:
```go
// AdminService como punto de entrada único para el admin
type AdminService struct {
    dashboard  *AdminDashboardService
    inventory  *AdminInventoryService
    item       *AdminItemService
    employee   *AdminEmployeeService
    category   *AdminCategoryService
    supplier   *AdminSupplierService
}

func (s *AdminService) Dashboard() *AdminDashboardService { return s.dashboard }
func (s *AdminService) Inventory() *AdminInventoryService { return s.inventory }
// ...
```

**When to use**: Para agrupar múltiples servicios relacionados con un único actor (admin) bajo un facade. Simplifica el wiring en `router.go`.

---

### 7. Soft Delete via Campo `active`

**Category**: Database

**Evidence**: Usado en:
- `internal/domain/entity/employee.go` (campo `active bool`)
- `internal/domain/entity/item.go` (campo `active bool`)
- `internal/domain/entity/category.go` (campo `active bool`)
- `internal/domain/entity/supplier.go` (campo `active bool`)

**Example**:
```go
// Entidad con soft-delete
type Item struct {
    ID     uint16 `json:"id"`
    Active bool   `json:"active"`
    // ...
}

// El repo filtra por active=true por defecto
// El endpoint PATCH /{id}/status cambia solo el campo active
```

**When to use**: Para todas las entidades de catálogo (items, empleados, categorías, proveedores). Nunca se eliminan registros físicamente, solo se desactivan.

---

### 8. Campos Calculados Fuera de la DB

**Category**: Database / Architecture

**Evidence**: Usado en:
- `internal/application/service/admin_inventory.go` (calcula `TotalItems`, `ItemsWithDiff` en memoria)
- `internal/domain/inventory/expected.go` (calcula `expected_at_end` en código)
- `internal/domain/entity/category.go` (`item_count` no está en DB)

**Example**:
```go
// El repo devuelve datos crudos
// El servicio calcula lo derivado
for _, detail := range details {
    if domain.HasDiscrepancyFromExpectedEnd(detail) {
        itemsWithDiff++
    }
}
inventory.ItemsWithDiff = itemsWithDiff
```

**When to use**: Para campos derivados o calculados. Nunca poner lógica de negocio en SQL. El SQL solo hace SELECT/INSERT/UPDATE/DELETE con datos crudos.

---

## Categoría: Architecture

### 9. Inyección de Dependencias Manual en Router

**Category**: Architecture

**Evidence**: Usado en:
- `internal/interface/router/router.go` (único lugar de wiring)

**Example**:
```go
// En router.go, toda la inyección de dependencias
db := database.NewMySQL(cfg.Database)
jwtMgr := auth.NewJWTManager(cfg.JWT)

employeeRepo := repository.NewMySQLEmployeeRepository(db)
inventoryRepo := repository.NewMySQLInventoryRepository(db)

authSvc := service.NewAuthService(employeeRepo, jwtMgr)
inventorySvc := service.NewInventoryService(inventoryRepo, detailRepo, itemRepo, enricher)
adminSvc := service.NewAdminService(...)

authHandler := authhandler.New(authSvc)
// ...
```

**When to use**: Para wiring de toda la aplicación. Sin frameworks DI. El router es el único lugar donde se instancian repositorios y servicios.

---

### 10. Interfaces de Repositorio en el Dominio

**Category**: Architecture

**Evidence**: Usado en:
- `internal/domain/repository/employee_repository.go`
- `internal/domain/repository/inventory_repository.go`
- `internal/domain/repository/item_repository.go`
- `internal/domain/repository/category_repository.go`
- `internal/domain/repository/supplier_repository.go`
- `internal/domain/repository/measurement_unit_repository.go`

**Example**:
```go
// Interfaz en el dominio (capa 3)
// internal/domain/repository/employee_repository.go
type EmployeeRepository interface {
    FindByID(ctx context.Context, id uint16) (*entity.Employee, error)
    FindByUsername(ctx context.Context, username string) (*entity.Employee, error)
    Create(ctx context.Context, e *entity.Employee) error
    Update(ctx context.Context, e *entity.Employee) error
    UpdateStatus(ctx context.Context, id uint16, active bool) error
    FindAll(ctx context.Context, filter EmployeeFilter) ([]*entity.Employee, int, error)
    FindAllActive(ctx context.Context) ([]*entity.Employee, error)
}

// Implementación en infraestructura (capa 4)
// internal/infrastructure/repository/mysql_employee_repository.go
type MySQLEmployeeRepository struct { db *sql.DB }
func (r *MySQLEmployeeRepository) FindByID(ctx context.Context, id uint16) (*entity.Employee, error) { ... }
```

**When to use**: Toda interacción con la DB debe ir detrás de una interfaz definida en el dominio. Permite mockear en tests y respetar la inversión de dependencias.

---

### 11. Enums como Tipos Alias de String

**Category**: Architecture

**Evidence**: Usado en:
- `internal/domain/entity/employee.go` (`type Role string`)
- `internal/domain/entity/item.go` (`type ItemType string`, `type InventoryFrequency string`)
- `internal/domain/entity/inventory.go` (`type InventoryType string`, `type Schedule string`, `type InventoryStatus string`)

**Example**:
```go
type Role string

const (
    RoleEmployee Role = "employee"
    RoleAdmin    Role = "admin"
)
```

**When to use**: Para campos con valores fijos (enums en DB). Usar tipo alias de string para type-safety en Go y compatibilidad directa con JSON y MySQL ENUM.

---

## Categoría: Security

### 12. JWT con Claims Personalizados

**Category**: Security

**Evidence**: Usado en:
- `internal/infrastructure/auth/jwt.go`
- `internal/interface/middleware/auth_middleware.go`

**Example**:
```go
// Claims personalizados
type Claims struct {
    jwt.RegisteredClaims
    EmployeeID uint16 `json:"employee_id"`
    Username   string `json:"username"`
    Role       Role   `json:"role"`
}

// Extraer del context en handlers
employeeID := ctx.Value("employee_id").(uint16)
role := ctx.Value("role").(string)
```

**When to use**: Para autenticación JWT. Los claims se inyectan en el context por el `AuthMiddleware` y están disponibles para todos los handlers protegidos.

---

### 13. Separación de Middleware por Rol

**Category**: Security

**Evidence**: Usado en:
- `internal/interface/middleware/auth_middleware.go` (`AuthMiddleware`, `AdminOnly`)
- `internal/interface/router/router.go`

**Example**:
```go
// Grupo protegido para cualquier empleado
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(jwtMgr))
    r.Get("/employees/me", authHandler.GetMe)
    // ...
})

// Grupo solo para admin
r.Group(func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(jwtMgr))
    r.Use(middleware.AdminOnly)
    r.Get("/admin/dashboard", adminDashboardHandler.GetDashboard)
    // ...
})
```

**When to use**: Para proteger rutas por rol. `AuthMiddleware` valida JWT para cualquier usuario; `AdminOnly` restringe adicionalmente al rol "admin".

---

## Categoría: Testing (Planificado)

### 14. Tests Unitarios para Dominio Puro

**Category**: Testing

**Evidence**: ⚠️ PLANIFICADO (no hay tests existentes documentados). Mencionado en:
- `docs/PLAN_AJUSTES_ESTRUCTURALES_SENIOR.md`
- `ARCHITECTURE.md`

**Example**:
```go
// Patrón propuesto para tests de dominio
func TestExpectedAtEnd(t *testing.T) {
    detail := &entity.InventoryDetail{
        SuggestedValue: ptr(uint16(10)),
        StockReceived:  ptr(uint16(5)),
        UnitsSold:      ptr(uint16(3)),
    }
    expected := domain.ExpectedAtEnd(detail)
    assert.Equal(t, uint16(12), expected) // 10 + 5 - 3
}
```

**When to use**: Las reglas de dominio puras (`ExpectedAtEnd`, `ExpectedForAdmin`, `HasDiscrepancy`, `Enricher`) deben ser testeadas con unit tests sin dependencias de DB o servicios.

---

## Categoría: Frontend

### 15. Angular Signals para Estado de Servicios

**Category**: Frontend

**Evidence**: Usado en:
- `src/app/core/services/auth.service.ts`
- `src/app/core/services/inventory.service.ts`
- `src/app/core/services/admin.service.ts`

**Example**:
```typescript
// Servicio con signals
@Injectable({ providedIn: 'root' })
export class AuthService {
  private _employee = signal<Employee | null>(null);
  readonly employee = this._employee.asReadonly();

  login(req: LoginRequest): Observable<LoginResponse> {
    return this.http.post<LoginResponse>('/api/auth/login', req).pipe(
      tap(res => this._employee.set(res.employee))
    );
  }
}
```

**When to use**: Para state management en Angular. Usar signals en lugar de BehaviorSubject. Los componentes consumen el signal directamente en el template.

---

### 16. Auth Interceptor Funcional

**Category**: Frontend / Security

**Evidence**: Usado en:
- `src/app/core/interceptors/auth.interceptor.ts`
- `src/app/app.config.ts`

**Example**:
```typescript
// Interceptor funcional (no clase)
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const token = localStorage.getItem('token');
  if (token) {
    req = req.clone({ setHeaders: { Authorization: `Bearer ${token}` } });
  }
  return next(req).pipe(
    catchError(err => {
      if (err.status === 401) { /* redirect to login */ }
      return throwError(() => err);
    })
  );
};
```

**When to use**: Para agregar el JWT a todos los requests HTTP. El interceptor funcional se configura en `app.config.ts` con `withInterceptors([authInterceptor])`.

---

### 17. Guards Funcionales para Rutas

**Category**: Frontend / Security

**Evidence**: Usado en:
- `src/app/core/guards/auth.guard.ts`
- Definidos en `app.routes.ts` y `admin.routes.ts`

**Example**:
```typescript
// Guard funcional (no clase)
export const authGuard: CanActivateFn = (route, state) => {
  const auth = inject(AuthService);
  const router = inject(Router);
  if (auth.employee()) return true;
  return router.createUrlTree(['/login']);
};

// Uso en rutas
{ path: 'inventory', component: HomeComponent, canActivate: [authGuard] }
```

**When to use**: Para proteger rutas por autenticación o rol. Usar guards funcionales (no clase CanActivate) en Angular 20.
