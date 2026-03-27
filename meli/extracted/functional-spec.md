# Especificación Funcional — Loopi

**Versión**: 1.0.0 (extraída de código + documentación)
**Fecha**: 2026-03-26
**Idioma**: Español
**Confianza**: Los indicadores reflejan la fuente de cada dato.

> **Leyenda de confianza:**
> - ✅✅ VERIFIED — Confirmado en código + documentación, coincide
> - 🔸 CODE_ONLY — Solo en código (confiable pero no documentado)
> - ⚠️ DOCS_ONLY — Solo en docs (planificado, NO implementado aún)
> - ❓ UNKNOWN — Información insuficiente

---

## 1. Descripción del Sistema

**Loopi** es una aplicación web de gestión de inventario para negocios de retail/alimentos. Permite a los empleados registrar conteos físicos de productos e insumos en múltiples turnos del día, y a los administradores supervisar, auditar y gestionar toda la información del negocio.

**Contexto de negocio**: ✅✅
- Negocio en Colombia (moneda COP, zona horaria America/Bogota)
- Inventarios se realizan en múltiples turnos: apertura, mediodía y cierre
- Los empleados cuentan físicamente los items; el sistema compara con el esperado
- Las discrepancias (diferencias entre conteo real y esperado) son el indicador principal de calidad del proceso

---

## 2. Contexto del Sistema

### 2.1 Actores ✅✅

| Actor | Tipo | Descripción |
|-------|------|-------------|
| **Empleado** | Humano (interno) | Usuario operativo. Realiza conteos de inventario, registra ventas y compras durante su turno. |
| **Administrador** | Humano (interno) | Usuario de gestión. Supervisa inventarios, edita datos, gestiona catálogos (items, empleados, categorías, proveedores). |
| **Frontend Angular** | Sistema | Aplicación web SPA. Consuma el API REST del backend. Hospedada en Firebase Hosting. |

### 2.2 Dependencias del Sistema ✅✅

| Servicio | Tipo | Propósito |
|---------|------|-----------|
| **MySQL (Cloud SQL)** | Base de datos | Almacenamiento principal de todos los datos |
| **Google App Engine** | Hosting backend | Despliegue del API REST Go |
| **Firebase Hosting** | Hosting frontend | Despliegue de la SPA Angular |

### 2.3 Integraciones Externas 🔸

No se detectaron integraciones con sistemas externos (no webhooks, no colas de mensajes, no APIs de terceros). El sistema es autónomo.

---

## 3. Casos de Uso

### Flujo Operativo del Empleado

#### UC-001: Autenticación ✅✅
**Actor**: Empleado, Administrador
**Precondición**: Usuario con cuenta activa en el sistema
**Flujo**:
1. El usuario ingresa su nombre de usuario y contraseña
2. El sistema valida las credenciales (bcrypt)
3. El sistema retorna un JWT válido por 24 horas (configurable)
4. El usuario queda autenticado en la aplicación

**Reglas**:
- Solo usuarios con estado `active = true` pueden autenticarse
- Las contraseñas se almacenan con hash bcrypt (cost 12)
- El JWT contiene: employee_id, username, role

#### UC-002: Ver Schedule Sugerido 🔸
**Actor**: Empleado
**Precondición**: Empleado autenticado
**Flujo**:
1. El sistema determina el turno sugerido según la hora actual (zona horaria Colombia):
   - 06:00–11:00 → `opening`
   - 11:00–16:00 → `noon`
   - 16:00–22:00 → `closing`
2. Retorna el tipo de inventario sugerido (`daily`) y el schedule correspondiente

#### UC-003: Crear Inventario ✅✅
**Actor**: Empleado, Administrador
**Precondición**: Empleado autenticado
**Flujo**:
1. El empleado selecciona tipo de inventario (`daily`, `weekly`, `monthly`) y schedule (para `daily`: `opening`, `noon`, `closing`)
2. El sistema verifica que no exista ya un inventario `in_progress` con la misma combinación (fecha, tipo, schedule)
   - Si existe: retorna el inventario existente
   - Si no existe: crea uno nuevo
3. El sistema pre-popula los detalles con los items activos que correspondan a la frecuencia del inventario
4. Cada item recibe como `suggested_value` el `real_value` del inventario anterior del mismo tipo/schedule (Enricher)

**Reglas**:
- Unicidad: un solo inventario por (fecha, tipo, schedule)
- Los items incluidos dependen de su `inventory_frequency`: `daily` incluye items con frecuencia `daily`; `weekly` incluye `daily` + `weekly`; `monthly` incluye todos
- El inventario `initial` es especial: solo puede existir uno en el sistema

#### UC-004: Registrar Conteo Físico ✅✅
**Actor**: Empleado
**Precondición**: Inventario en estado `in_progress`
**Flujo**:
1. El empleado accede a la lista de items del inventario
2. Para cada item, ingresa el conteo físico real (`real_value`)
3. El sistema guarda el conteo y calcula si existe discrepancia

**Reglas**:
- Un item está "completo" cuando `real_value != null`
- Todos los items deben estar completos para poder completar el inventario

#### UC-005: Registrar Ventas y Compras ✅✅
**Actor**: Empleado
**Precondición**: Inventario en estado `in_progress`; aplica a inventarios que requieren ventas/compras
**Flujo**:
1. Para cada item con discrepancia (o para todos según el flujo), el empleado ingresa:
   - `stock_received`: unidades recibidas (compras)
   - `units_sold`: unidades vendidas (solo para inventarios no-opening)
2. El sistema actualiza el detalle y recalcula el valor esperado

**Reglas**:
- **Daily opening**: NO requiere ventas ni compras
- **Daily noon/closing**: SÍ requiere ventas y compras
- **Weekly/Monthly**: Solo requiere compras (las ventas las registra el POS)
- **Initial**: NO requiere ventas ni compras

#### UC-006: Ver Discrepancias ✅✅
**Actor**: Empleado
**Precondición**: Inventario en estado `in_progress`
**Flujo**:
1. El empleado solicita ver los items con discrepancia del inventario
2. El sistema calcula: `expected_at_end = suggested + stock_received - units_sold` (clamp >= 0)
3. Retorna los items donde `real_value != expected_at_end`

#### UC-007: Ver Resumen del Inventario ✅✅
**Actor**: Empleado
**Precondición**: Inventario en estado `in_progress`
**Flujo**:
1. El empleado solicita el resumen del inventario antes de completarlo
2. El sistema muestra todos los items con: sugerido, real, diferencia, tiene discrepancia
3. Muestra si el inventario puede completarse (`can_complete` = true si todos los items tienen `real_value`)

#### UC-008: Completar Inventario ✅✅
**Actor**: Empleado
**Precondición**: Todos los items del inventario tienen `real_value` registrado
**Flujo**:
1. El empleado confirma la finalización del inventario
2. El sistema verifica que todos los items estén completos
3. El sistema marca el inventario como `completed` y registra `completed_at`
4. El sistema calcula el conteo de items con discrepancia (para la respuesta)
5. Retorna el número de discrepancias encontradas

**Reglas**:
- Un inventario `initial` no genera discrepancias al completarse

#### UC-009: Ver Inventarios en Progreso ✅✅
**Actor**: Empleado
**Precondición**: Empleado autenticado
**Flujo**:
1. El empleado consulta sus inventarios que están en estado `in_progress`
2. El sistema retorna la lista de inventarios activos del empleado

### Flujo Administrativo

#### UC-010: Ver Dashboard ✅✅
**Actor**: Administrador
**Precondición**: Administrador autenticado
**Flujo**:
1. El administrador accede al dashboard
2. El sistema calcula en paralelo (3 queries):
   - Inventarios del día actual
   - Inventarios del día con discrepancias (por completar)
   - Inventarios completados hoy
3. Muestra: total inventarios hoy, con discrepancias, sin discrepancias, pendientes

#### UC-011: Crear Inventario Inicial (Baseline) ✅✅
**Actor**: Administrador
**Precondición**: No existe inventario `initial` en el sistema
**Flujo**:
1. El administrador selecciona el empleado responsable
2. El sistema crea un inventario con tipo `initial` incluyendo todos los items activos
3. Este inventario sirve como punto de partida para calcular los valores sugeridos del primer inventario del ciclo

**Reglas**:
- Solo puede existir un inventario `initial` en el sistema
- El administrador puede cargar mermas (`shrinkage`) en el inventario inicial

#### UC-012: Auditar Inventarios Históricos ✅✅
**Actor**: Administrador
**Precondición**: Administrador autenticado
**Flujo**:
1. El administrador filtra inventarios por: rango de fechas, tipo, empleado, si tienen discrepancias
2. El sistema retorna lista paginada con estadísticas por inventario
3. El administrador puede acceder al detalle de cualquier inventario

#### UC-013: Editar Detalle de Inventario (Mermas) ✅✅
**Actor**: Administrador
**Precondición**: Administrador autenticado (inventario puede estar completado)
**Flujo**:
1. El administrador accede al detalle de un inventario (incluso completado)
2. Puede editar cualquier campo del detalle: `suggested_value`, `real_value`, `stock_received`, `units_sold`, `shrinkage`
3. El campo `shrinkage` (mermas) permite registrar unidades descontadas por vencimiento o daño

**Reglas**:
- Solo el administrador puede editar mermas
- Las mermas no restan del `suggested_value` del siguiente período
- La fórmula del esperado para admin incluye mermas: `expected_for_admin = suggested - shrinkage + stock_received - units_sold` (clamp >= 0)

#### UC-014: Gestionar Items (Productos e Insumos) ✅✅
**Actor**: Administrador
**Flujo**: CRUD completo
- **Crear**: nombre único, tipo (producto/insumo), categoría, proveedor (opcional), costo (COP), unidad de medida, frecuencia de inventario (`daily`, `weekly`, `monthly`)
- **Actualizar**: mismos campos
- **Activar/Desactivar**: control de participación en inventarios
- **Especial al crear**: si hay inventarios en progreso que correspondan a la frecuencia del item, el item se agrega automáticamente a esos inventarios

#### UC-015: Gestionar Empleados ✅✅
**Actor**: Administrador
**Flujo**: CRUD completo
- **Crear**: username único, nombre, apellido, rol (employee/admin), datos opcionales (documento, teléfono, email, fecha de nacimiento)
- **Actualizar**: mismos campos
- **Activar/Desactivar**: empleados inactivos no pueden autenticarse
- **Reset Password**: restablece la contraseña al valor por defecto determinístico:
  - Si tiene `document_number` + `birth_date`: `document_number + birth_year`
  - Si solo tiene `document_number`: `document_number`
  - Fallback: `"password123"`

#### UC-016: Gestionar Categorías ✅✅
**Actor**: Administrador
**Flujo**:
- **Crear**: nombre único; `display_order` se asigna automáticamente como max + 1
- **Actualizar**: nombre, display_order
- **Reordenar**: drag-and-drop en el frontend; el admin envía el nuevo orden completo
- **Activar/Desactivar**

#### UC-017: Gestionar Proveedores ✅✅
**Actor**: Administrador
**Flujo**: CRUD completo
- **Crear**: razón social, NIT (TaxID) único, datos de contacto (nombre, teléfono, email)
- **Actualizar**: mismos campos
- **Activar/Desactivar**

#### UC-018: Ver Perfil Propio ✅✅
**Actor**: Empleado, Administrador
**Precondición**: Autenticado
**Flujo**: Retorna los datos del empleado autenticado (sin password hash)

---

## 4. Reglas de Negocio

### 4.1 Reglas de Inventario ✅✅

| ID | Regla |
|----|-------|
| RN-001 | `sugerido_siguiente = real_value` del inventario anterior del mismo tipo/schedule/item. Las mermas NO restan del sugerido del siguiente período. |
| RN-002 | `expected_at_end (empleado) = suggested + stock_received - units_sold` (clamp >= 0, sin mermas) |
| RN-003 | `expected_for_admin = suggested - shrinkage + stock_received - units_sold` (clamp >= 0, con mermas) |
| RN-004 | Discrepancia = `real_value != expected_at_end` |
| RN-005 | Para daily closing, el inventario anterior busca en orden: noon del mismo día → opening del mismo día → initial (fallback) |
| RN-006 | Un inventario `daily opening` NO requiere registrar ventas ni compras |
| RN-007 | Un inventario `daily noon/closing` SÍ requiere registrar ventas y compras |
| RN-008 | Un inventario `weekly/monthly` solo requiere compras (ventas = POS) |
| RN-009 | Unicidad de inventario: una sola combinación (fecha, tipo, schedule) por día |
| RN-010 | Para completar un inventario, TODOS los items deben tener `real_value` registrado |
| RN-011 | Solo puede existir un inventario `initial` en el sistema |
| RN-012 | El inventario `initial` no genera conteo de discrepancias al completarse |

### 4.2 Reglas de Empleados ✅✅

| ID | Regla |
|----|-------|
| RN-013 | `username` debe ser único en el sistema |
| RN-014 | Solo empleados activos pueden autenticarse |
| RN-015 | Password por defecto al resetear: `document_number + birth_year` (si tiene ambos), `document_number` (si solo tiene documento), `"password123"` (fallback) |

### 4.3 Reglas de Catálogos ✅✅

| ID | Regla |
|----|-------|
| RN-016 | Nombre de categoría debe ser único |
| RN-017 | `display_order` de categoría = max actual + 1 al crear |
| RN-018 | NIT (TaxID) de proveedor debe ser único |
| RN-019 | Nombre de item debe ser único |
| RN-020 | `cost` se almacena en COP sin decimales (INT UNSIGNED) |

---

## 5. Modelos de Datos (Resumen Funcional)

### 5.1 Empleado ✅✅
Persona del negocio con acceso al sistema. Tiene nombre, credenciales de acceso, rol (empleado/administrador) y datos personales opcionales (documento, teléfono, email, fecha de nacimiento).

### 5.2 Item ✅✅
Producto o insumo del negocio. Pertenece a una categoría, puede tener un proveedor, tiene costo, unidad de medida y frecuencia de inventario (diaria/semanal/mensual).

### 5.3 Inventario ✅✅
Sesión de conteo. Tiene fecha, tipo (daily/weekly/monthly/initial), turno (opening/noon/closing para daily), estado (en progreso/completado) y empleado responsable.

### 5.4 Detalle de Inventario ✅✅
Registro del conteo de un item específico dentro de un inventario. Contiene: valor sugerido (del período anterior), valor real (conteo físico), compras recibidas, unidades vendidas y mermas. Es el dato central del sistema.

### 5.5 Categoría ✅✅
Agrupación de items para organización y visualización. Tiene orden de display configurable.

### 5.6 Proveedor ✅✅
Empresa proveedora de items. Se identifica por su NIT (TaxID único).

### 5.7 Unidad de Medida ✅✅
Catálogo estático: Unidad, Gramos, Litros, Metros, Mililitros, Onzas.

---

## 6. Features Planificadas (NO Implementadas)

> ⚠️ Estas features están documentadas en `docs/` pero NO están implementadas en el código actual.

| Feature | Descripción | Documento |
|---------|-------------|-----------|
| **4 Turnos** | Reemplazar opening/noon/closing por morning_open/morning_close/afternoon_open/afternoon_close | `docs/INTEGRACION_TURNOS_EN_REFACTOR.md` |
| **Cierre sin conteo físico** | En el nuevo modelo de 4 turnos, los cierres solo registran ventas/compras; el `real_value` es calculado por el sistema | `docs/INTEGRACION_TURNOS_EN_REFACTOR.md` |
| **Refactor de reglas de dominio** | Centralizar `ExpectedAtEnd`, `ExpectedForAdmin`, `HasDiscrepancy`, `Enricher` en `internal/domain/inventory` | `docs/PLAN_AJUSTES_ESTRUCTURALES_SENIOR.md` |
| **Tests automatizados** | Unitarios (dominio puro) + integración (servicios con repos) | `docs/PLAN_AJUSTES_ESTRUCTURALES_SENIOR.md` |
| **Eliminar código muerto** | Remover `FindDiscrepancies`, métodos obsoletos de entidad `HasDiscrepancy()`/`Difference()` | `docs/AUTOCRITICA_DISCREPANCIAS_Y_PLAN.md` |
