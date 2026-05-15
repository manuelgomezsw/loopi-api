---
description: "Tareas de implementación para el feature de autenticación (migrado)"
---

# Tareas: Autenticación

**Feature Branch**: `feature/auth` (migrado)
**Estado**: migrated — todas las tareas completadas
**Constitution**: `.specify/memory/constitution.md`

## Convenciones de rutas (loopi-api)

| Tipo de código | Ruta |
|---------------|------|
| Handler | `internal/interface/handler/auth/auth_handler.go` |
| Service | `internal/application/service/auth_service.go` |
| Infra auth | `internal/infrastructure/auth/jwt.go`, `password.go` |
| Middleware | `internal/interface/middleware/auth_middleware.go` |
| Tests | mismo paquete que el código bajo prueba, `*_test.go` |

---

## Fase 1: Dominio e Infraestructura

- [x] T001 Entidad `Employee` con `Role` en `internal/domain/entity/employee.go`
- [x] T002 Interfaz `EmployeeRepository` con `FindByUsername` y `FindByID` en `internal/domain/repository/employee_repository.go`
- [x] T003 [P] `JWTManager.GenerateToken` y `ValidateToken` en `internal/infrastructure/auth/jwt.go`
- [x] T004 [P] `HashPassword` y `CheckPassword` (bcrypt cost=12) en `internal/infrastructure/auth/password.go`

---

## Fase 2: Middleware

- [x] T005 `middleware.AuthMiddleware` — extrae Bearer token, valida JWT, inyecta `employee_id` en `context.Context`
- [x] T006 `middleware.AdminOnly` — verifica `role == "admin"`, retorna HTTP 403 si no

---

## Fase 3: User Story 1 — Login (Prioridad: P1) 🎯 MVP

**Goal**: Empleado autentica con username+password y recibe JWT.
**Test**: `go test ./internal/application/service/... -run TestLogin`

### 🔴 TDD Fase 1: Tests fallando *(gap — pendiente)*

- [ ] T007 ⚠️ Escribir fake `EmployeeRepository` para tests de `auth_service`
- [ ] T008 ⚠️ Escribir `TestLogin_Success`, `TestLogin_InvalidPassword`, `TestLogin_UserNotFound`, `TestLogin_DBError` en `internal/application/service/auth_service_test.go`
- [ ] T009 ⚠️ Ejecutar `go test ./...` — confirmar que FALLAN

### 🟢 TDD Fase 2: Implementación

- [x] T010 [US1] `auth_service.Login`: `FindByUsername` → `CheckPassword` → `GenerateToken` → log + metric OTel
- [x] T011 **Ejecutar `go test ./...` — confirmar que PASAN**

### Interface Layer

- [x] T012 [US1] `auth_handler.Login`: decode body → `req.Validate()` → `service.Login` → `response.RespondJSON`
- [x] T013 [US1] Registrar `POST /api/auth/login` (público) en `router.go`

---

## Fase 4: User Story 2 — GetMe (Prioridad: P2)

**Goal**: Empleado autenticado obtiene su propio perfil.
**Test**: `go test ./internal/application/service/... -run TestGetEmployeeByID`

### 🟢 Implementación (sin test — gap)

- [x] T014 [US2] `auth_service.GetEmployeeByID`: `FindByID` → `ErrNotFound` si nil
- [x] T015 [US2] `auth_handler.GetMe`: extraer `employee_id` del context → call service → `response.RespondJSON`
- [x] T016 [US2] Registrar `GET /api/employees/me` (requiere JWT) en `router.go`

---

## Fase N: Verificación de Integración

- [x] TXXX Ejecutar `make test` — suite completa pasa
- [x] TXXX Ejecutar `make build` — sin errores de compilación
- [x] TXXX Smoke test manual: `POST /api/auth/login` → verificar token en respuesta → `GET /api/employees/me` con token

---

## Gaps Identificados

| Gap | Severidad | Acción recomendada |
|-----|-----------|-------------------|
| Sin tests para `auth_service.Login` | Alta — cubre flujo crítico | Crear feature `004-auth-tests` con `/speckit-specify` |
| Sin OTel span para `auth.Login` | Baja | Agregar al implementar el siguiente feature en auth |
| Sin rate limiting en `POST /api/auth/login` | Media — riesgo de brute force | Evaluar middleware de rate limiting con `/speckit-specify` |

---

## Notas

- `password_hash` NUNCA en logs ni responses — verificado en código
- Ambas rutas de fallo (user no existe, password incorrecta) retornan HTTP 401 idéntico
