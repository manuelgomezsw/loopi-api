# Especificación Funcional: Inventario del Empleado

**Feature Branch**: `feature/employee-inventory` (migrado desde `master`)
**Estado**: migrated
**Fecha de migración**: 2026-05-14

## User Scenarios & Testing

### User Story 1 — Crear inventario (Prioridad: P1)

Un empleado inicia un nuevo inventario para la fecha y turno actuales. El sistema pre-popula automáticamente los ítems con valores sugeridos basados en el conteo real del inventario anterior.

**Por qué P1**: Es la operación de entrada del flujo principal — sin crear un inventario no existe ninguna otra operación.

**Test independiente**: `InventoryService.CreateInventory` con fake repos. Verificar que se llama `CreateBatch` con detalles pre-poblados y el valor sugerido correcto.

**Escenarios de aceptación**:

1. **Dado** ningún inventario existente para la fecha/tipo/turno, **Cuando** `POST /api/inventories` con `type=daily, schedule=opening`, **Entonces** HTTP 201 con el inventario creado (status `in_progress`) y sus ítems pre-poblados.
2. **Dado** un inventario `in_progress` para la misma fecha/tipo/turno, **Cuando** `POST /api/inventories`, **Entonces** HTTP 201 retornando el inventario existente (idempotente).
3. **Dado** un inventario ya `completed` para la misma fecha/tipo/turno, **Cuando** `POST /api/inventories`, **Entonces** HTTP 409 Conflict.
4. **Dado** `type=daily` sin `schedule`, **Cuando** `POST /api/inventories`, **Entonces** HTTP 400 `"schedule is required for daily inventories"`.
5. **Dado** `type=weekly`, **Cuando** `POST /api/inventories`, **Entonces** el campo `schedule` se ignora (se establece en `nil`).

---

### User Story 2 — Consultar turno sugerido (Prioridad: P2)

El cliente consulta qué tipo de inventario y turno corresponde a la hora actual para pre-seleccionar en la UI.

**Por qué P2**: Mejora la UX — el empleado no tiene que elegir el turno manualmente.

**Test independiente**: `InventoryService.GetSuggestedSchedule` — función pura basada en la hora actual (`datetime.Now()`).

**Escenarios de aceptación**:

1. **Dado** hora actual entre 06:00–10:59, **Cuando** `GET /api/inventories/suggested-schedule`, **Entonces** `{type:"daily", schedule:"opening"}`.
2. **Dado** hora actual entre 11:00–15:59, **Entonces** `schedule:"noon"`.
3. **Dado** hora actual entre 16:00–21:59, **Entonces** `schedule:"closing"`.
4. **Dado** hora fuera de rango (22:00–05:59), **Entonces** `schedule:"opening"` (default).

---

### User Story 3 — Contar y guardar ítems (Prioridad: P1)

El empleado registra el conteo físico de cada ítem del inventario.

**Por qué P1**: Es la operación central del inventario.

**Test independiente**: `InventoryService.SaveInventoryDetail` con fake repos verificando que `real_value` se actualiza y que el inventario completado es rechazado.

**Escenarios de aceptación**:

1. **Dado** inventario `in_progress` con ítem válido, **Cuando** `POST /api/inventories/{id}/details` con `{item_id, real_value}`, **Entonces** HTTP 200 `{saved:true, suggested_value}`.
2. **Dado** inventario ya `completed`, **Cuando** guardar detalle, **Entonces** HTTP 400 `"inventory is already completed"`.
3. **Dado** `item_id` inexistente en el inventario, **Cuando** guardar detalle, **Entonces** HTTP 400 `"item not found in this inventory"`.

---

### User Story 4 — Registrar ventas y compras (Prioridad: P2)

Para inventarios que lo requieren (noon, closing, weekly, monthly), el empleado registra las unidades vendidas y el stock recibido, actualizando el valor sugerido en tiempo real.

**Por qué P2**: Permite al sistema calcular el valor esperado al final del período para detectar discrepancias.

**Escenarios de aceptación**:

1. **Dado** inventario `daily/closing` (requiere ventas), **Cuando** `POST /api/inventories/{id}/sales` con `{item_id, stock_received, units_sold}`, **Entonces** HTTP 200 con `suggested_value` recalculado: `sugerido_anterior + stock_received - units_sold`.
2. **Dado** inventario `weekly` (solo compras), **Cuando** `POST /api/inventories/{id}/sales`, **Entonces** `units_sold` se fuerza a `0` — solo `stock_received` afecta el sugerido.
3. **Dado** inventario `daily/opening` (no requiere ventas), **Cuando** `POST /api/inventories/{id}/sales`, **Entonces** HTTP 400.

---

### User Story 5 — Ver discrepancias y completar inventario (Prioridad: P1)

El empleado revisa los ítems con diferencia entre conteo real y valor esperado, y cierra el inventario cuando todos los ítems están contados.

**Por qué P1**: El cierre del inventario es la condición de finalización del flujo.

**Test independiente**: `InventoryService.GetDiscrepancies` y `CompleteInventory` con fake repos — cubiertos por `inventory_service_integration_test.go`.

**Escenarios de aceptación**:

1. **Dado** inventario con todos los ítems contados, **Cuando** `POST /api/inventories/{id}/complete`, **Entonces** HTTP 200 `{completed:true, issues_created: N}` donde N = ítems con `real_value != expected_at_end`.
2. **Dado** un ítem sin contar (`real_value == nil`), **Cuando** completar, **Entonces** HTTP 400 `"not all items have been inventoried"`.
3. **Dado** inventario ya completado, **Cuando** completar de nuevo, **Entonces** HTTP 400 `"inventory is already completed"`.
4. **Dado** inventario con discrepancias, **Cuando** `GET /api/inventories/{id}/discrepancies`, **Entonces** solo los ítems donde `real_value != expected_at_end`.

---

### User Story 6 — Consultar inventarios recientes (Prioridad: P3)

El empleado consulta el último inventario completado y sus inventarios en progreso.

**Escenarios de aceptación**:

1. **Cuando** `GET /api/inventories/latest`, **Entonces** el inventario más reciente con status `completed`, o `{"inventory": null}` si no existe ninguno.
2. **Cuando** `GET /api/inventories/in-progress`, **Entonces** lista de inventarios `in_progress` asignados al empleado autenticado.

---

### Edge Cases

- Inventario `initial`: no genera discrepancias en `CompleteInventory` (es el inventario base de referencia).
- Valor sugerido se calcula solo con `real_value` anterior — mermas (`shrinkage`) no restan al sugerido.
- `expected_at_end = suggested_value + stock_received - units_sold` (inventarios con ventas) o `= suggested_value + stock_received` (solo compras).

## Requisitos Funcionales

- **FR-001**: El sistema DEBE pre-poblar detalles al crear un inventario, usando el `real_value` del inventario anterior como `suggested_value`.
- **FR-002**: Para inventarios `daily`, el campo `schedule` DEBE ser obligatorio.
- **FR-003**: El sistema DEBE ser idempotente en creación: si existe un inventario `in_progress` para la misma fecha/tipo/turno, retornarlo sin crear uno nuevo.
- **FR-004**: El sistema DEBE bloquear modificaciones a inventarios con status `completed`.
- **FR-005**: `CompleteInventory` DEBE fallar si algún ítem tiene `real_value == nil`.
- **FR-006**: Inventarios `weekly` y `monthly` DEBEN ignorar `units_sold` (forzar a 0) al guardar ventas/compras.
- **FR-007**: `GetDiscrepancies` DEBE retornar solo ítems donde `real_value != expected_at_end`.
- **FR-008**: La lógica de `expected_at_end` y `has_discrepancy` DEBE vivir en `internal/domain/inventory/` (funciones puras, testeable sin BD).

### Entidades clave

- **Inventory**: `id`, `inventory_date`, `inventory_type` (`daily`|`weekly`|`monthly`|`initial`), `schedule` (`opening`|`noon`|`closing`, nullable), `status` (`in_progress`|`completed`), `responsible_id`, `started_at`, `completed_at`.
- **InventoryDetail**: `id`, `inventory_id`, `item_id`, `suggested_value`, `real_value`, `stock_received`, `units_sold`, `shrinkage`.

## Puntos de Integración

| Módulo | Ruta | Tipo de interacción |
|--------|------|---------------------|
| InventoryRepository | `internal/domain/repository/inventory_repository.go` | FindByID, FindByDateTypeAndSchedule, FindPreviousInventory, Create, Complete |
| InventoryDetailRepository | `internal/domain/repository/inventory_repository.go` | FindByInventoryID, FindByInventoryAndItem, Update, CreateBatch |
| ItemRepository | `internal/domain/repository/item_repository.go` | FindActiveByInventoryType |
| Domain logic | `internal/domain/inventory/` | expected.go, discrepancy.go, enrich.go |
| OTel metric | `pkg/observability/` | Counter `loopi.inventory.movements` (action: create/complete) |

## Migración de Base de Datos

| # | Archivo | Tipo | Descripción |
|---|---------|------|-------------|
| 001 | `migrations/001_initial_schema.up.sql` | CREATE TABLE | `inventories`, `inventory_details`, `items`, `employees` |
| 003 | `migrations/003_inventory_frequency.up.sql` | ALTER TABLE | Agrega `inventory_type` e índice único fecha/tipo/turno a `inventories` |
| 009 | `migrations/009_initial_inventory_type.up.sql` | ALTER TABLE | Agrega valor `initial` al enum `inventory_type` |
| 010 | `migrations/010_inventory_details_shrinkage.up.sql` | ALTER TABLE | Agrega columna `shrinkage` a `inventory_details` |
| 013 | `migrations/013_create_schemas.sql` | CREATE TABLE | Esquemas adicionales (sin down migration) |

## Criterios de Éxito

- **SC-001**: Crear → contar todos los ítems → completar: flujo completo sin errores.
- **SC-002**: `CompleteInventory` retorna conteo correcto de discrepancias (validado en `inventory_service_integration_test.go`).
- **SC-003**: `GetDiscrepancies` retorna exactamente los mismos ítems que `CompleteInventory` contaría como discrepantes.
- **SC-004**: La lógica de `expected_at_end` pasa los tests de dominio en `internal/domain/inventory/expected_test.go`.
- **SC-005**: El metric OTel `loopi.inventory.movements` se incrementa en cada `Create` y `Complete`.

## Assumptions

- Solo un responsable por inventario (`responsible_id` = empleado autenticado al crear).
- El inventario `initial` es el inventario base — no detecta discrepancias, sirve como punto de referencia para `suggested_value`.
- Los ítems se filtran por `inventory_type` al pre-poblar (`FindActiveByInventoryType`).
