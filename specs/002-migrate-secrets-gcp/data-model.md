# Modelo de Datos: Migración de Credenciales a GCP Secret Manager

**Feature**: 002-migrate-secrets-gcp
**Fecha**: 2026-05-03

> Esta feature no introduce nuevas entidades en la base de datos. El modelo relevante
> es el de recursos de GCP Secret Manager y la configuración de App Engine.

---

## Recursos GCP Secret Manager

### Secreto: `loopi-db-password`

| Atributo | Valor |
|----------|-------|
| Nombre del recurso | `projects/quotes-api-100/secrets/loopi-db-password` |
| Política de replicación | `automatic` |
| Variable de entorno destino | `DB_PASSWORD` |
| Versión activa | `latest` (siempre apunta a la versión más reciente) |

### Secreto: `loopi-jwt-secret`

| Atributo | Valor |
|----------|-------|
| Nombre del recurso | `projects/quotes-api-100/secrets/loopi-jwt-secret` |
| Política de replicación | `automatic` |
| Variable de entorno destino | `JWT_SECRET` |
| Versión activa | `latest` |

---

## Mapeo: `app.yaml` → Variables de entorno de la instancia

### Variables que migran a `secret_env_variables`

| Variable actual en `env_variables` | Nuevo secreto en Secret Manager |
|------------------------------------|----------------------------------|
| `DB_PASSWORD`                      | `loopi-db-password`             |
| `JWT_SECRET`                       | `loopi-jwt-secret`              |

### Variables que permanecen en `env_variables` (no sensibles)

| Variable | Motivo |
|----------|--------|
| `SERVER_PORT` | Puerto de escucha — no sensible |
| `DB_INSTANCE_CONNECTION` | Identificador de instancia Cloud SQL — no sensible |
| `DB_USER` | Usuario de base de datos — no sensible |
| `APP_ENV` | Entorno de ejecución — no sensible |
| `JWT_EXPIRATION_HOURS` | Duración de tokens — no sensible |
| `TZ` | Zona horaria — no sensible |
| `LOG_LEVEL` | Nivel de logging — no sensible |
| `LOG_FORMAT` | Formato de logging — no sensible |
| `SERVICE_NAME` | Nombre del servicio — no sensible |
| `SERVICE_VERSION` | Versión del servicio — no sensible |
| `SERVICE_ENV` | Entorno del servicio — no sensible |

---

## Permisos IAM

| Principal | Rol | Alcance |
|-----------|-----|---------|
| `quotes-api-100@appspot.gserviceaccount.com` | `roles/secretmanager.secretAccessor` | Proyecto `quotes-api-100` |
