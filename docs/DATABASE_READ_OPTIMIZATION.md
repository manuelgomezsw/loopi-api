# Modelado de base de datos para reducir consultas y cálculos en lectura

Este documento razona sobre **persistir información consolidada** (totales, conteos de discrepancia) en lugar de recalcular en cada lectura, y propone alternativas de modelado en el contexto actual del proyecto.

---

## 1. Contexto actual

### Qué se calcula hoy en lectura

- **Listado de inventarios (admin):** para cada inventario se necesita `total_items` y `items_with_diff`. Hoy se obtienen cargando todos los `inventory_details` del inventario, aplicando Enricher (suggested desde inventario anterior), y evaluando la regla de dominio `ExpectedForAdmin` + `HasDiscrepancyFromExpectedEnd`. Una query de listado + N queries de detalles (o un batch que aún no existe).
- **Dashboard:** conteo de inventarios “con discrepancias” del día se hace cargando inventarios completados del día y, para cada uno, contando ítems con discrepancia (misma lógica). Además se construye la lista de “discrepancias recientes” desde detalles recientes, enriqueciendo y filtrando por regla de dominio.
- **Al cerrar inventario:** `CompleteInventory` ya calcula `discrepancyCount` en memoria con la misma regla (`ExpectedAtEnd` + `HasDiscrepancyFromExpectedEnd`) pero **no persiste** ese número; solo hace `UPDATE inventories SET status = 'completed', completed_at = NOW()`.

La entidad `Inventory` tiene campos `TotalItems` e `ItemsWithDiff` en memoria (“computed fields for admin views”) que se rellenan en cada request, no en base de datos.

### Dependencia de la regla de “expected”

- **Vista empleado / cierre:** `ExpectedAtEnd(d) = suggested + stock_received - units_sold` (sin shrinkage).
- **Vista admin:** `ExpectedForAdmin(d) = suggested - shrinkage + stock_received - units_sold`.

Si el admin edita `real_value`, `stock_received`, `units_sold` o `shrinkage` en un inventario ya cerrado (`UpdateInventoryDetail`), el valor “expected” y por tanto “tiene discrepancia” puede cambiar. Cualquier consolidado que persista totales/conteos debe definir **qué pasa cuando el admin edita**: actualizar el consolidado o aceptar desfase hasta un próximo recálculo.

---

## 2. Enfoque: consolidar al cierre (write-time)

La idea es: **en el momento en que el inventario se marca como completado**, calcular y guardar los agregados que hoy se recalculan en cada lectura. Así:

- Las lecturas (listado, dashboard) leen columnas o filas ya calculadas y evitan N consultas a `inventory_details` y el Enricher para esos casos.
- El costo de cómputo se paga una vez (al cerrar) en lugar de en cada request.

Condición importante: **cuando el admin edita un detalle de un inventario completado**, hay que decidir si se actualiza el consolidado para mantener consistencia o se deja “fijo” el snapshot del cierre.

---

## 3. Alternativas de modelado

### 3.1 Columnas en la tabla `inventories`

**Qué hacer:** agregar a `inventories` columnas como `total_items` (SMALLINT UNSIGNED) e `items_with_diff` (SMALLINT UNSIGNED), pobladas al llamar a `Complete`.

**Ventajas:**

- Una sola tabla; listado y dashboard leen directamente `inventories` sin tocar `inventory_details` para los totales.
- Implementación simple: en `Complete` (o en el caso de uso que cierra) además del `UPDATE status/completed_at` hacer un `UPDATE` con los valores calculados (total_items = len(details), items_with_diff = discrepancyCount).
- Migración acotada: una migración ADD COLUMN (nullable al inicio, rellenar con backfill si se desea para datos históricos).

**Desventajas / consideraciones:**

- Si el admin edita un detalle después del cierre, los totales pueden quedar desactualizados a menos que:
  - **Opción A:** cada vez que se llame a `UpdateInventoryDetail` para un inventario completado, se recalculen total_items e items_with_diff para ese inventario y se haga un `UPDATE inventories SET total_items = ?, items_with_diff = ? WHERE id = ?`. Consistencia fuerte, un poco más de lógica en el update.
  - **Opción B:** no actualizar estas columnas tras ediciones; aceptar que son “snapshot al cierre” y que en vistas de “resumen histórico” puedan verse desfasadas hasta que se agregue un job de recálculo (opcional).
- Solo resuelve totales por inventario; no resuelve por sí solo la lista de “discrepancias recientes” del dashboard (ahí seguís necesitando detalles o una tabla de snapshot de discrepancias).

**Recomendación:** Es la opción más directa para eliminar las N consultas de “conteo por inventario” en listado y en el bloque “with/without discrepancies” del dashboard. Conviene definir desde el inicio la política (actualizar en UpdateInventoryDetail vs snapshot fijo).

---

### 3.2 Tabla de consolidado por inventario (`inventory_summaries` o similar)

**Qué hacer:** una tabla 1:1 con inventarios completados, por ejemplo `inventory_summaries (inventory_id PK, total_items, items_with_diff, created_at)`, escrita al cerrar el inventario.

**Ventajas:**

- Separa datos transaccionales (inventories, inventory_details) de datos derivados; el esquema de `inventories` no se ensancha.
- Permite extender el resumen con más métricas (por ejemplo total_value_at_cost, cantidad de ítems con stock_received > 0, etc.) sin tocar la tabla principal.
- Podés tener políticas distintas: por ejemplo solo rellenar para status = completed (inserción en el mismo flujo de Complete).

**Desventajas:**

- Una JOIN (o una segunda query) en cada listado/dashboard que use el resumen.
- La misma cuestión de consistencia cuando el admin edita un detalle: actualizar la fila de resumen en UpdateInventoryDetail o dejarla como snapshot.

**Cuándo preferirla:** Si querés mantener `inventories` “puro” y/o preves muchos más campos consolidados por inventario en el futuro, esta opción escala mejor que poner muchas columnas en `inventories`.

---

### 3.3 Tabla de snapshot de discrepancias (para reportes / “recientes”)

**Qué hacer:** al cerrar, además de (o en lugar de) solo totales, persistir **qué ítems tuvieron discrepancia**: por ejemplo `inventory_discrepancy_snapshots (inventory_id, item_id, expected_value, actual_value, difference, inventory_date, inventory_type)` con una fila por par (inventory_id, item_id) con discrepancia.

**Ventajas:**

- El dashboard “discrepancias recientes” puede leer de esta tabla (con límite y orden por fecha) **sin** cargar todos los details ni enriquecer; una o dos queries acotadas.
- Sirve como registro histórico “así quedó al cierre”; útil para reportes y auditoría.

**Desventajas:**

- Duplicación de lógica: al cerrar tenés que escribir en esta tabla usando la misma regla de expected/discrepancy que el dominio. Si la regla cambia, los datos nuevos serán correctos pero los históricos en la tabla reflejan la regla antigua (habitualmente aceptable para snapshot).
- Si el admin edita un detalle después del cierre, tenés que decidir: actualizar/insertar/borrar en esta tabla para ese item (consistencia) o no tocarla (snapshot fijo al cierre).

**Recomendación:** Muy útil si el cuello de botella principal es el armado de “recent discrepancies” con muchos details y enriquecimiento. Se puede combinar con 3.1 o 3.2: columnas o tabla de totales para listado/dashboard de conteos, y esta tabla para la lista de recientes.

---

### 3.4 “Vista materializada” simulada (tabla + actualización controlada)

MySQL no tiene vistas materializadas nativas. Se puede simular:

- Una tabla (por ejemplo `inventory_list_view` o reutilizar `inventory_summaries`) con columnas precalculadas.
- Actualización en momentos bien definidos:
  - **Al cierre:** insert/update en esa tabla (mismo flujo que 3.1/3.2).
  - **Al editar detalle (admin):** update de la fila afectada (recalcular ese inventario).
  - Opcional: un job nocturno que recalculase filas “sospechosas” o todas.

Es el mismo concepto que 3.1/3.2 pero pensado como “capa de lectura materializada”: la aplicación lee siempre de esta tabla para listados/dashboard y no toca details salvo para detalle completo o edición.

---

### 3.5 No persistir consolidado: optimizar solo las lecturas (batch + cache)

**Alternativa:** no añadir columnas ni tablas; seguir calculando en lectura pero:

- **Batch:** como en REFACTORING_GUIDELINES.md, métodos en repositorio que traigan muchos details por muchos inventory_ids en una o dos queries; el servicio calcula total_items/items_with_diff en memoria. Reduce round-trips pero no elimina el cálculo.
- **Cache (ej. Redis):** clave por `inventory_id` (o por “dashboard:date”) con total_items e items_with_diff; TTL corto o invalidación al cerrar inventario / al editar detalle. Las lecturas leen del cache cuando existe.

**Ventajas:** Sin cambios de esquema; consistencia más fácil (el cálculo siempre usa los detalles actuales o el cache se invalida).  
**Desventajas:** Siguen existiendo cálculos y posiblemente muchas lecturas a details si el cache falla o no aplica; la latencia del primer request o tras invalidación sigue siendo mayor.

---

## 4. Consistencia cuando el admin edita un detalle

Tras `UpdateInventoryDetail` en un inventario ya completado, los campos que pueden cambiar son `real_value`, `stock_received`, `units_sold`, `shrinkage`. Con eso, `ExpectedForAdmin` y “tiene discrepancia” pueden cambiar para ese ítem.

Opciones:

| Política | Descripción | Pros / contras |
|----------|--------------|-----------------|
| **Actualizar consolidado en cada edición** | En UpdateInventoryDetail, después de actualizar el detail, recalculá total_items e items_with_diff para ese inventario (cargando sus details, aplicando dominio) y actualizá `inventories` o `inventory_summaries`. | Lecturas siempre coherentes con el detalle. Coste extra en el path de edición (una lectura de details + un update de resumen). |
| **Snapshot fijo al cierre** | No actualizar total_items/items_with_diff (ni snapshot de discrepancias) cuando el admin edita. | Implementación más simple; listado/dashboard pueden mostrar valores “al cierre”. Si en otra pantalla se muestra el detalle actual, puede haber diferencia de “número de ítems con diff” entre listado y detalle. |
| **Invalidar cache** | Si usás solo cache (3.5), al editar invalidar la entrada para ese inventory_id para que la próxima lectura recalcule. | No aplica si no usás cache; si usás cache + consolidado, podés combinar: consolidado al cierre + invalidación de cache en edición y recálculo bajo demanda. |

Recomendación práctica: si adoptás columnas en `inventories` o tabla de resumen, **actualizar el consolidado en UpdateInventoryDetail** mantiene el modelo simple y evita sorpresas en la UI (mismo número en listado y en detalle).

---

## 5. Otras alternativas (complementarias o sustitutas)

- **Réplicas de lectura:** Distribuyen la carga pero no reducen la cantidad de consultas ni los cálculos; siguen siendo útiles para escalar lecturas una vez que las queries estén optimizadas.
- **Índices y queries batch:** Asegurar índices adecuados en `inventory_details (inventory_id)` y, si se usa, en la tabla de snapshots; junto con `FindDetailsByInventoryIDs` (batch) ya mejoran mucho sin tocar el modelo de consolidado.
- **Eventos de dominio:** Al “InventoryCompleted” publicar un evento que un consumidor use para rellenar una tabla de reportes o un cache; el cierre no hace el write del consolidado, lo hace un listener. Útil si más adelante tenés varios consumidores (analytics, notificaciones, etc.).

---

## 6. Resumen y sugerencia

| Objetivo | Enfoque recomendado |
|----------|---------------------|
| Reducir consultas y cálculos en **listado** y en **conteos del dashboard** (with/without discrepancies) | Persistir **total_items** e **items_with_diff** (en `inventories` o en `inventory_summaries`) **al cerrar** y, si el admin puede editar detalles de inventarios cerrados, **actualizar ese consolidado en UpdateInventoryDetail**. |
| Reducir el costo de **“discrepancias recientes”** (muchos details + enricher) | Tabla de **snapshot de discrepancias** al cierre (por inventory_id + item_id), leída por el dashboard con límite; opcionalmente actualizada o no cuando el admin edita (según si querés “recientes” siempre al día o snapshot histórico). |
| Evitar tocar esquema | Optimizar solo con **batch** en repositorio + opcional **cache** (3.5). |

En términos de modelado de base de datos para “evitar tantos llamados y consultas en línea” cuando el inventario se cierra, la opción más alineada con tu idea es: **consolidar en write-time** usando columnas en `inventories` (o una tabla de resumen 1:1) para totales y, si hace falta, una tabla de snapshot de discrepancias para el listado de recientes; y definir una política clara de actualización del consolidado cuando el admin edita un detalle de un inventario completado.
