# Loopi API

Sistema de control de inventario para punto de venta de café de especialidad.

## Requisitos

- Go 1.21+
- MySQL 8.0+
- [golang-migrate](https://github.com/golang-migrate/migrate) (opcional, para migraciones)

## Configuración

1. Copia el archivo de ejemplo de variables de entorno:

```bash
cp .env.example .env
```

2. Edita `.env` con tus credenciales de MySQL:

```env
SERVER_PORT=8080
APP_ENV=development
DB_HOST=localhost
DB_PORT=3306
DB_USER=loopi
DB_PASSWORD=loopi_secret
JWT_SECRET=tu-clave-secreta-cambiar-en-produccion
JWT_EXPIRATION_HOURS=24
```

> `APP_ENV=development` conecta al schema `loopi_dev`. Ver [Modelo de schemas](#modelo-de-schemas).

## Base de Datos

### Modelo de schemas

La aplicación usa dos schemas MySQL separados según el entorno:

| `APP_ENV` | Schema | Uso |
|-----------|--------|-----|
| `development` | `loopi_dev` | Desarrollo local |
| `test` | `loopi_dev` | Tests locales |
| `production` | `loopi_prod` | GCP App Engine |

La variable `APP_ENV` es **obligatoria**. La aplicación falla al arrancar si no está definida o tiene un valor no reconocido.

### Configuración inicial en GCP Cloud SQL

Ejecutar como DBA en la instancia de Cloud SQL:

```bash
# 1. Crear los schemas
mysql -u <dba-user> -p < migrations/013_create_schemas.sql

# 2. Volcar datos existentes del schema legacy
mysqldump -u <user> -p loopi > loopi_backup.sql

# 3. Restaurar en loopi_dev
mysql -u <user> -p loopi_dev < loopi_backup.sql

# 4. Restaurar en loopi_prod
mysql -u <user> -p loopi_prod < loopi_backup.sql

# 5. Verificar conectividad (APP_ENV=production en app.yaml)
# 6. Eliminar el schema legacy tras verificar
mysql -u <dba-user> -p -e "DROP SCHEMA loopi;"
```

**Rollback** (si el paso 4 falla antes del deploy):
```sql
DROP SCHEMA IF EXISTS loopi_prod;
DROP SCHEMA IF EXISTS loopi_dev;
-- El schema `loopi` permanece intacto como respaldo
```

### Opción 1: Docker (desarrollo local)

```bash
docker run --name loopi-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=loopi_dev \
  -e MYSQL_USER=loopi \
  -e MYSQL_PASSWORD=loopi_secret \
  -p 3306:3306 \
  -d mysql:8.0
```

### Opción 2: MySQL Local

Crea el schema de desarrollo manualmente:

```sql
CREATE SCHEMA loopi_dev DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'loopi'@'localhost' IDENTIFIED BY 'loopi_secret';
GRANT ALL PRIVILEGES ON loopi_dev.* TO 'loopi'@'localhost';
FLUSH PRIVILEGES;
```

### Ejecutar Migraciones

```bash
# Aplicar schema inicial
mysql -u loopi -p loopi_dev < migrations/001_initial_schema.up.sql

# Insertar datos de prueba
mysql -u loopi -p loopi_dev < migrations/002_seed_data.sql
```

## Ejecución

```bash
# Cargar variables de entorno
source .env

# O exportar manualmente
export DB_HOST=localhost DB_PORT=3306 DB_USER=loopi DB_PASSWORD=loopi_secret DB_NAME=loopi JWT_SECRET=secret

# Ejecutar
make run
# o
go run cmd/api/main.go
```

## Endpoints

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/auth/login` | Login |
| GET | `/api/employees/me` | Perfil del empleado |
| GET | `/api/inventories/suggested-schedule` | Schedule sugerido |
| GET | `/api/inventories/latest` | Último inventario |
| POST | `/api/inventories` | Crear inventario |
| GET | `/api/inventories/:id/items` | Items del inventario |
| POST | `/api/inventories/:id/details` | Guardar detalle |
| GET | `/api/inventories/:id/summary` | Resumen |
| POST | `/api/inventories/:id/complete` | Completar |

## Testing con Postman

Importa la colección de Postman desde `postman/Loopi_API.postman_collection.json`.

### Credenciales de prueba

- **Usuario**: `juan`
- **Contraseña**: `password123`

## Estructura del Proyecto

```
loopi-api/
├── cmd/api/                    # Punto de entrada
├── internal/
│   ├── domain/                 # Entidades e interfaces
│   ├── application/            # Servicios y DTOs
│   ├── infrastructure/         # Repositorios, auth, DB
│   └── interface/              # Handlers y router
├── pkg/                        # Paquetes compartidos
├── migrations/                 # Scripts SQL
└── postman/                    # Colección Postman
```
