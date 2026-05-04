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
DB_HOST=localhost
DB_PORT=3306
DB_USER=loopi
DB_PASSWORD=loopi_secret
DB_NAME=loopi
JWT_SECRET=tu-clave-secreta-cambiar-en-produccion
JWT_EXPIRATION_HOURS=24
```

## Base de Datos

### Opción 1: Docker (recomendado)

```bash
docker run --name loopi-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=loopi \
  -e MYSQL_USER=loopi \
  -e MYSQL_PASSWORD=loopi_secret \
  -p 3306:3306 \
  -d mysql:8.0
```

### Opción 2: MySQL Local

Crea la base de datos manualmente:

```sql
CREATE DATABASE loopi CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE USER 'loopi'@'localhost' IDENTIFIED BY 'loopi_secret';
GRANT ALL PRIVILEGES ON loopi.* TO 'loopi'@'localhost';
FLUSH PRIVILEGES;
```

### Ejecutar Migraciones

```bash
# Aplicar schema inicial
mysql -u loopi -p loopi < migrations/001_initial_schema.up.sql

# Insertar datos de prueba
mysql -u loopi -p loopi < migrations/002_seed_data.sql
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
