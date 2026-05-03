# Feature Specification: Migración de Credenciales a GCP Secret Manager

**Feature Branch**: `002-migrate-secrets-gcp`
**Created**: 2026-05-03
**Status**: Draft
**Input**: User description: "Migrar las credenciales que tenga hardcoded y las que no en un único lugar: Google Cloud Secret Manager y que se inyecten los valores en tiempo de deploy."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Credenciales eliminadas del repositorio (Priority: P1)

Como desarrollador, quiero que ninguna credencial sensible esté almacenada en el repositorio de código, para que una brecha en el repositorio no exponga secretos de producción.

**Why this priority**: El problema más urgente es que `app.yaml` está actualmente trackeado en git con `DB_PASSWORD` y `JWT_SECRET` en texto plano. Esto representa un riesgo de seguridad inmediato.

**Independent Test**: Validar que `app.yaml` en el repositorio no contiene valores de credenciales sensibles; solo referencias a secretos de GCP Secret Manager.

**Acceptance Scenarios**:

1. **Given** el repositorio tiene `app.yaml` trackeado, **When** se revisa el contenido del archivo, **Then** no debe contener ningún valor de contraseña, token, ni clave secreta en texto plano.
2. **Given** la aplicación está desplegada en App Engine, **When** la instancia arranca, **Then** recibe las credenciales correctas inyectadas desde Secret Manager como variables de entorno.

---

### User Story 2 - Centralización de todos los secretos en Secret Manager (Priority: P2)

Como operador, quiero que todos los secretos de la aplicación (producción y desarrollo) estén centralizados en GCP Secret Manager, para tener un único lugar de gestión, rotación y auditoría de credenciales.

**Why this priority**: Unificar la fuente de verdad para secretos simplifica la rotación, el acceso y la auditoría; es la base para una postura de seguridad sostenible.

**Independent Test**: Verificar que el proceso de deploy obtiene todos los valores sensibles exclusivamente desde Secret Manager, sin fuentes alternativas en texto plano.

**Acceptance Scenarios**:

1. **Given** los secretos `DB_PASSWORD` y `JWT_SECRET` existen en Secret Manager, **When** se ejecuta un deploy a producción, **Then** App Engine inyecta sus valores como variables de entorno sin intervención manual.
2. **Given** un nuevo miembro del equipo clona el repositorio, **When** revisa `app.yaml` y `CLAUDE.md`, **Then** encuentra instrucciones claras sobre cómo obtener y configurar los secretos para desarrollo local.

---

### User Story 3 - Desarrollo local sin credenciales en el repositorio (Priority: P3)

Como desarrollador local, quiero poder ejecutar la aplicación localmente sin necesitar credenciales hardcodeadas, usando un archivo `.env` que no se suba al repositorio.

**Why this priority**: El flujo de desarrollo local no debe degradarse; `.env` ya está gitignoreado y debe seguir siendo el mecanismo local.

**Independent Test**: Verificar que `pkg/config/config.go` sigue cargando correctamente desde variables de entorno (incluyendo `.env` via `godotenv`), sin cambios en la lógica de configuración.

**Acceptance Scenarios**:

1. **Given** el desarrollador tiene un `.env` local con las credenciales, **When** ejecuta `go run cmd/api/main.go`, **Then** la aplicación arranca y conecta a la base de datos correctamente.
2. **Given** no existe `.env` local, **When** la aplicación arranca, **Then** falla con un mensaje de error claro indicando qué variable de entorno falta.

---

### Edge Cases

- ¿Qué pasa si un secreto no existe en Secret Manager al momento del deploy? → El deploy debe fallar con error descriptivo antes de que la instancia arranque.
- ¿Qué pasa si se intenta acceder a Secret Manager sin los permisos IAM necesarios? → El deploy falla con un error de autorización auditable.
- ¿Qué sucede con secretos que existían en el historial de git? → Se documenta el proceso de rotación de credenciales comprometidas como paso obligatorio.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El archivo `app.yaml` NO DEBE contener valores de credenciales sensibles en texto plano; debe referenciar secretos de GCP Secret Manager.
- **FR-002**: Los secretos `DB_PASSWORD` y `JWT_SECRET` DEBEN estar almacenados en GCP Secret Manager como versiones con nombre.
- **FR-003**: App Engine DEBE inyectar los secretos de Secret Manager como variables de entorno usando `secret_env_variables` en `app.yaml`, de forma que `pkg/config/config.go` los lea sin modificaciones.
- **FR-004**: El archivo `.env.example` DEBE documentar qué variables son secretos gestionados por Secret Manager y cómo obtenerlos para desarrollo local.
- **FR-005**: Las credenciales actualmente expuestas en el historial de git DEBEN ser rotadas manualmente por el desarrollador: generar nuevos valores y subirlos a Secret Manager vía `gcloud secrets create` / `gcloud secrets versions add`; los valores anteriores deben considerarse comprometidos e invalidados.
- **FR-006**: Los nuevos secretos que se agreguen en el futuro DEBEN seguir el mismo patrón de Secret Manager (documentado en `CLAUDE.md`).
- **FR-007**: La cuenta de servicio de App Engine DEBE recibir el rol `roles/secretmanager.secretAccessor` en el proyecto GCP como parte del proceso de setup de esta feature, documentando el comando `gcloud` exacto en las instrucciones.

### Key Entities

- **Secret**: Un valor sensible con nombre y versión almacenado en GCP Secret Manager (ej. `loopi-db-password`, `loopi-jwt-secret`). Tiene versiones inmutables; la versión activa es la `latest`.
- **Referencia de secreto en app.yaml**: Declaración en `app.yaml` que indica a App Engine qué secreto de Secret Manager mapear a qué variable de entorno en tiempo de deploy.

## Integration Points with Existing System

**This feature interacts with the following existing modules**:

| Module | Path | Interaction Type |
|--------|------|-----------------|
| Configuración | `pkg/config/config.go` | Sin cambios — sigue leyendo desde env vars |
| Deploy config | `app.yaml` | Modificado — credenciales reemplazadas por referencias a Secret Manager |
| Documentación local | `.env.example` | Actualizado — instrucciones para desarrollo local |
| Documentación del proyecto | `CLAUDE.md` | Actualizado — convención para agregar futuros secretos |

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El archivo `app.yaml` trackeado en git no contiene ningún valor sensible — verificable con `grep -E "password|secret|token" app.yaml`.
- **SC-002**: Un deploy exitoso a App Engine con `app.yaml` actualizado confirma que la inyección desde Secret Manager funciona correctamente.
- **SC-003**: Las credenciales anteriormente expuestas en el historial de git han sido rotadas y los nuevos valores solo existen en Secret Manager.
- **SC-004**: Un desarrollador nuevo puede configurar su entorno local siguiendo únicamente las instrucciones en `.env.example`, sin acceder a valores hardcodeados en el repositorio.

## Clarifications

### Sesión 2026-05-03

- Q: ¿Cómo se integran los secretos de Secret Manager en el deploy de App Engine? → A: Opción A — referencias nativas `secret_env_variables` en `app.yaml`; App Engine inyecta el valor como variable de entorno automáticamente en deploy, sin cambios en `pkg/config/config.go`.
- Q: ¿Quién genera los nuevos valores de credenciales al rotar las credenciales comprometidas? → A: Opción A — el desarrollador genera y sube los nuevos valores manualmente vía `gcloud secrets create` / `gcloud secrets versions add`; las credenciales nunca pasan por código ni CI/CD.
- Q: ¿El permiso IAM `roles/secretmanager.secretAccessor` ya está configurado en GCP o debe configurarse en esta feature? → A: Opción B — debe configurarse como parte de esta feature; incluir el comando `gcloud` correspondiente en las instrucciones de deploy.

## Assumptions

- El proyecto GCP ya existe (`quotes-api-100`) y el servicio Secret Manager está habilitado o puede habilitarse.
- La cuenta de servicio de App Engine aún no tiene el rol `roles/secretmanager.secretAccessor`; configurarlo es parte del alcance de esta feature.
- La sintaxis de referencias a secretos en `app.yaml` para App Engine Standard es compatible con el runtime `go122`.
- `pkg/config/config.go` no requiere cambios ya que lee todas las configuraciones desde variables de entorno — Secret Manager solo cambia cómo se inyectan esas variables en producción.
- Las credenciales actuales en `app.yaml` (historial de git) serán rotadas como parte de esta feature; los valores anteriores deben considerarse comprometidos.
- El archivo `.env` local (gitignoreado) sigue siendo el mecanismo para desarrollo local; no se integra con Secret Manager en este alcance.
