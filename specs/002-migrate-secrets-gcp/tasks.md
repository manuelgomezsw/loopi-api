# Tasks: Migración de Credenciales a GCP Secret Manager

**Feature Branch**: `002-migrate-secrets-gcp`
**Input**: `specs/002-migrate-secrets-gcp/spec.md` + `specs/002-migrate-secrets-gcp/plan.md`
**Constitution**: `.specify/memory/constitution.md`

## Formato: `[ID] [P?] [US?] Descripción`

- **[P]**: Puede ejecutarse en paralelo (archivos distintos, sin dependencias incompletas)
- **[USN]**: User story a la que pertenece la tarea
- Se incluyen rutas exactas relativas a la raíz del proyecto

> ⚠️ **Esta feature no modifica código Go.** No hay entidades, repositorios ni handlers nuevos.
> Los cambios son exclusivamente en `app.yaml`, `.env.example` y `CLAUDE.md`.

---

## Fase 1: Setup y Verificación GCP

**Propósito**: Confirmar que los prerequisitos en GCP están en orden antes de modificar el repositorio.

- [x] T001 Verificar que la rama activa es `002-migrate-secrets-gcp` y fue creada desde `develop`: `git log --oneline develop..HEAD`
- [x] T002 Verificar que Secret Manager API está habilitado en el proyecto: `gcloud services list --enabled --project=quotes-api-100 --filter="name:secretmanager"`
- [x] T003 Verificar que la cuenta de servicio de App Engine tiene el rol `roles/secretmanager.secretAccessor`: `gcloud projects get-iam-policy quotes-api-100 --flatten="bindings[].members" --format="table(bindings.role,bindings.members)" --filter="bindings.members:quotes-api-100@appspot.gserviceaccount.com"`
- [x] T004 [P] Verificar que el secreto `loopi-db-password` existe y tiene al menos una versión activa: `gcloud secrets versions list loopi-db-password --project=quotes-api-100`
- [x] T005 [P] Verificar que el secreto `loopi-jwt-secret` existe y tiene al menos una versión activa: `gcloud secrets versions list loopi-jwt-secret --project=quotes-api-100`

**Checkpoint**: Todos los secretos existen en GCP y los permisos IAM están configurados. ✅

---

## Fase 2: Rotación de Credenciales Comprometidas

**Propósito**: Invalidar los valores que quedaron expuestos en el historial de git antes de proceder.

> ⚠️ **CRÍTICO**: Ejecutar esta fase ANTES de modificar `app.yaml`. Los valores actuales en git son comprometidos y no deben usarse.

- [x] T006 Conectarse a Cloud SQL (instancia `quotes-api-100:us-east1:quotes-instance`) y cambiar la contraseña del usuario `loopi-user` con un valor nuevo y fuerte (mínimo 16 caracteres): `ALTER USER 'loopi-user'@'%' IDENTIFIED BY 'NUEVA_CONTRASEÑA'; FLUSH PRIVILEGES;`
- [x] T007 Subir el nuevo valor de `DB_PASSWORD` a Secret Manager como nueva versión: `echo -n 'NUEVA_CONTRASEÑA' | gcloud secrets versions add loopi-db-password --data-file=- --project=quotes-api-100`
- [x] T008 Generar un nuevo JWT secret aleatorio de 256 bits: `openssl rand -base64 32`
- [x] T009 Subir el nuevo valor de `JWT_SECRET` a Secret Manager como nueva versión: `echo -n 'NUEVO_JWT_SECRET' | gcloud secrets versions add loopi-jwt-secret --data-file=- --project=quotes-api-100`

**Checkpoint**: Las credenciales comprometidas han sido rotadas. Los valores anteriores ya no son válidos. ✅

---

## Fase 3: US1 — Eliminar Credenciales del Repositorio 🎯 MVP

**Goal**: `app.yaml` no contiene ningún valor sensible en texto plano.
**Independent Test**: `grep -E 'DB_PASSWORD|JWT_SECRET' app.yaml` no debe retornar líneas con valores (`"..."`) — solo la clave `secret_env_variables`.

- [x] T010 [US1] Modificar `app.yaml`: eliminar `DB_PASSWORD` y `JWT_SECRET` del bloque `env_variables` y agregar el bloque `secret_env_variables` con las referencias a `loopi-db-password` y `loopi-jwt-secret` (versión `latest`)
- [x] T011 [US1] Verificar que `app.yaml` no contiene valores sensibles: `grep -iE '"[^"]*(password|secret|token|key)[^"]*"' app.yaml` — debe retornar vacío o solo claves no sensibles

**Checkpoint**: `app.yaml` puede commitearse sin exponer credenciales. ✅

---

## Fase 4: US2 — Centralización y Documentación

**Goal**: Un desarrollador nuevo entiende cómo funcionan los secretos leyendo solo `.env.example` y `CLAUDE.md`.

- [x] T012 [P] [US2] Actualizar `.env.example`: reemplazar la sección de variables sensibles con instrucciones que expliquen que `DB_PASSWORD` y `JWT_SECRET` son gestionados por GCP Secret Manager en producción, y cómo obtener valores para desarrollo local
- [x] T013 [P] [US2] Actualizar `CLAUDE.md`: agregar sección "Gestión de Secretos" con la convención para agregar futuros secretos (crear en Secret Manager + agregar `secret_env_variables` + documentar en `.env.example`), incluyendo la tabla de secretos actuales con sus nombres en Secret Manager

**Checkpoint**: La convención de secretos está documentada y es aplicable a futuros desarrollos. ✅

---

## Fase 5: US3 — Verificación de Desarrollo Local

**Goal**: La aplicación sigue arrancando correctamente con `.env` local, sin cambios en `pkg/config/config.go`.

- [x] T014 [US3] Confirmar que `pkg/config/config.go` no requiere modificaciones — la función `Load()` ya lee `DB_PASSWORD` y `JWT_SECRET` desde `os.LookupEnv`; documentar esta confirmación como comentario en el plan
- [x] T015 [US3] Verificar que el `.env` local tiene los nuevos valores rotados (del paso T007/T009) y que la aplicación arranca correctamente: `go run cmd/api/main.go` — debe mostrar `"msg":"server started"` sin errores de DB

**Checkpoint**: Desarrollo local funciona sin cambios de código. ✅

---

## Fase 6: Verificación del Deploy

**Propósito**: Confirmar que el deploy con `secret_env_variables` funciona en App Engine.

- [ ] T016 Ejecutar deploy a App Engine: `gcloud app deploy app.yaml --project=quotes-api-100`
- [ ] T017 Verificar en los logs que la aplicación arrancó correctamente y conectó a la DB: `gcloud app logs tail -s default --project=quotes-api-100` — buscar `"msg":"server started"` sin errores
- [ ] T018 Ejecutar el endpoint `/ping` en producción para confirmar que la app responde: `curl https://loopi.appspot.com/ping`

**Checkpoint**: Deploy exitoso con credenciales inyectadas desde Secret Manager. ✅

---

## Dependencias entre fases

```
Fase 1 (verificación GCP)
  └── Fase 2 (rotación de credenciales)
        └── Fase 3 (modificar app.yaml) ← MVP
              ├── Fase 4 (documentación)  ← paralela con Fase 5
              ├── Fase 5 (verificar local)
              └── Fase 6 (verificar deploy) ← depende de Fase 3
```

## Ejecución en paralelo

Dentro de la Fase 4, T012 y T013 pueden ejecutarse en paralelo (archivos distintos).
Las Fases 4 y 5 pueden ejecutarse en paralelo entre sí una vez completada la Fase 3.

## Alcance MVP

El MVP es **Fase 3 (T010–T011)** — con eso el repositorio ya no contiene credenciales y el objetivo de seguridad principal está cumplido.
Las Fases 4–5 completan la documentación y verificación local.
La Fase 6 es la validación final en producción.
