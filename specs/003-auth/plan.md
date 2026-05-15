# Plan de Implementación: Autenticación

**Branch**: `feature/auth` (migrado) | **Fecha**: 2026-05-14 | **Spec**: `specs/003-auth/spec.md`

## Resumen

Sistema de autenticación basado en JWT para empleados de la tienda. El empleado envía username+password, el sistema verifica con bcrypt y retorna un token JWT firmado con HS256 que incluye `employee_id`, `username` y `role`. Un segundo endpoint permite al cliente obtener el perfil del usuario autenticado.

## Contexto Técnico (loopi-api — Bloqueado por Constitution)

⚠️ **El siguiente stack está bloqueado por `.specify/memory/constitution.md` — sin sustituciones**:

| Categoría | Tecnología | Versión |
|-----------|------------|---------|
| Lenguaje | Go | 1.24+ |
| HTTP Router | go-chi/chi | v5.2.4 |
| Base de datos | MySQL 8.0 via go-sql-driver | v1.9.3 |
| Auth | golang-jwt/jwt | v5.3.1 |
| Hashing | golang.org/x/crypto/bcrypt | cost=12 |
| Observabilidad | OpenTelemetry SDK | v1.43.0 |
| Logging | log/slog (stdlib) | stdlib |
| Testing | testing (stdlib) + fakes escritos a mano | stdlib |

## Verificación de Cumplimiento Constitution

*GATE: Todos deben pasar antes de escribir código.*

- [x] **Clean Architecture**: Handler en `internal/interface/handler/auth/`, service en `internal/application/service/`, infra en `internal/infrastructure/auth/`
- [x] **Repository Contracts**: `EmployeeRepository` definida en `internal/domain/repository/` antes de implementar
- [x] **slog Only**: No `fmt.Printf`, no loggers externos
- [x] **Error Handling**: `apperrors.ErrInvalidCredentials` para credenciales inválidas; `response.Respond*` en handlers
- [x] **Sin Duplicación**: `JWTManager` y `HashPassword` en `internal/infrastructure/auth/` — no duplicados
- [x] **Directory Contract**: Todos los archivos en sus rutas correctas
- [x] **GitFlow**: Implementado en rama feature desde develop

## Estructura del Proyecto

### Documentación

```text
specs/003-auth/
├── spec.md      # especificación funcional
├── plan.md      # este archivo
└── tasks.md     # tareas de implementación
```

### Archivos Implementados

```text
internal/
├── domain/
│   ├── entity/
│   │   └── employee.go                            # Role("employee"|"admin"), Employee struct
│   └── repository/
│       └── employee_repository.go                 # FindByUsername, FindByID
├── infrastructure/
│   └── auth/
│       ├── jwt.go                                 # JWTManager: GenerateToken, ValidateToken
│       └── password.go                            # HashPassword, CheckPassword (bcrypt cost=12)
├── application/
│   └── service/
│       └── auth_service.go                        # Login, GetEmployeeByID; metric loopi.auth.logins
└── interface/
    ├── middleware/
    │   └── auth_middleware.go                     # Valida JWT, inyecta employee_id en context
    ├── handler/
    │   └── auth/
    │       └── auth_handler.go                    # Login, GetMe
    └── router/
        └── router.go                              # POST /api/auth/login, GET /api/employees/me
```

### Puntos de Integración

| Componente existente | Ruta | Tipo de cambio |
|----------------------|------|----------------|
| EmployeeRepository | `internal/domain/repository/employee_repository.go` | Implementada |
| Router | `internal/interface/router/router.go` | Rutas registradas |
| Config | `pkg/config/config.go` | `JWTConfig` (Secret, ExpirationHours) |
| Secrets GCP | `app.yaml` | `JWT_SECRET` → `loopi-jwt-secret` |

## Fases de Implementación

### Fase 0: Verificación

- [x] Confirmar que `EmployeeRepository` está definida en dominio antes de la implementación
- [x] Confirmar `JWT_SECRET` gestionado por GCP Secret Manager (no en código)

### Fase 1: Dominio e Infraestructura Auth

- [x] Entidad `Employee` con `Role` (`employee` | `admin`) en `internal/domain/entity/employee.go`
- [x] Interfaz `EmployeeRepository` con `FindByUsername` y `FindByID`
- [x] `JWTManager` con `GenerateToken` (HS256, claims: `employee_id`, `username`, `role`, `exp`) y `ValidateToken`
- [x] `password.go`: `HashPassword` (bcrypt cost 12) y `CheckPassword`

### Fase 2: Middleware de Auth

- [x] `middleware.AuthMiddleware`: extrae Bearer token → `ValidateToken` → inyecta `employee_id` en context
- [x] `middleware.AdminOnly`: verifica que `role == "admin"` en el context

### Fase 3: Servicio de Aplicación

- [x] `auth_service.Login`: `FindByUsername` → `CheckPassword` → `GenerateToken` → log INFO/WARN/ERROR → metric OTel
- [x] `auth_service.GetEmployeeByID`: `FindByID` → `ErrNotFound` si nil
- [x] OTel metric `loopi.auth.logins` con atributo `result` (success / invalid_credentials / error)

### Fase 4: Handler e Interface

- [x] `auth_handler.Login`: decode + validate request → call service → map a `LoginResponse`
- [x] `auth_handler.GetMe`: extraer `employee_id` del context → call service → map a `EmployeeResponse`
- [x] Rutas registradas: `POST /api/auth/login` (público), `GET /api/employees/me` (requiere JWT)

### Fase 5: Tests

- [x] Tests de dominio: `password_test.go` — `HashPassword` + `CheckPassword`
- ⚠️ Tests pendientes: `auth_service_test.go` — Login (éxito, credenciales inválidas, error DB)

### Fase 6: Observabilidad OTel

- [x] Metric `loopi.auth.logins` — `metric.Int64Counter` con atributo `result`
- [ ] OTel span para `auth.Login` — no implementado (gap identificado)

## Complexity Tracking

| Excepción | Por qué necesaria | Alternativa más simple descartada porque |
|-----------|-------------------|------------------------------------------|
| Misma respuesta HTTP 401 para user-not-found y wrong-password | Seguridad: no revelar si el username existe | Respuestas diferenciadas exponen enumeración de usuarios |
