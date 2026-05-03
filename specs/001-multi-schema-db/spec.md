# Feature Specification: Multi-Schema Database Configuration

**Feature Branch**: `feature/001-multi-schema-db`
**Created**: 2026-05-03
**Status**: Draft
**Input**: User description: "Actualmente el proyecto tiene una única base de datos que sirve de dev, test y prod. Quiero realizar una configuración diferente: tener una para dev/test y otra para prod. Al ser una aplicación en GCP con MySQL me gustaría que fueran schemas diferentes."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Aplicación usa schema correcto según entorno (Priority: P1)

Como operador del sistema, cuando la aplicación arranca en un entorno (dev, test o prod), debe conectarse automáticamente al schema de base de datos correspondiente sin intervención manual.

**Why this priority**: Es el comportamiento central de la feature — sin esto, todos los demás escenarios fallan.

**Independent Test**: Test de integración que verifica que el servicio de configuración devuelve el DSN correcto según la variable de entorno `APP_ENV`.

**Acceptance Scenarios**:

1. **Given** `APP_ENV=development`, **When** la aplicación inicializa la conexión a la base de datos, **Then** se conecta al schema `loopi_dev`
2. **Given** `APP_ENV=test`, **When** la aplicación inicializa la conexión a la base de datos, **Then** se conecta al schema `loopi_dev` (mismo schema que dev)
3. **Given** `APP_ENV=production`, **When** la aplicación inicializa la conexión a la base de datos, **Then** se conecta al schema `loopi_prod`
4. **Given** `APP_ENV` no está definida, **When** la aplicación intenta arrancar, **Then** falla con error descriptivo indicando que la variable es requerida

---

### User Story 2 - Configuración de schemas en GCP (Priority: P2)

Como desarrollador, quiero que los schemas estén documentados y sean reproducibles en GCP Cloud SQL, para que cualquier despliegue nuevo pueda configurarse correctamente.

**Why this priority**: Necesario para que el equipo pueda crear y mantener los entornos correctamente, pero no bloquea el funcionamiento de la aplicación una vez configurado.

**Independent Test**: Revisión manual de los scripts de migración/inicialización que crean cada schema.

**Acceptance Scenarios**:

1. **Given** un nuevo Cloud SQL instance, **When** se ejecutan los scripts de inicialización, **Then** se crean los schemas `loopi_dev` y `loopi_prod` correctamente
2. **Given** un schema ya existente, **When** se ejecutan los scripts de inicialización, **Then** no se sobreescribe ni pierde datos existentes

---

### Edge Cases

- ¿Qué ocurre si `APP_ENV` tiene un valor desconocido (ej: `staging`)? → La aplicación falla al arrancar con mensaje de error claro.
- ¿Qué ocurre si el schema configurado no existe en la base de datos? → La conexión falla y se reporta el error con el nombre del schema esperado.
- ¿Qué ocurre si las credenciales de conexión son incorrectas para el entorno? → Error de autenticación con log estructurado nivel ERROR.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema DEBE seleccionar el schema de base de datos basándose en la variable de entorno `APP_ENV`. Los únicos valores válidos son `development`, `test` y `production`.
- **FR-002**: Los entornos `development` y `test` DEBEN usar el mismo schema (`loopi_dev`).
- **FR-003**: El entorno `production` DEBE usar un schema separado (`loopi_prod`).
- **FR-004**: La aplicación DEBE fallar al arrancar con un error que incluya el nombre de la variable faltante o inválida y la lista de valores válidos, si `APP_ENV` no está configurada o tiene un valor no reconocido.
- **FR-005**: La configuración de conexión DEBE ser modificable mediante variables de entorno sin recompilar la aplicación. Las credenciales de BD (`DB_USER`, `DB_PASSWORD`) son compartidas entre `loopi_dev` y `loopi_prod`.
- **FR-006**: Deben existir tanto un script SQL como instrucciones en el README del proyecto para crear y poblar ambos schemas en GCP Cloud SQL.
- **FR-007**: Los schemas `loopi_dev` y `loopi_prod` DEBEN inicializarse con los mismos datos del schema `loopi` existente (dump completo de estructura y datos).
- **FR-008**: El schema `loopi` DEBE ser eliminado una vez que ambos schemas nuevos estén operativos y verificados.
- **FR-009**: Si la inicialización de `loopi_prod` falla (migraciones parciales u otro error), DEBE existir un procedimiento documentado de rollback que restaure el estado previo sin pérdida de datos.

### Key Entities

- **DatabaseConfig**: Representa la configuración de conexión a la base de datos — incluye host, puerto, usuario, contraseña y nombre del schema. El schema se resuelve dinámicamente según `APP_ENV`.
- **Environment**: Enumeración de entornos válidos (`development`, `test`, `production`) que determina qué schema usar.

## Integration Points with Existing System

**This feature interacts with the following existing modules**:

| Module | Path | Interaction Type |
|--------|------|-----------------|
| Config / DB init | `cmd/api/main.go` | Modifica inicialización de conexión para usar schema según entorno |
| App Engine config | `app.yaml` | Agrega variable `APP_ENV=production` para deploy en GCP |
| Variables de entorno | `.env` / entorno local | Requiere `APP_ENV=development` o `APP_ENV=test` en local |

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: La aplicación arranca correctamente y conecta al schema esperado en los tres entornos (dev, test, prod) sin cambios de código entre despliegues.
- **SC-002**: Todos los acceptance scenarios de User Story 1 pasan como tests automatizados.
- **SC-003**: Al ejecutar en producción (GCP App Engine), la aplicación nunca accede al schema de dev/test.
- **SC-004**: El tiempo de arranque de la aplicación no se incrementa respecto al estado actual.

## Assumptions

- GCP Cloud SQL ya está provisionado con el schema `loopi` activo; solo se necesita crear los schemas adicionales dentro de esa misma instancia.
- Las credenciales de conexión (`DB_USER`, `DB_PASSWORD`) son compartidas entre `loopi_dev` y `loopi_prod`.
- La variable `APP_ENV` se configura en `app.yaml` para producción (`production`) y en el entorno local del desarrollador para dev/test (`development`).
- Ambos schemas se inicializan con un dump completo de `loopi` (estructura + datos). A partir del cutover, cada schema evoluciona de forma independiente.
- Los nombres `loopi_dev` y `loopi_prod` están confirmados.
- El schema `loopi` se elimina solo después de verificar que ambos schemas nuevos funcionan correctamente (no se elimina de forma automática).
- No hay CI/CD configurado actualmente; los tests de integración se ejecutan solo localmente.
