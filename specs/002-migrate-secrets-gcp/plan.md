# Plan de Implementación: Migración de Credenciales a GCP Secret Manager

**Rama**: `002-migrate-secrets-gcp` | **Fecha**: 2026-05-03 | **Spec**: `specs/002-migrate-secrets-gcp/spec.md`

## Resumen

Migrar las credenciales sensibles (`DB_PASSWORD`, `JWT_SECRET`) que actualmente están en texto plano en `app.yaml` (trackeado en git) hacia GCP Secret Manager, inyectándolas en tiempo de deploy mediante `secret_env_variables`. No se modifica el código Go; `pkg/config/config.go` sigue leyendo desde variables de entorno sin cambios.

## Contexto Técnico (loopi-api — Bloqueado por Constitución)

⚠️ **El siguiente stack está bloqueado por `.specify/memory/constitution.md` — sin sustituciones**:

| Categoría | Tecnología | Versión |
|-----------|------------|---------|
| Lenguaje | Go | 1.24+ |
| HTTP Router | go-chi/chi | v5.2.4 |
| Base de datos | MySQL 8.0 via go-sql-driver | v1.9.3 |
| Auth | golang-jwt/jwt | v5.3.1 |
| Observabilidad | OpenTelemetry SDK | v1.43.0 |
| Logging | log/slog (stdlib) | stdlib |
| Testing | testing (stdlib) + fakes escritos a mano | stdlib |
| Build | `make build` / `go build` | — |
| Deploy | Google App Engine Standard | — |
| Secrets | GCP Secret Manager | — |

## Constitution Compliance Check

*GATE: Todos deben pasar antes de implementar.*

- [x] **Clean Architecture**: No aplica — no se crean handlers, services ni repositories nuevos
- [x] **Repository Contracts**: No aplica — no se modifica ningún repositorio
- [x] **slog Only**: No aplica — no se toca código de logging
- [x] **Error Handling**: No aplica — no se modifica ningún handler
- [x] **No Duplication**: `pkg/config/config.go` no cambia; se reutiliza el mecanismo existente de env vars
- [x] **Directory Contract**: Los cambios son en `app.yaml`, `.env.example`, `CLAUDE.md` — fuera del directorio Go (`internal/`)
- [x] **GitFlow**: Rama `002-migrate-secrets-gcp` creada desde `develop`

## Estructura del Proyecto

### Documentación (esta feature)

```text
specs/002-migrate-secrets-gcp/
├── spec.md         # especificación funcional
├── plan.md         # este archivo
├── research.md     # investigación técnica
├── data-model.md   # mapeo de secretos y permisos
└── tasks.md        # generado por /speckit-tasks
```

### Archivos Nuevos / Modificados

```text
app.yaml                          # MODIFICADO: secrets → secret_env_variables
.env.example                      # MODIFICADO: instrucciones de desarrollo local
CLAUDE.md                         # MODIFICADO: convención para nuevos secretos
```

> **No se crea ni modifica ningún archivo Go.**

### Puntos de Integración con el Código Existente

| Componente Existente | Ubicación | Tipo de Cambio |
|----------------------|-----------|----------------|
| Deploy config | `app.yaml` | Reemplazar `env_variables` de credenciales con `secret_env_variables` |
| Config cargador | `pkg/config/config.go` | Sin cambios — sigue usando `os.LookupEnv` |
| Documentación local | `.env.example` | Actualizar con instrucciones de secretos |
| Instrucciones Claude | `CLAUDE.md` | Agregar convención de secretos |

---

## Fases de Implementación

### Fase 0: Setup GCP (prerequisito — fuera del repositorio)

Estos pasos se ejecutan una sola vez en GCP antes de modificar cualquier archivo del repositorio.

**0.1 Habilitar Secret Manager API**
```bash
gcloud services enable secretmanager.googleapis.com --project=quotes-api-100
```

**0.2 Configurar permisos IAM para App Engine**
```bash
gcloud projects add-iam-policy-binding quotes-api-100 \
  --member="serviceAccount:quotes-api-100@appspot.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

**0.3 Generar nuevas credenciales (rotar las comprometidas)**

Las credenciales actuales en el historial de git deben considerarse comprometidas. Generar nuevos valores:

```bash
# Nuevo JWT Secret (256 bits aleatorios)
openssl rand -base64 32
# Guardar el output — se usará en el paso 0.5
```

Para `DB_PASSWORD`: generar una contraseña fuerte (mínimo 16 caracteres, letras + números + símbolos).

**0.4 Actualizar contraseña en Cloud SQL**
```bash
# Conectarse a la instancia Cloud SQL y ejecutar:
# ALTER USER 'loopi-user'@'%' IDENTIFIED BY 'NUEVA_CONTRASEÑA_FUERTE';
# FLUSH PRIVILEGES;
```

**0.5 Crear secretos en Secret Manager**
```bash
# Crear el secreto DB_PASSWORD
echo -n 'NUEVA_CONTRASEÑA_FUERTE' | gcloud secrets create loopi-db-password \
  --data-file=- \
  --replication-policy="automatic" \
  --project=quotes-api-100

# Crear el secreto JWT_SECRET
echo -n 'NUEVO_JWT_SECRET_BASE64' | gcloud secrets create loopi-jwt-secret \
  --data-file=- \
  --replication-policy="automatic" \
  --project=quotes-api-100
```

> **Verificación**: `gcloud secrets list --project=quotes-api-100` debe mostrar ambos secretos.

---

### Fase 1: Modificar `app.yaml`

Reemplazar las entradas sensibles de `env_variables` por referencias a `secret_env_variables`.

**Resultado esperado de `app.yaml`**:
```yaml
service: loopi
runtime: go122

instance_class: F1

automatic_scaling:
  min_instances: 0
  max_instances: 2
  target_cpu_utilization: 0.65

env_variables:
  SERVER_PORT: "8080"
  DB_INSTANCE_CONNECTION: "quotes-api-100:us-east1:quotes-instance"
  DB_USER: "loopi-user"
  APP_ENV: "production"
  JWT_EXPIRATION_HOURS: "24"
  TZ: "America/Bogota"
  LOG_LEVEL: "info"
  LOG_FORMAT: "json"
  SERVICE_NAME: "loopi-api"
  SERVICE_VERSION: "1.0.0"
  SERVICE_ENV: "production"

secret_env_variables:
  - name: DB_PASSWORD
    secret: loopi-db-password
    version: latest
  - name: JWT_SECRET
    secret: loopi-jwt-secret
    version: latest

handlers:
  - url: /.*
    script: auto
    secure: always
```

> **Verificación**: `grep -E '"[^"]*"' app.yaml | grep -iE "password|secret|token"` no debe retornar resultados con valores sensibles.

---

### Fase 2: Actualizar `.env.example`

Agregar instrucciones claras sobre qué variables son gestionadas por Secret Manager en producción y cómo obtenerlas para desarrollo local.

**Resultado esperado de `.env.example`** (sección de secretos):
```bash
# =============================================================================
# SECRETOS — Gestionados por GCP Secret Manager en producción
# Para desarrollo local: obtener los valores con el equipo y colocarlos aquí.
# NUNCA commitear .env con valores reales.
# =============================================================================

# Contraseña de la base de datos MySQL
# Producción: inyectada automáticamente desde Secret Manager (loopi-db-password)
DB_PASSWORD=

# Clave secreta para firmar JWT
# Producción: inyectada automáticamente desde Secret Manager (loopi-jwt-secret)
JWT_SECRET=
```

---

### Fase 3: Actualizar `CLAUDE.md`

Agregar una sección de convención para futuros secretos, de forma que cualquier nueva credencial siga el mismo patrón.

**Sección a agregar en `CLAUDE.md`** (bajo la sección de Stack/arquitectura):

```markdown
## Gestión de Secretos — Convención obligatoria

### Regla: todos los secretos van a GCP Secret Manager

Nunca agregar valores sensibles en `app.yaml`, código fuente, ni en ningún archivo trackeado por git.

**Campos que se consideran secretos:**
- Contraseñas de base de datos
- Claves privadas JWT / API keys
- Tokens de integración con servicios externos

**Para agregar un nuevo secreto:**

1. Crear el secreto en Secret Manager:
   ```bash
   echo -n 'VALOR_SECRETO' | gcloud secrets create loopi-<nombre-descriptivo> \
     --data-file=- --replication-policy="automatic" --project=quotes-api-100
   ```

2. Agregar referencia en `app.yaml` bajo `secret_env_variables`:
   ```yaml
   secret_env_variables:
     - name: NOMBRE_VAR_ENV
       secret: loopi-<nombre-descriptivo>
       version: latest
   ```

3. Agregar la variable en `pkg/config/config.go` con `getEnv("NOMBRE_VAR_ENV", "")`.

4. Documentar en `.env.example` con instrucciones para desarrollo local.

**Secretos actuales:**
| Variable de entorno | Secreto en Secret Manager     |
|---------------------|-------------------------------|
| `DB_PASSWORD`       | `loopi-db-password`           |
| `JWT_SECRET`        | `loopi-jwt-secret`            |
```

---

### Fase 4: Verificación del deploy

**4.1 Deploy de prueba**
```bash
gcloud app deploy app.yaml --project=quotes-api-100
```

**4.2 Verificar que la aplicación arranca correctamente**
```bash
gcloud app logs tail -s default --project=quotes-api-100
```
Buscar: `"msg":"server started"` sin errores de conexión a DB ni de configuración.

**4.3 Verificar que `app.yaml` no contiene secretos**
```bash
grep -iE "password|secret|token" app.yaml
```
Solo debe aparecer `secret_env_variables` (la clave del bloque), no valores sensibles.

---

## Complejidad y Excepciones

> Esta feature no tiene excepciones a la constitución. No se modifica código Go.

| Excepción | Motivo | Alternativa más simple descartada por |
|-----------|--------|---------------------------------------|
| — | — | — |
