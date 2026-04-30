# Especificación Técnica — Loopi

**Versión**: 1.0.0 (extraída de código + documentación)
**Fecha**: 2026-03-26
**Idioma**: Español
**Módulo Go**: `github.com/manuelgomezsw/loopi-api`

> **Leyenda de confianza:**
> - ✅✅ VERIFIED — Confirmado en código + documentación
> - 🔸 CODE_ONLY — Solo en código (confiable)
> - ⚠️ DOCS_ONLY — Solo en docs (planificado, NO implementado)

---

## 1. Arquitectura del Sistema

### 1.1 Visión General ✅✅

```
┌─────────────────────────────────────────────────────────────────┐
│  loopi-web (Angular 20 + Firebase Hosting)                       │
│  https://loopi-c048d.web.app                                     │
└────────────────────────────┬────────────────────────────────────┘
                             │ HTTPS + JWT
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  loopi-api (Go 1.24 + Google App Engine)                         │
│  https://loopi-dot-quotes-api-100.ue.r.appspot.com               │
│                                                                  │
│  interface/ (HTTP handlers + router)                             │
│      ↓                                                           │
│  application/ (servicios + DTOs)                                 │
│      ↓                                                           │
│  domain/ (entidades + repositorios interfaces + reglas puras)    │
│      ↓                                                           │
│  infrastructure/ (repos MySQL + JWT + bcrypt)                    │
└────────────────────────────┬────────────────────────────────────┘
                             │ TCP + Unix Socket
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│  MySQL 8.0 (Google Cloud SQL)                                    │
│  IP: 34.23.218.229                                               │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Stack Tecnológico ✅✅

| Componente | Tecnología | Versión |
|-----------|-----------|---------|
| **Backend lenguaje** | Go | 1.24.0 |
| **Router HTTP** | go-chi/chi | v5.2.4 |
| **Base de datos** | MySQL | 8.0 |
| **Driver DB** | go-sql-driver/mysql | v1.9.3 |
| **Autenticación** | JWT (HS256) | golang-jwt/jwt v5.3.1 |
| **Password hashing** | bcrypt | golang.org/x/crypto |
| **Variables de entorno** | godotenv | v1.5.1 |
| **Concurrencia** | errgroup | golang.org/x/sync |
| **Logging** | log/slog (stdlib) | Go 1.21+ |
| **Frontend framework** | Angular | 20.1.0 |
| **Frontend estilos** | Tailwind CSS | 4.1.18 |
| **Frontend hosting** | Firebase Hosting | - |
| **Backend hosting** | Google App Engine Standard | - |

### 1.3 Arquitectura Backend: Clean Architecture ✅✅

```
internal/
├── interface/          # Capa 1: HTTP (handlers, middleware, router)
├── application/        # Capa 2: Casos de uso (services, DTOs)
├── domain/             # Capa 3: Negocio puro (entities, repos interfaces, reglas)
└── infrastructure/     # Capa 4: Implementaciones (MySQL repos, JWT, bcrypt)
```

**Principios**:
- Repositorios = solo datos (SELECT/INSERT/UPDATE/DELETE). CERO lógica de negocio.
- Toda lógica en servicios (orquestación) y dominio (reglas puras).
- Testabilidad: reglas puras con unit tests; servicios testeados con mocks de repos.

---

## 2. API Endpoints

### 2.1 Base URLs

| Ambiente | URL |
|----------|-----|
| Producción | `https://loopi-dot-quotes-api-100.ue.r.appspot.com` |
| Local | `http://localhost:8080` |

### 2.2 Autenticación ✅✅

- **Tipo**: Bearer JWT (HS256)
- **Header**: `Authorization: Bearer <token>`
- **Algoritmo**: HS256
- **Expiración**: configurable (default 24h via `JWT_EXPIRATION_HOURS`)
- **Claims**: `employee_id` (uint16), `username` (string), `role` (employee|admin)
- **Issuer**: `"loopi-api"`

### 2.3 Rutas Públicas ✅✅

| Método | Path | Descripción | Auth |
|--------|------|-------------|------|
| GET | `/health` | Health check. Respuesta: `{"status":"ok"}` | Ninguna |
| POST | `/api/auth/login` | Login. Body: `{username, password}`. Respuesta: `{token, employee}` | Ninguna |

### 2.4 Rutas del Empleado (JWT requerido) ✅✅

| Método | Path | Descripción |
|--------|------|-------------|
| GET | `/api/employees/me` | Perfil del empleado autenticado |
| GET | `/api/inventories/latest` | Último inventario completado del empleado |
| GET | `/api/inventories/in-progress` | Inventarios en progreso del empleado |
| GET | `/api/inventories/suggested-schedule` | Schedule sugerido según hora (Colombia) |
| POST | `/api/inventories` | Crear inventario. Body: `{inventory_type, schedule?, date}` |
| GET | `/api/inventories/{inventoryID}/items` | Items del inventario con valores sugeridos |
| POST | `/api/inventories/{inventoryID}/details` | Guardar conteo físico. Body: `{item_id, real_value}` |
| GET | `/api/inventories/{inventoryID}/discrepancies` | Items con discrepancia en el inventario |
| POST | `/api/inventories/{inventoryID}/sales` | Guardar ventas/compras. Body: `{item_id, stock_received?, units_sold?}` |
| GET | `/api/inventories/{inventoryID}/summary` | Resumen del inventario antes de completar |
| POST | `/api/inventories/{inventoryID}/complete` | Completar inventario |

### 2.5 Rutas Admin (JWT + rol "admin") ✅✅

| Método | Path | Query Params | Descripción |
|--------|------|-------------|-------------|
| GET | `/api/admin/dashboard` | `days` (default 3) | Estadísticas del dashboard |
| GET | `/api/admin/inventories` | `page`, `page_size`, `date_from`, `date_to`, `inventory_type`, `employee_id`, `has_discrepancies` | Lista paginada de inventarios |
| GET | `/api/admin/inventories/active-count` | - | Conteo de inventarios en progreso |
| POST | `/api/admin/inventories/initial` | - | Crear inventario inicial. Body: `{responsible_id}` |
| GET | `/api/admin/inventories/{inventoryID}` | - | Detalle completo de inventario |
| PUT | `/api/admin/inventories/{inventoryID}/details/{detailID}` | - | Editar detalle. Body: `{suggested_value?, real_value?, stock_received?, units_sold?, shrinkage?}` |
| GET | `/api/admin/measurement-units` | - | Listado de unidades de medida |
| GET | `/api/admin/items` | `page`, `page_size`, `type`, `frequency`, `active`, `search` | Lista paginada de items |
| POST | `/api/admin/items` | - | Crear item |
| GET | `/api/admin/items/{itemID}` | - | Detalle de item |
| PUT | `/api/admin/items/{itemID}` | - | Actualizar item |
| PATCH | `/api/admin/items/{itemID}/status` | - | Activar/desactivar item |
| GET | `/api/admin/employees` | `page`, `page_size`, `role`, `active`, `search` | Lista paginada de empleados |
| GET | `/api/admin/employees/active` | - | Todos los empleados activos (para dropdowns) |
| POST | `/api/admin/employees` | - | Crear empleado |
| GET | `/api/admin/employees/{employeeID}` | - | Detalle de empleado |
| PUT | `/api/admin/employees/{employeeID}` | - | Actualizar empleado |
| PATCH | `/api/admin/employees/{employeeID}/status` | - | Activar/desactivar empleado |
| POST | `/api/admin/employees/{employeeID}/reset-password` | - | Reset password a valor por defecto |
| GET | `/api/admin/categories` | - | Listado de categorías |
| POST | `/api/admin/categories` | - | Crear categoría |
| POST | `/api/admin/categories/reorder` | - | Reordenar categorías |
| GET | `/api/admin/categories/{categoryID}` | - | Detalle de categoría |
| PUT | `/api/admin/categories/{categoryID}` | - | Actualizar categoría |
| PATCH | `/api/admin/categories/{categoryID}/status` | - | Activar/desactivar categoría |
| GET | `/api/admin/suppliers` | `page`, `page_size`, `active`, `search` | Lista paginada de proveedores |
| GET | `/api/admin/suppliers/active` | - | Todos los proveedores activos |
| POST | `/api/admin/suppliers` | - | Crear proveedor |
| GET | `/api/admin/suppliers/{supplierID}` | - | Detalle de proveedor |
| PUT | `/api/admin/suppliers/{supplierID}` | - | Actualizar proveedor |
| PATCH | `/api/admin/suppliers/{supplierID}/status` | - | Activar/desactivar proveedor |

### 2.6 Middlewares ✅✅

| Middleware | Scope | Función |
|-----------|-------|---------|
| `middleware.RequestLogger` | Global | Log estructurado JSON de requests (reemplaza chi.Logger) — incluye request_id, method, path, status, latency_ms |
| `chi.Recoverer` | Global | Recupera panics |
| `chi.RequestID` | Global | Header X-Request-ID |
| `chi.RealIP` | Global | IP real del cliente |
| `cors.Handler` | Global | CORS (origins permitidos en sección 6) |
| `AuthMiddleware` | `/api/*` (excepto `/api/auth/*`) | Valida JWT, inyecta employee_id/role/username en context |
| `AdminOnly` | `/api/admin/*` | Verifica role == "admin", retorna 403 si no |

---

## 3. Modelos de Datos

### 3.1 Employee ✅✅

| Campo | Tipo | Nullable | Notas |
|-------|------|----------|-------|
| `id` | uint16 | No | PK AUTO_INCREMENT |
| `username` | string | No | UNIQUE, VARCHAR(50) |
| `password_hash` | string | No | bcrypt (cost 12), no expuesto en API |
| `name` | string | No | VARCHAR(50) |
| `last_name` | string | No | VARCHAR(50) |
| `document_type` | *string | Sí | VARCHAR(10) |
| `document_number` | *string | Sí | VARCHAR(20) |
| `phone` | *string | Sí | VARCHAR(20) |
| `email` | *string | Sí | VARCHAR(100) |
| `birth_date` | *time.Time | Sí | DATE |
| `role` | Role | No | ENUM: `employee`, `admin` |
| `active` | bool | No | DEFAULT true |
| `created_at` | time.Time | No | DATETIME |
| `updated_at` | time.Time | No | DATETIME |

**Método de dominio**: `FullName() string` → `Name + " " + LastName`

### 3.2 Item ✅✅

| Campo | Tipo | Nullable | Notas |
|-------|------|----------|-------|
| `id` | uint16 | No | PK AUTO_INCREMENT |
| `type` | ItemType | No | ENUM: `product`, `supply` |
| `name` | string | No | UNIQUE, VARCHAR(70) |
| `active` | bool | No | DEFAULT true |
| `inventory_frequency` | InventoryFrequency | No | ENUM: `daily`, `weekly`, `monthly` |
| `category_id` | uint16 | No | FK → categories.id |
| `supplier_id` | *uint16 | Sí | FK → suppliers.id |
| `cost` | uint32 | No | COP sin decimales, DEFAULT 0 |
| `measurement_unit_id` | uint16 | No | FK → measurement_units.id |
| `created_at` | time.Time | No | DATETIME |
| `updated_at` | time.Time | No | DATETIME |
| `category` | *Category | Sí | Relación (JOIN) |
| `supplier` | *Supplier | Sí | Relación (JOIN) |
| `measurement_unit` | *MeasurementUnit | Sí | Relación (JOIN) |

### 3.3 Inventory ✅✅

| Campo | Tipo | Nullable | Notas |
|-------|------|----------|-------|
| `id` | uint32 | No | PK AUTO_INCREMENT |
| `inventory_date` | time.Time | No | DATE |
| `inventory_type` | InventoryType | No | ENUM: `daily`, `weekly`, `monthly`, `initial` |
| `schedule` | *Schedule | Sí | ENUM: `opening`, `noon`, `closing` — solo para `daily` |
| `status` | InventoryStatus | No | ENUM: `in_progress`, `completed` |
| `responsible_id` | uint16 | No | FK → employees.id |
| `started_at` | time.Time | No | DATETIME |
| `completed_at` | *time.Time | Sí | DATETIME |
| `created_at` | time.Time | No | DATETIME |

**Constraint único**: `(inventory_date, inventory_type, schedule)`

**Métodos de dominio**:
- `IsCompleted() bool`
- `IsInitial() bool`
- `IsDaily() bool`
- `RequiresSalesAndPurchases() bool` → false para initial y daily-opening; true para el resto
- `RequiresPurchasesOnly() bool` → true para weekly y monthly

### 3.4 InventoryDetail ✅✅

| Campo | Tipo | Nullable | Notas |
|-------|------|----------|-------|
| `id` | uint32 | No | PK AUTO_INCREMENT |
| `inventory_id` | uint32 | No | FK → inventories.id ON DELETE CASCADE |
| `item_id` | uint16 | No | FK → items.id |
| `suggested_value` | *uint16 | Sí | Del período anterior (Enricher) |
| `real_value` | *uint16 | Sí | Conteo físico del empleado |
| `stock_received` | *uint16 | Sí | Compras recibidas |
| `units_sold` | *uint16 | Sí | Unidades vendidas |
| `shrinkage` | *uint16 | Sí | Mermas (solo admin) |
| `created_at` | time.Time | No | DATETIME |
| `updated_at` | time.Time | No | DATETIME |

**Constraint único**: `(inventory_id, item_id)`
**Método**: `IsComplete() bool` → `real_value != nil`

### 3.5 Category ✅✅

| Campo | Tipo | Notas |
|-------|------|-------|
| `id` | uint16 | PK |
| `name` | string | UNIQUE |
| `display_order` | int | DEFAULT 0 |
| `active` | bool | DEFAULT true |
| `created_at` | time.Time | |
| `updated_at` | time.Time | |
| `item_count` | int | Campo calculado (no en DB) |

### 3.6 Supplier ✅✅

| Campo | Tipo | Notas |
|-------|------|-------|
| `id` | uint16 | PK |
| `business_name` | string | VARCHAR(100) |
| `tax_id` | string | UNIQUE (NIT), VARCHAR(20) |
| `contact_name` | string | DEFAULT '' |
| `contact_phone` | string | DEFAULT '' |
| `contact_email` | string | DEFAULT '' |
| `active` | bool | DEFAULT true |
| `created_at` | time.Time | |
| `updated_at` | time.Time | |
| `item_count` | int | Campo calculado |

### 3.7 MeasurementUnit ✅✅

| Campo | Tipo | Notas |
|-------|------|-------|
| `id` | uint16 | PK |
| `code` | string | UNIQUE |
| `name` | string | |

**Valores semilla**: unit/Unidad, grams/Gramos, liters/Litros, meters/Metros, milliliters/Mililitros, ounces/Onzas

---

## 4. Lógica de Dominio

### 4.1 Fórmulas Centrales ✅✅

```go
// expected.go — internal/domain/inventory/

// ExpectedAtEnd: usado por empleado y en summary/discrepancies
// Fórmula: suggested + stock_received - units_sold (clamp >= 0, sin mermas)
func ExpectedAtEnd(d *InventoryDetail) uint16 {
    suggested := valueOf(d.SuggestedValue)
    received := valueOf(d.StockReceived)
    sold := valueOf(d.UnitsSold)
    if result := int(suggested) + int(received) - int(sold); result > 0 {
        return uint16(result)
    }
    return 0
}

// ExpectedForAdmin: usado por el administrador (incluye mermas)
// Fórmula: suggested - shrinkage + stock_received - units_sold (clamp >= 0)
func ExpectedForAdmin(d *InventoryDetail) uint16 {
    // misma lógica restando d.Shrinkage
}

// HasDiscrepancyFromExpectedEnd: discrepancia vs expected_at_end
func HasDiscrepancyFromExpectedEnd(d *InventoryDetail) bool {
    return d.RealValue != nil && *d.RealValue != ExpectedAtEnd(d)
}
```

### 4.2 Enricher (sugerido del período anterior) ✅✅

```
Para cada item en un inventario nuevo:
  suggested_value = real_value del inventario anterior del mismo (tipo, schedule, item_id)

Búsqueda del "anterior" para daily:
  closing   → busca noon (mismo día) → opening (mismo día) → initial
  noon      → busca opening (mismo día) → initial
  opening   → busca closing (día anterior) → initial

Las mermas NO restan del suggested del siguiente período.
```

### 4.3 Schedule Sugerido 🔸

```
Hora actual (America/Bogota):
  06:00–11:00 → schedule: opening
  11:00–16:00 → schedule: noon
  16:00–22:00 → schedule: closing
  Fuera de rango → sin sugerencia definida
```

### 4.4 Concurrencia ✅✅

```go
// Dashboard: 3 queries en paralelo con errgroup
eg, ctx := errgroup.WithContext(ctx)
eg.Go(func() error { /* count today */ })
eg.Go(func() error { /* count pending */ })
eg.Go(func() error { /* count completed */ })
eg.Wait()

// AdminInventoryService.GetInventoryDetail: 2 queries en paralelo
eg.Go(func() error { /* get inventory */ })
eg.Go(func() error { /* get details */ })
```

---

## 5. Infraestructura

### 5.1 Configuración ✅✅

**Variables de entorno** (`pkg/config/config.go`):

| Variable | Tipo | Default | Descripción |
|----------|------|---------|-------------|
| `SERVER_PORT` | string | `"8080"` | Puerto del servidor |
| `DB_HOST` | string | `"localhost"` | Host de MySQL |
| `DB_PORT` | string | `"3306"` | Puerto de MySQL |
| `DB_USER` | string | `"loopi"` | Usuario de MySQL |
| `DB_PASSWORD` | string | - | Contraseña de MySQL |
| `DB_NAME` | string | `"loopi"` | Nombre de la base de datos |
| `DB_INSTANCE_CONNECTION` | string | - | Unix socket para Cloud SQL (App Engine) |
| `JWT_SECRET` | string | - | Secreto para JWT |
| `JWT_EXPIRATION_HOURS` | int | 24 | Horas de validez del token |
| `TZ` | string | `"America/Bogota"` | Zona horaria de la aplicación |
| `LOG_LEVEL` | string | `"info"` | Nivel de log: debug, info, warn, error |
| `LOG_FORMAT` | string | `"text"` | Formato: text (dev) / json (producción + Cloud Logging) |

### 5.2 Conexión a Base de Datos ✅✅

```go
// Pool de conexiones
maxOpenConns = 25
maxIdleConns = 5
maxLifetime  = 5 * time.Minute

// Soporte Unix Socket para Cloud SQL en producción
// DSN: user:password@unix(/cloudsql/instance)?parseTime=true&loc=America/Bogota
```

### 5.3 CORS ✅✅

| Origin | Ambiente |
|--------|----------|
| `http://localhost:4200` | Desarrollo |
| `http://127.0.0.1:4200` | Desarrollo |
| `https://loopi-c048d.web.app` | Producción |
| `https://loopi-c048d.firebaseapp.com` | Producción |

**Métodos**: GET, POST, PUT, PATCH, DELETE, OPTIONS
**Headers**: Accept, Authorization, Content-Type, X-Request-ID
**MaxAge**: 300 segundos

### 5.4 Deploy Backend (Google App Engine) ✅✅

- **app.yaml**: runtime Go, servicio `loopi`
- **Proyecto GCP**: `quotes-api-100`
- **URL producción**: `https://loopi-dot-quotes-api-100.ue.r.appspot.com`
- **Región**: `ue` (us-east)

### 5.5 Deploy Frontend (Firebase) ✅✅

- **Proyecto Firebase**: `loopi-c048d`
- **URL producción**: `https://loopi-c048d.web.app`
- **SPA rewrite**: todas las rutas redirigen a `index.html`
- **PWA**: Service Worker habilitado

---

## 6. Arquitectura Frontend Angular

### 6.1 Estructura Principal ✅✅

```
src/app/
├── core/
│   ├── guards/         # authGuard, publicGuard, adminGuard, employeeGuard
│   ├── interceptors/   # auth.interceptor (Bearer token + manejo 401)
│   ├── models/         # employee.model, inventory.model, admin.model
│   └── services/       # auth.service, inventory.service, admin.service, storage.service
├── features/
│   ├── auth/login/     # Página de login
│   ├── admin/          # Layout admin + páginas (dashboard, inventories, items, employees, categories, suppliers)
│   ├── inventory/      # Flujo empleado (home, schedule-select, item-entry, discrepancy-review, sales-entry, summary, confirmation)
│   └── shared/         # role-redirect.component
└── shared/
    └── components/     # header, loading, numeric-input, progress-bar
```

### 6.2 State Management ✅✅

- **Angular Signals** en todos los servicios (no NgRx/Store)
- **Lazy loading** de todas las rutas
- **Standalone components** (Angular 20, sin módulos NgModule)
- **Auth interceptor funcional** (patrón funcional, no clase)

### 6.3 Flujo de Rutas ✅✅

```
/ → RoleRedirectComponent → /admin (si admin) | /inventory (si employee)

/login                     → LoginComponent (publicGuard)
/inventory                 → HomeComponent (authGuard)
/inventory/schedule        → ScheduleSelectComponent
/inventory/:id/item        → ItemEntryComponent
/inventory/:id/review      → DiscrepancyReviewComponent
/inventory/:id/sales       → SalesEntryComponent
/inventory/:id/summary     → SummaryComponent
/inventory/:id/confirmation → ConfirmationComponent

/admin                     → AdminLayoutComponent (authGuard + adminGuard)
/admin/dashboard           → DashboardComponent
/admin/inventories         → InventoryListComponent
/admin/inventories/:id     → InventoryDetailComponent
/admin/items               → ItemListComponent
/admin/employees           → EmployeeListComponent
/admin/categories          → CategoryListComponent
/admin/suppliers           → SupplierListComponent
```

---

## 7. Mapa de Propiedad del Código

| Componente | Archivos Primarios (1.0) | Archivos de Soporte (0.8) | Archivos Compartidos (0.4) |
|-----------|------------------------|--------------------------|--------------------------|
| Auth | `internal/interface/handler/auth/auth_handler.go`, `internal/infrastructure/auth/jwt.go` | `internal/application/service/auth_service.go` | `pkg/config/config.go` |
| InventoryService (core) | `internal/application/service/inventory_service.go` | `internal/domain/inventory/expected.go`, `internal/domain/inventory/enrich.go` | `internal/infrastructure/repository/mysql_inventory_repository.go` |
| AdminInventoryService | `internal/application/service/admin_inventory.go` | `internal/interface/handler/admin/inventory_handler.go` | `internal/domain/inventory/expected.go` |
| Employee CRUD | `internal/application/service/admin_employee.go` | `internal/interface/handler/admin/employee_handler.go` | `internal/infrastructure/repository/mysql_employee_repository.go` |
| Domain Entities | `internal/domain/entity/*.go` | - | Usadas por toda la app |
| Domain Rules | `internal/domain/inventory/expected.go`, `internal/domain/inventory/enrich.go`, `internal/domain/inventory/discrepancy.go` | - | Usadas por InventoryService y AdminService |
| Router/Wiring | `internal/interface/router/router.go` | - | Punto de entrada de toda la configuración |
| Config | `pkg/config/config.go` | `.env`, `app.yaml` | Usado por toda la app |
| DB Schema | `migrations/*.sql` | - | Base del sistema |

---

## 8. Deuda Técnica Identificada

| ID | Prioridad | Descripción | Fuente |
|----|-----------|-------------|--------|
| DT-001 | 🔴 Alta | Ausencia de tests automatizados (unitarios y de integración) | docs/PLAN_AJUSTES_ESTRUCTURALES_SENIOR.md |
| DT-002 | 🟡 Media | `GetRecentDiscrepancies` y `GetDashboardStats` potencialmente usan definición antigua | docs/AUTOCRITICA_DISCREPANCIAS.md |
| DT-003 | 🟡 Media | `FindDiscrepancies` en repo es código muerto con definición incorrecta | docs/AUTOCRITICA_DISCREPANCIAS.md |
| DT-004 | 🟡 Media | Métodos obsoletos `HasDiscrepancy()` y `Difference()` en entidad (si aún existen) | docs/PLAN_AJUSTES_ESTRUCTURALES_SENIOR.md |
| DT-005 | 🟢 Baja | README.md desactualizado (documenta 10 de 37 endpoints) | Análisis de código |
| DT-006 | 🟢 Baja | Lógica del Enricher potencialmente duplicada en InventoryService y AdminService | docs/AUTOCRITICA_DISCREPANCIAS.md |

---

## 9. Observabilidad y Logging ✅✅

### 9.1 Librería y Estrategia

- **Librería**: `log/slog` (stdlib Go 1.21+) — sin dependencias externas
- **Destino**: stdout → Google Cloud Logging (parsea JSON automáticamente en GAE)
- **Formato producción**: JSON con campo `severity` (requerido por Cloud Logging)
- **Formato desarrollo**: texto legible (controlado por `LOG_FORMAT=text`)
- **Paquete**: `pkg/logger` — factory `New(level, format)`, `WithContext`, `FromContext`

### 9.2 Niveles de Log

| Nivel | Cuándo |
|-------|--------|
| DEBUG | Queries SQL, estado interno (solo dev, `LOG_LEVEL=debug`) |
| INFO | Eventos de negocio exitosos: login, inventario creado/completado, empleado creado |
| WARN | Anomalías recuperables: credenciales inválidas, item no añadido a inventarios activos |
| ERROR | Fallos que requieren atención: errores de DB, errores de infraestructura |

### 9.3 Campos Estándar por Capa

**Todos los logs incluyen**: `severity`, `time`, `message`

**HTTP layer** (middleware `RequestLogger`):
```
request_id, method, path, status, latency_ms, remote_ip, user_agent
```

**Service layer**:
```
request_id (propagado por contexto), operation ("domain.Method"), error (solo WARN/ERROR)
```

### 9.4 Propagación por Contexto (Context Pattern)

```go
// El middleware inyecta un logger con request_id en el context de cada request
reqLog := log.With("request_id", reqID)
ctx := pkglogger.WithContext(r.Context(), reqLog)

// Los servicios consumen sin cambio de constructor ni firma de métodos
logger.FromContext(ctx).InfoContext(ctx, "inventory completed",
    "operation", "inventory.Complete",
    "inventory_id", inventoryID,
    "discrepancies", discrepancyCount,
)
```

### 9.5 Middleware — Orden de Ejecución

```
RequestID → RealIP → RequestLogger → Recoverer → CORS → AuthMiddleware → AdminOnly
```

`RequestID` debe ejecutarse ANTES de `RequestLogger` para que el request_id esté disponible.

### 9.6 Seguridad — Campos PROHIBIDOS en Logs

`password`, `password_hash`, `token`, `jwt`, `secret`, `authorization`
Cuerpos completos de request/response, PII sin enmascarar.

### 9.7 Pasos Futuros

- Métricas con OpenTelemetry → Cloud Monitoring
- Alertas en Cloud Monitoring por tasa de errores 5xx
- Slow query logging en repositorios (queries > 500ms)
