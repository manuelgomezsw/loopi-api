# loopi-api — Instrucciones para Claude

## Idioma del workflow SDD

**Todo el output del workflow SDD (speckit) debe estar en español (es-CO).**

Esto incluye: spec.md, plan.md, tasks.md, research.md, data-model.md, checklists y todos los mensajes generados por comandos `/speckit-*`. El código fuente, identificadores, nombres de archivos y strings técnicos (variables, rutas, SQL) permanecen en inglés.

## Gestión de Secretos — Convención obligatoria

**Todos los valores sensibles van a GCP Secret Manager.** Nunca en `app.yaml`, código fuente ni archivos trackeados por git.

### Campos considerados secretos
- Contraseñas de base de datos
- Claves privadas JWT / API keys
- Tokens de integración con servicios externos

### Secretos actuales

| Variable de entorno | Secreto en Secret Manager | Proyecto GCP |
|---------------------|---------------------------|--------------|
| `DB_PASSWORD` | `loopi-db-password` | `quotes-api-100` |
| `JWT_SECRET` | `loopi-jwt-secret` | `quotes-api-100` |

### Para agregar un nuevo secreto

1. Crear el secreto en Secret Manager:
   ```bash
   echo -n 'VALOR' | gcloud secrets create loopi-<nombre> \
     --data-file=- --replication-policy="automatic" --project=quotes-api-100
   ```
2. Agregar referencia en `app.yaml` bajo `secret_env_variables`:
   ```yaml
   secret_env_variables:
     - name: NOMBRE_VAR_ENV
       secret: loopi-<nombre>
       version: latest
   ```
3. Leer el valor en `pkg/config/config.go` con `getEnv("NOMBRE_VAR_ENV", "")`.
4. Documentar en `.env.example` con instrucciones para desarrollo local.

---

## Stack y arquitectura

- **Lenguaje**: Go 1.24, módulo `github.com/manuelgomezsw/loopi-api`
- **Framework HTTP**: go-chi/chi v5
- **Arquitectura**: Clean Architecture — `interface/ → application/ → domain/ → infrastructure/`
- **Deploy**: Google App Engine Standard, Cloud SQL (MySQL 8.0)
- **Entry point activo**: `cmd/api/main.go`

---

## GitFlow — Regla obligatoria para toda nueva feature

Antes de escribir cualquier línea de código en una nueva feature, seguir estos pasos:

### 1. Crear rama desde `develop`
```bash
git checkout develop
git pull origin develop
git checkout -b feature/<nombre-descriptivo>
# Ejemplos: feature/admin-reports, feature/password-policy
```

### 2. Commits durante el desarrollo
- Commits atómicos al completar cada subtarea del plan
- Formato obligatorio: `[ADD]`, `[CHANGE]`, `[FIX]`, `[REMOVE]` + descripción en inglés
- Ejemplos:
  - `[ADD] pkg/logger: structured slog logger with GCP handler`
  - `[CHANGE] config: add LogConfig with LOG_LEVEL and LOG_FORMAT`
  - `[FIX] auth: handle nil employee before password check`

### 3. Finalizar feature
```bash
git checkout develop
git merge --no-ff feature/<nombre>
# O bien: PR de feature → develop
```

### Reglas para Claude
- NUNCA commitear directamente en `master` o `develop`
- SIEMPRE crear rama `feature/` antes de implementar
- NUNCA mezclar múltiples features en una sola rama
- Un plan de Claude = una rama feature

### Claude en worktrees (Claude Code)
Cuando Claude Code abre un worktree, la rama se llama `claude/xxx` automáticamente.
**Esa rama NO es válida para desarrollo.** Antes del primer commit, ejecutar obligatoriamente:

```bash
git checkout develop
git pull origin develop
git checkout -b feature/<nombre-descriptivo>
```

El PR **siempre** debe apuntar a `develop`, nunca a `master`.

---

## Logging — Reglas obligatorias

### Librería
- Usar SIEMPRE `log/slog` de stdlib. **No introducir** zap, logrus, zerolog ni otros.
- El logger global se configura en `cmd/api/main.go` via `slog.SetDefault()`.
- El paquete de logging está en `pkg/logger/`.

### Context Pattern
```go
// En servicios: obtener logger del contexto
log := logger.FromContext(ctx)
log.InfoContext(ctx, "mensaje", "key", value)

// Nunca:
log.Printf(...)         // stdlib sin estructura
fmt.Printf(...)        // print de debug
slog.Info(...)         // sin contexto (pierde request_id)
```

### Niveles
| Nivel | Cuándo |
|-------|--------|
| `DEBUG` | Queries SQL, estado interno — solo dev |
| `INFO` | Eventos de negocio exitosos (login, crear entidad, completar inventario) |
| `WARN` | Anomalías recuperables: ErrNotFound, ErrInvalidCredentials, regla de negocio |
| `ERROR` | Fallos que requieren atención: errores de DB, errores de infraestructura |

### Campos obligatorios en cada log de servicio
```go
log.InfoContext(ctx, "descripción del evento",
    "operation", "domain.Method",  // OBLIGATORIO: formato "servicio.Método"
    "entity_id",  entityID,        // cuando aplique
    "error",      err,             // solo en WARN/ERROR
)
```

El campo `request_id` **no se agrega manualmente** — el middleware `RequestLogger` ya lo inyecta en el contexto.

### Dónde loguear
- ✅ **Servicios**: eventos de negocio y errores de infraestructura
- ✅ **`cmd/api/main.go`**: startup, errores fatales
- ❌ **Handlers**: NO — los servicios ya capturan los errores
- ❌ **Repositorios**: solo DEBUG para queries lentas (futuro)

### Al agregar un nuevo servicio o método
1. Importar `"github.com/manuelgomezsw/loopi-api/pkg/logger"`
2. En operaciones de escritura (create/update/delete/complete): log INFO al éxito
3. En errores de DB/infra: log ERROR antes del return
4. En errores de dominio (not found, conflict, invalid): log WARN antes del return

### Seguridad — campos PROHIBIDOS en logs
```
password, password_hash, token, jwt, secret, authorization
```
Nunca loguear cuerpos completos de request/response ni PII sin enmascarar.

---

## Actualizar technical-spec.md

Siempre que se agregue/cambie:
- Una nueva dependencia → actualizar sección 1.2 (Stack Tecnológico)
- Un nuevo middleware → actualizar sección 2.6
- Una nueva variable de entorno → actualizar sección 5.1
- Una nueva capacidad de observabilidad → actualizar sección 9

---

## Convenciones de código

- Repositorios: solo datos (SELECT/INSERT/UPDATE/DELETE) — CERO lógica de negocio
- Toda lógica en servicios (`application/service/`) y dominio (`domain/`)
- Errores de dominio: usar `pkg/errors` (`apperrors.ErrNotFound`, `apperrors.New(code, msg)`)
- Respuestas HTTP: usar `internal/interface/response` (`RespondJSON`, `RespondError`, `RespondSuccess`)
- No agregar dependencias externas sin evaluar alternativas stdlib primero

<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan at
`specs/002-migrate-secrets-gcp/plan.md`
<!-- SPECKIT END -->
