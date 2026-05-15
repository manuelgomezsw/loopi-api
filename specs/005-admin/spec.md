# Especificación Funcional: Panel de Administración

**Feature Branch**: `feature/admin` (migrado desde `master`)
**Estado**: migrated
**Fecha de migración**: 2026-05-14

## User Scenarios & Testing

### User Story 1 — Dashboard (Prioridad: P1)

El administrador consulta el estado del día: cuántos inventarios se realizaron, cuántos tienen discrepancias y cuántos están pendientes.

**Por qué P1**: Es la vista principal del panel admin — da visibilidad inmediata del estado operativo.

**Escenarios de aceptación**:

1. **Dado** 3 inventarios completados hoy (2 con discrepancias) y 1 en progreso, **Cuando** `GET /api/admin/dashboard`, **Entonces** `{stats:{today_inventories:3, with_discrepancies:2, without_discrepancies:1, pending_inventories:1}}`.
2. **Dado** sin inventarios hoy, **Entonces** todos los contadores en 0.

---

### User Story 2 — Gestión de Inventarios (Admin) (Prioridad: P1)

El administrador lista, filtra y edita inventarios completados, pudiendo corregir valores de detalle y crear el inventario inicial (baseline).

**Escenarios de aceptación**:

1. **Cuando** `GET /api/admin/inventories?page=1&page_size=20`, **Entonces** lista paginada con `total`, `page`, `total_pages`.
2. **Dado** filtros por fecha, tipo o empleado, **Cuando** `GET /api/admin/inventories?date_from=...&employee_id=...`, **Entonces** lista filtrada.
3. **Cuando** `GET /api/admin/inventories/{id}`, **Entonces** detalle completo con todos los ítems y sus valores.
4. **Cuando** `PUT /api/admin/inventories/{id}/details/{detailID}` con campos a corregir, **Entonces** HTTP 200.
5. **Cuando** `POST /api/admin/inventories/initial`, **Entonces** crea un inventario de tipo `initial` para el día.
6. **Cuando** `GET /api/admin/inventories/active-count`, **Entonces** cantidad de inventarios `in_progress` en este momento.

---

### User Story 3 — Gestión de Ítems (Prioridad: P1)

El administrador gestiona el catálogo de ítems que se inventarían: crear, editar, activar/desactivar, asignar categoría, proveedor y frecuencia.

**Escenarios de aceptación**:

1. **Cuando** `GET /api/admin/items?page=1&page_size=20`, **Entonces** lista paginada con filtros opcionales (`type`, `frequency`, `active`, `search`).
2. **Cuando** `POST /api/admin/items` con datos válidos, **Entonces** HTTP 201 con ítem creado. Si `add_to_active_inventories=true`, el ítem se agrega a todos los inventarios `in_progress`.
3. **Cuando** `PUT /api/admin/items/{id}`, **Entonces** actualiza todos los campos del ítem.
4. **Cuando** `PATCH /api/admin/items/{id}/status` con `{active: false}`, **Entonces** desactiva el ítem.
5. **Cuando** `GET /api/admin/measurement-units`, **Entonces** lista todas las unidades de medida activas.

---

### User Story 4 — Gestión de Empleados (Prioridad: P1)

El administrador gestiona el personal: crear empleados, editar datos, activar/desactivar y resetear contraseñas.

**Escenarios de aceptación**:

1. **Cuando** `GET /api/admin/employees?role=employee&active=true&page=1`, **Entonces** lista paginada con filtros.
2. **Cuando** `POST /api/admin/employees` con `{username, password, name, last_name, role}`, **Entonces** HTTP 201 con empleado creado (password hasheada con bcrypt).
3. **Cuando** `PUT /api/admin/employees/{id}`, **Entonces** actualiza datos del empleado (sin cambiar password).
4. **Cuando** `PATCH /api/admin/employees/{id}/status`, **Entonces** activa o desactiva el empleado.
5. **Cuando** `POST /api/admin/employees/{id}/reset-password`, **Entonces** resetea la contraseña a un valor generado.
6. **Cuando** `GET /api/admin/employees/active`, **Entonces** lista todos los empleados activos (sin paginación, para selects de UI).

---

### User Story 5 — Gestión de Categorías (Prioridad: P2)

El administrador gestiona las categorías de ítems, incluyendo su orden de visualización.

**Escenarios de aceptación**:

1. **Cuando** `GET /api/admin/categories`, **Entonces** lista de categorías ordenadas por `display_order`.
2. **Cuando** `POST /api/admin/categories` con `{name}`, **Entonces** HTTP 201 con categoría creada.
3. **Cuando** `PUT /api/admin/categories/{id}` con `{name, active}`, **Entonces** actualiza la categoría.
4. **Cuando** `PATCH /api/admin/categories/{id}/status`, **Entonces** activa o desactiva.
5. **Cuando** `POST /api/admin/categories/reorder` con `[{id, display_order}]`, **Entonces** actualiza el orden de múltiples categorías en una operación.

---

### User Story 6 — Gestión de Proveedores (Prioridad: P2)

El administrador gestiona el catálogo de proveedores asociados a los ítems.

**Escenarios de aceptación**:

1. **Cuando** `GET /api/admin/suppliers?active=true&search=...&page=1`, **Entonces** lista paginada filtrada.
2. **Cuando** `POST /api/admin/suppliers` con `{business_name, tax_id, contact_name, contact_phone, contact_email}`, **Entonces** HTTP 201.
3. **Cuando** `PUT /api/admin/suppliers/{id}`, **Entonces** actualiza todos los campos.
4. **Cuando** `PATCH /api/admin/suppliers/{id}/status`, **Entonces** activa o desactiva.
5. **Cuando** `GET /api/admin/suppliers/active`, **Entonces** lista todos los activos (para selects de UI).

---

### User Story 7 — Unidades de Medida (Prioridad: P3)

El administrador consulta las unidades de medida disponibles para asignar a ítems.

**Escenarios de aceptación**:

1. **Cuando** `GET /api/admin/measurement-units`, **Entonces** lista de todas las unidades activas (`id`, `name`, `abbreviation`).

---

### Edge Cases

- Todos los endpoints admin requieren `role == "admin"` via `middleware.AdminOnly` → HTTP 403 si es empleado.
- Cualquier error de BD retorna HTTP 500 con log ERROR en el servicio (no en el handler).
- `CreateEmployee` con username duplicado → error de BD de unicidad → HTTP 409 o HTTP 500 según manejo.

## Requisitos Funcionales

- **FR-001**: TODOS los endpoints `/api/admin/*` DEBEN estar protegidos por `middleware.AdminOnly`.
- **FR-002**: Todas las listas DEBEN soportar paginación (`page`, `page_size`).
- **FR-003**: `CreateEmployee` DEBE hashear la contraseña con bcrypt antes de guardarla.
- **FR-004**: `ResetEmployeePassword` DEBE generar una nueva contraseña y hashearla.
- **FR-005**: `CreateItem` con `add_to_active_inventories=true` DEBE agregar el ítem a todos los inventarios `in_progress`.
- **FR-006**: `ReorderCategories` DEBE actualizar `display_order` de múltiples categorías en una sola operación.
- **FR-007**: El Dashboard DEBE mostrar stats del día actual únicamente.
- **FR-008**: `AdminService` DEBE ser un facade que delega a servicios especializados por sub-dominio.

### Entidades clave

- **Item**: `id`, `type` (`product`|`ingredient`), `name`, `inventory_frequency` (`daily`|`weekly`|`monthly`|`all`), `category_id`, `supplier_id`, `cost`, `measurement_unit_id`, `active`.
- **Category**: `id`, `name`, `display_order`, `active`.
- **Supplier**: `id`, `business_name`, `tax_id`, `contact_name`, `contact_phone`, `contact_email`, `active`.
- **MeasurementUnit**: `id`, `name`, `abbreviation`, `active`.

## Puntos de Integración

| Módulo | Ruta | Tipo de interacción |
|--------|------|---------------------|
| AdminService (facade) | `internal/application/service/admin_service.go` | Delega a 6 sub-servicios |
| InventoryRepository | `internal/domain/repository/inventory_repository.go` | Dashboard, ListInventories, GetInventoryDetail |
| EmployeeRepository | `internal/domain/repository/employee_repository.go` | CRUD empleados, reset password |
| ItemRepository | `internal/domain/repository/item_repository.go` | CRUD ítems, agregar a inventarios activos |
| CategoryRepository | `internal/domain/repository/category_repository.go` | CRUD categorías, reorden |
| SupplierRepository | `internal/domain/repository/supplier_repository.go` | CRUD proveedores |
| MeasurementUnitRepository | `internal/domain/repository/measurement_unit_repository.go` | List |
| Enricher | `internal/domain/inventory/enrich.go` | Estadísticas de discrepancias en listas |
| AdminOnly middleware | `internal/interface/middleware/auth_middleware.go` | Guard de acceso |

## Contrato HTTP — Endpoints Admin

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/api/admin/dashboard` | Stats del día |
| GET | `/api/admin/inventories` | Lista paginada + filtros |
| GET | `/api/admin/inventories/active-count` | Count in_progress |
| POST | `/api/admin/inventories/initial` | Crear inventario inicial |
| GET | `/api/admin/inventories/{id}` | Detalle completo |
| PUT | `/api/admin/inventories/{id}/details/{detailID}` | Editar detalle |
| GET | `/api/admin/measurement-units` | Lista unidades |
| GET | `/api/admin/items` | Lista paginada + filtros |
| POST | `/api/admin/items` | Crear ítem |
| GET | `/api/admin/items/{id}` | Detalle |
| PUT | `/api/admin/items/{id}` | Editar |
| PATCH | `/api/admin/items/{id}/status` | Activar/desactivar |
| GET | `/api/admin/employees` | Lista paginada + filtros |
| GET | `/api/admin/employees/active` | Todos activos (sin paginar) |
| POST | `/api/admin/employees` | Crear |
| GET | `/api/admin/employees/{id}` | Detalle |
| PUT | `/api/admin/employees/{id}` | Editar |
| PATCH | `/api/admin/employees/{id}/status` | Activar/desactivar |
| POST | `/api/admin/employees/{id}/reset-password` | Resetear contraseña |
| GET | `/api/admin/categories` | Lista |
| POST | `/api/admin/categories` | Crear |
| POST | `/api/admin/categories/reorder` | Reordenar |
| GET | `/api/admin/categories/{id}` | Detalle |
| PUT | `/api/admin/categories/{id}` | Editar |
| PATCH | `/api/admin/categories/{id}/status` | Activar/desactivar |
| GET | `/api/admin/suppliers` | Lista paginada + filtros |
| GET | `/api/admin/suppliers/active` | Todos activos |
| POST | `/api/admin/suppliers` | Crear |
| GET | `/api/admin/suppliers/{id}` | Detalle |
| PUT | `/api/admin/suppliers/{id}` | Editar |
| PATCH | `/api/admin/suppliers/{id}/status` | Activar/desactivar |

## Migración de Base de Datos

| # | Archivo | Tipo | Descripción |
|---|---------|------|-------------|
| 001 | `migrations/001_initial_schema.up.sql` | CREATE TABLE | `employees`, `items`, `inventories`, `inventory_details` |
| 005 | `migrations/005_categories.up.sql` | CREATE TABLE | `categories` con `display_order` |
| 006 | `migrations/006_suppliers.up.sql` | CREATE TABLE | `suppliers` |
| 007 | `migrations/007_items_add_category_supplier_cost.up.sql` | ALTER TABLE | `category_id`, `supplier_id`, `cost`, `measurement_unit_id` en `items` |
| 008 | `migrations/008_document_type_nuip.up.sql` | ALTER TABLE | `document_type`, `document_number` en `employees` |
| 012 | `migrations/012_measurement_units.up.sql` | CREATE TABLE | `measurement_units` |
| 013 | `migrations/013_create_schemas.sql` | Multiple | Esquemas adicionales |

## Criterios de Éxito

- **SC-001**: Todos los endpoints admin retornan HTTP 403 para tokens de rol `employee`.
- **SC-002**: Paginación consistente en todas las listas — `total`, `page`, `page_size`, `total_pages`.
- **SC-003**: `CreateEmployee` nunca expone `password_hash` en respuesta ni en logs.
- **SC-004**: Dashboard refleja el estado real del día en tiempo real.

## Assumptions

- Solo existe un rol admin — no hay sub-roles dentro del admin.
- El admin puede ver y editar inventarios de todos los empleados.
- `ResetEmployeePassword` genera contraseña temporal — el flujo de cambio por el propio empleado no está implementado.
