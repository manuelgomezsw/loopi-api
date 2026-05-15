# Especificación Funcional: Autenticación

**Feature Branch**: `feature/auth` (migrado desde `master`)
**Estado**: migrated
**Fecha de migración**: 2026-05-14

## User Scenarios & Testing

### User Story 1 — Login de empleado (Prioridad: P1)

Un empleado ingresa su username y contraseña para obtener un token JWT que le permite usar la API.

**Por qué P1**: Sin autenticación no existe ninguna otra funcionalidad de la API.

**Test independiente**: `AuthService.Login` probado con un fake de `EmployeeRepository` que retorna un empleado con hash conocido.

**Escenarios de aceptación**:

1. **Dado** un empleado activo con username `"juan"` y contraseña correcta, **Cuando** `POST /api/auth/login`, **Entonces** HTTP 200 con `token` (JWT firmado) y datos del empleado.
2. **Dado** una contraseña incorrecta, **Cuando** `POST /api/auth/login`, **Entonces** HTTP 401 `"invalid credentials"`.
3. **Dado** un username inexistente, **Cuando** `POST /api/auth/login`, **Entonces** HTTP 401 `"invalid credentials"` (misma respuesta — no revela si el usuario existe).
4. **Dado** un body con `username` vacío, **Cuando** `POST /api/auth/login`, **Entonces** HTTP 400 con error de validación.
5. **Dado** un error de base de datos en `FindByUsername`, **Cuando** `POST /api/auth/login`, **Entonces** HTTP 500 y log ERROR en el servicio.

---

### User Story 2 — Obtener perfil propio (Prioridad: P2)

Un empleado autenticado consulta sus propios datos de perfil.

**Por qué P2**: Necesario para que el cliente conozca el rol y nombre del usuario logueado.

**Test independiente**: `AuthService.GetEmployeeByID` con fake repo retornando empleado por ID.

**Escenarios de aceptación**:

1. **Dado** un token JWT válido en `Authorization: Bearer <token>`, **Cuando** `GET /api/employees/me`, **Entonces** HTTP 200 con `id`, `username`, `name`, `last_name`, `full_name`, `role`.
2. **Dado** un token expirado o inválido, **Cuando** `GET /api/employees/me`, **Entonces** HTTP 401.
3. **Dado** token válido pero empleado eliminado de la BD, **Cuando** `GET /api/employees/me`, **Entonces** HTTP 404.

---

### Edge Cases

- Token con `employee_id` inexistente → `apperrors.ErrNotFound` → HTTP 404.
- Token con firma inválida (clave JWT modificada) → `middleware.AuthMiddleware` rechaza → HTTP 401.
- Body de login con JSON malformado → HTTP 400.
- Contraseña vacía en el request → validación `req.Validate()` → HTTP 400.

## Requisitos Funcionales

- **FR-001**: El sistema DEBE validar `username` y `password` no vacíos antes de consultar la base de datos.
- **FR-002**: El sistema DEBE comparar la contraseña usando bcrypt con costo 12.
- **FR-003**: El sistema DEBE generar un token JWT firmado con HS256, incluyendo `employee_id`, `username` y `role` en los claims.
- **FR-004**: El JWT DEBE expirar en `JWT_EXPIRATION_HOURS` horas (configurable vía env, default 24h).
- **FR-005**: El sistema DEBE retornar la misma respuesta HTTP 401 tanto para credenciales incorrectas como para usuario inexistente (no revelar si el usuario existe).
- **FR-006**: El endpoint `GET /api/employees/me` DEBE estar protegido por `middleware.AuthMiddleware`.

### Entidades clave

- **Employee**: `id`, `username`, `password_hash` (nunca expuesto en responses), `name`, `last_name`, `role` (`employee` | `admin`), `active`.

## Puntos de Integración

| Módulo | Ruta | Tipo de interacción |
|--------|------|---------------------|
| EmployeeRepository | `internal/domain/repository/employee_repository.go` | `FindByUsername`, `FindByID` |
| JWTManager | `internal/infrastructure/auth/jwt.go` | `GenerateToken`, `ValidateToken` |
| AuthMiddleware | `internal/interface/middleware/auth_middleware.go` | Inyecta `employee_id` en context |
| OTel metric | `pkg/observability/` | Counter `loopi.auth.logins` con atributo `result` |

## Contrato HTTP

| Método | Endpoint | Auth | Request Body | Response |
|--------|----------|------|--------------|----------|
| POST | `/api/auth/login` | No | `{"username":"","password":""}` | `{"token":"","employee":{...}}` |
| GET | `/api/employees/me` | Bearer JWT | — | `{"id":1,"username":"","name":"","last_name":"","full_name":"","role":""}` |

## Criterios de Éxito

- **SC-001**: Login exitoso retorna JWT válido que el middleware acepta en llamadas subsecuentes.
- **SC-002**: Credenciales inválidas nunca exponen si el username existe.
- **SC-003**: El metric OTel `loopi.auth.logins` se incrementa en cada intento (éxito, credenciales inválidas y error).
- **SC-004**: El campo `password_hash` nunca aparece en ninguna respuesta HTTP ni en logs.

## Assumptions

- Los empleados son creados únicamente por el admin — no hay registro público.
- La expiración del token es configurable vía `JWT_EXPIRATION_HOURS` (env / Secret Manager en producción).
- No existe refresh token — el cliente debe re-autenticarse al expirar.
