# 📋 Resumen de Endpoints - Loopi API

## 🔐 Authentication

| Método | Endpoint        | Descripción              | Auth | Rol |
| ------ | --------------- | ------------------------ | ---- | --- |
| `POST` | `/auth/login`   | Autenticación de usuario | ❌   | -   |
| `POST` | `/auth/context` | Selección de contexto    | ✅   | -   |

## 🏢 Franchises

| Método | Endpoint           | Descripción                   | Auth | Rol   |
| ------ | ------------------ | ----------------------------- | ---- | ----- |
| `GET`  | `/franchises/`     | Obtener todas las franquicias | ✅   | -     |
| `GET`  | `/franchises/{id}` | Obtener franquicia por ID     | ✅   | admin |
| `POST` | `/franchises/`     | Crear franquicia              | ✅   | admin |

## 🏪 Stores

| Método   | Endpoint                                 | Descripción                     | Auth | Rol   |
| -------- | ---------------------------------------- | ------------------------------- | ---- | ----- |
| `GET`    | `/stores/`                               | Obtener todas las tiendas       | ✅   | admin |
| `GET`    | `/stores/{id}`                           | Obtener tienda por ID           | ✅   | admin |
| `POST`   | `/stores/`                               | Crear tienda                    | ✅   | admin |
| `PUT`    | `/stores/{id}`                           | Actualizar tienda               | ✅   | admin |
| `DELETE` | `/stores/{id}`                           | Eliminar tienda                 | ✅   | admin |
| `GET`    | `/stores/franchise/{franchiseID}`        | Tiendas por franquicia          | ✅   | -     |
| `GET`    | `/stores/franchise/{franchiseID}/active` | Tiendas activas por franquicia  | ✅   | -     |
| `GET`    | `/stores/with-employee-count`            | Tiendas con conteo de empleados | ✅   | admin |
| `GET`    | `/stores/statistics`                     | Estadísticas de tiendas         | ✅   | admin |

## 👥 Employees

| Método   | Endpoint                             | Descripción                  | Auth | Rol   |
| -------- | ------------------------------------ | ---------------------------- | ---- | ----- |
| `GET`    | `/employees/`                        | Obtener todos los empleados  | ✅   | admin |
| `GET`    | `/employees/{id}`                    | Obtener empleado por ID      | ✅   | admin |
| `POST`   | `/employees/`                        | Crear empleado               | ✅   | admin |
| `PUT`    | `/employees/{id}`                    | Actualizar empleado          | ✅   | admin |
| `DELETE` | `/employees/{id}`                    | Eliminar empleado            | ✅   | admin |
| `GET`    | `/employees/active`                  | Empleados activos            | ✅   | admin |
| `GET`    | `/employees/store/{store_id}`        | Empleados por tienda         | ✅   | admin |
| `GET`    | `/employees/store/{store_id}/active` | Empleados activos por tienda | ✅   | admin |

## ⏰ Employee Hours

| Método | Endpoint                            | Descripción              | Auth | Rol   |
| ------ | ----------------------------------- | ------------------------ | ---- | ----- |
| `GET`  | `/employee-hours/{id}/monthly`      | Resumen mensual de horas | ✅   | admin |
| `GET`  | `/employee-hours/{id}/daily`        | Resumen diario de horas  | ✅   | admin |
| `GET`  | `/employee-hours/{id}/yearly`       | Resumen anual de horas   | ✅   | admin |
| `GET`  | `/employee-hours/{id}/working-days` | Días trabajados          | ✅   | admin |
| `GET`  | `/employee-hours/{id}`              | Resumen mensual (legacy) | ✅   | admin |

## 📅 Calendar

| Método | Endpoint                     | Descripción                     | Auth | Rol   |
| ------ | ---------------------------- | ------------------------------- | ---- | ----- |
| `GET`  | `/calendar/holidays`         | Obtener feriados                | ✅   | admin |
| `GET`  | `/calendar/month-summary`    | Resumen mensual del calendario  | ✅   | admin |
| `GET`  | `/calendar/enhanced-summary` | Resumen mejorado del calendario | ✅   | admin |
| `GET`  | `/calendar/working-days`     | Días laborables                 | ✅   | admin |
| `POST` | `/calendar/clear-cache`      | Limpiar caché del calendario    | ✅   | admin |

## 🔄 Shifts

| Método | Endpoint                              | Descripción                       | Auth | Rol   |
| ------ | ------------------------------------- | --------------------------------- | ---- | ----- |
| `POST` | `/shifts/`                            | Crear turno                       | ✅   | admin |
| `GET`  | `/shifts/`                            | Obtener todos los turnos          | ✅   | admin |
| `GET`  | `/shifts/single`                      | Obtener turno específico          | ✅   | admin |
| `GET`  | `/shifts/store/{store_id}`            | Turnos por tienda                 | ✅   | admin |
| `GET`  | `/shifts/store/{store_id}/statistics` | Estadísticas de turnos por tienda | ✅   | admin |
| `GET`  | `/shifts/period`                      | Turnos por período                | ✅   | admin |

## 📊 Shift Planning

| Método | Endpoint                         | Descripción                | Auth | Rol   |
| ------ | -------------------------------- | -------------------------- | ---- | ----- |
| `POST` | `/shift-planning/preview`        | Vista previa de proyección | ✅   | admin |
| `GET`  | `/shift-planning/summary`        | Resumen de proyección      | ✅   | admin |
| `GET`  | `/shift-planning/projected-days` | Días proyectados           | ✅   | admin |

## 🚫 Absences

| Método | Endpoint                | Descripción                           | Auth | Rol   |
| ------ | ----------------------- | ------------------------------------- | ---- | ----- |
| `POST` | `/absences/`            | Crear ausencia                        | ✅   | admin |
| `GET`  | `/absences/monthly`     | Ausencias por empleado y mes          | ✅   | admin |
| `GET`  | `/absences/date-range`  | Ausencias por rango de fechas         | ✅   | admin |
| `GET`  | `/absences/total-hours` | Total de horas de ausencia            | ✅   | admin |
| `GET`  | `/absences/`            | Ausencias por empleado y mes (legacy) | ✅   | admin |

## 🎯 Novelties

| Método | Endpoint                         | Descripción                           | Auth | Rol   |
| ------ | -------------------------------- | ------------------------------------- | ---- | ----- |
| `POST` | `/novelties/`                    | Crear novedad                         | ✅   | admin |
| `GET`  | `/novelties/monthly`             | Novedades por empleado y mes          | ✅   | admin |
| `GET`  | `/novelties/date-range`          | Novedades por rango de fechas         | ✅   | admin |
| `GET`  | `/novelties/total-hours-by-type` | Total de horas por tipo               | ✅   | admin |
| `GET`  | `/novelties/types-summary`       | Resumen de tipos de novedades         | ✅   | admin |
| `GET`  | `/novelties/`                    | Novedades por empleado y mes (legacy) | ✅   | admin |

## 📝 Parámetros Comunes

### Query Parameters Frecuentes

- `year`: Año (ej: `2025`)
- `month`: Mes (ej: `1` para enero)
- `day`: Día (ej: `15`)
- `employee`: ID del empleado
- `store`: ID de la tienda
- `franchise`: ID de la franquicia
- `active`: Filtro por activos (`true`/`false`)
- `from`: Fecha de inicio (`YYYY-MM-DD`)
- `to`: Fecha de fin (`YYYY-MM-DD`)
- `type`: Tipo de novedad

### Path Parameters

- `{id}`: ID del recurso
- `{franchiseID}`: ID de la franquicia
- `{store_id}`: ID de la tienda
- `{shift_id}`: ID del turno

## 🔒 Middleware y Seguridad

### JWT Middleware

- **Aplica a**: Todos los endpoints excepto `/auth/login`
- **Header**: `Authorization: Bearer {token}`

### Role Middleware

- **admin**: Requerido para la mayoría de operaciones CRUD
- **employee**: Acceso limitado a sus propios datos

### Franchise Access Middleware

- **Aplica a**: Endpoints de empleados y horas
- **Validación**: Acceso solo a datos de la franquicia del usuario

## 📊 Códigos de Estado HTTP

| Código | Descripción                                |
| ------ | ------------------------------------------ |
| `200`  | OK - Operación exitosa                     |
| `201`  | Created - Recurso creado                   |
| `204`  | No Content - Eliminación exitosa           |
| `400`  | Bad Request - Error en los datos           |
| `401`  | Unauthorized - Token inválido/faltante     |
| `403`  | Forbidden - Sin permisos                   |
| `404`  | Not Found - Recurso no encontrado          |
| `500`  | Internal Server Error - Error del servidor |

## 🏗️ Estructura de Responses

### Success Response

```json
{
    "data": {...},
    "status": "success"
}
```

### Error Response

```json
{
  "error": "Mensaje de error",
  "status": "error"
}
```

### Login Response

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```
