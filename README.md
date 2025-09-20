# Loopi API

Sistema de gestión de turnos, empleados y franquicias desarrollado en Go con arquitectura hexagonal.

## 🏗️ Arquitectura del Proyecto

```
/loopi-api/
├── cmd/                    # Punto de entrada principal
│   └── server/             # main.go
├── internal/               # Lógica de negocio encapsulada
│   ├── domain/             # Entidades (modelos puros)
│   ├── usecase/            # Casos de uso (servicios de aplicación)
│   ├── repository/         # Interfaces (puertos)
│   └── delivery/           # Adaptadores entrantes
│       ├── http/           # Handlers HTTP
│       └── middleware/     # Middlewares compartidos
├── pkg/                    # Código reutilizable o común
├── config/                 # Variables de entorno, config de base de datos
├── scripts/                # Scripts SQL y herramientas de desarrollo
├── postman_collections/    # Colecciones de Postman para testing
├── go.mod
└── go.sum
```

### Descripción de Directorios

- **cmd/server**: Punto de entrada main.go, carga de config, setup de router
- **internal/domain**: Entidades puras (User, Shift, Franchise, etc.)
- **internal/usecase**: Lógica de aplicación (AssignShiftService, LoginService, etc.)
- **internal/repository**: Interfaces tipo UserRepository, ShiftRepository
- **internal/delivery/http**: Controladores HTTP (authHandler.go, employeeHandler.go)
- **pkg/**: Utilidades generales: logger, response JSON, etc.
- **config/**: Archivos de configuración: .env, config.go, config.yaml

## 📬 Testing con Postman

Este proyecto incluye colecciones completas de Postman para probar todos los endpoints de la API.

### Archivos Disponibles

1. **`postman_collections.json`** - Colección completa con todos los endpoints
2. **`postman_collections_separate/`** - Colecciones separadas por módulo
3. **`README_POSTMAN.md`** - Guía detallada de uso
4. **`API_ENDPOINTS_SUMMARY.md`** - Resumen de todos los endpoints

### Importar Colecciones

1. Abrir Postman
2. Clic en "Import" → "Upload Files"
3. Seleccionar `postman_collections.json` para importar todo
4. O importar colecciones individuales desde `postman_collections_separate/`

### Configuración Rápida

Las colecciones incluyen variables predefinidas:

- `baseUrl`: http://localhost:8080
- `token`: Se llena automáticamente al hacer login
- `franchiseId`, `storeId`, `employeeId`: IDs para testing

## 🚀 Endpoints Principales

### 🔐 Authentication

- `POST /auth/login` - Login de usuario
- `POST /auth/context` - Selección de contexto

### 🏢 Business Entities

- **Franchises**: CRUD de franquicias
- **Stores**: Gestión de tiendas por franquicia
- **Employees**: Gestión completa de empleados

### ⏰ Time Management

- **Shifts**: Creación y gestión de turnos
- **Employee Hours**: Resúmenes de horas trabajadas
- **Absences**: Registro de ausencias
- **Novelties**: Gestión de novedades (horas extra, etc.)

### 📊 Analytics & Planning

- **Calendar**: Gestión de feriados y días laborables
- **Shift Planning**: Proyección de turnos y planificación

Para más detalles, consulta `API_ENDPOINTS_SUMMARY.md`.

## 🏗️ Arquitectura del Sistema

Este proyecto implementa **Clean Architecture** (Arquitectura Hexagonal) con separación clara de responsabilidades:

### 📊 Diagrama de Arquitectura

```
🌐 Cliente → 📡 Handlers → ⚙️ Use Cases → 🏢 Domain ← 🗄️ Repositories → 🗃️ Database
                    ↑              ↑                           ↑
               🛡️ Middlewares   📦 DI Container        🔧 MySQL/GORM
```

### 📚 Documentación Detallada

- **[🏗️ ARCHITECTURE.md](./ARCHITECTURE.md)** - Arquitectura completa del sistema con diagramas
- **[🚀 DEPLOYMENT.md](./DEPLOYMENT.md)** - Guías de deployment, Docker y Kubernetes
- **[📬 README_POSTMAN.md](./README_POSTMAN.md)** - Guía de testing con Postman
- **[📋 API_ENDPOINTS_SUMMARY.md](./API_ENDPOINTS_SUMMARY.md)** - Resumen de endpoints

### 🧩 Componentes Principales

- **Domain Layer**: Entidades y reglas de negocio puras
- **Use Case Layer**: Lógica de aplicación y orquestación
- **Infrastructure Layer**: Repositorios, base de datos, APIs externas
- **Presentation Layer**: Handlers HTTP, middleware, routing

### 🔐 Security Stack

- **JWT Authentication** con roles y contextos
- **Multi-tenant** con acceso por franquicia
- **Middleware Pipeline** (CORS → JWT → Roles → Franchise)
- **Input Validation** y manejo de errores centralizado

## 🔧 Desarrollo

### 🛠️ Debugging y Desarrollo en VS Code

El proyecto incluye configuraciones preconfiguradas para VS Code que automáticamente limpian procesos conflictivos:

- **"Launch Server"**: Inicia el servidor en modo debug con limpieza automática
- **"Launch Server with .env"**: Inicia el servidor cargando variables desde `.env` con limpieza automática
- **"Launch Server (Custom Port)"**: Inicia en puerto 3000 para evitar conflictos

#### Solución de Problemas de Puerto

Si encuentras el error `"bind: address already in use"`, usa cualquiera de estas opciones:

**Opción 1: Script Automático (Recomendado)**

```bash
./scripts/kill-server.sh
```

**Opción 2: Script de Limpieza Extrema (Solo casos desesperados)**

```bash
./scripts/kill-all-go.sh  # ⚠️ Mata TODOS los procesos Go del sistema
```

**Opción 3: Comandos Manuales**

```bash
# Ver qué está usando el puerto 8080
lsof -i :8080

# Matar procesos específicos
pkill -f "go run.*server"    # Procesos go run
pkill -f "__debug_b"         # Debugger de VS Code
pkill -f "server"            # Binarios del servidor

# Matar cualquier proceso en puerto 8080
lsof -ti :8080 | xargs kill -9
```

**Opción 4: Usar Configuración del IDE**
Las configuraciones de debug ya incluyen limpieza automática, simplemente ejecuta "Launch Server with .env" desde VS Code.

### Prerequisitos

- **Go 1.19+**
- **MySQL 8.0+**
- **Postman** (para testing de APIs)
- **Docker** (opcional, para deployment)

### Configuración Local

1. **Clonar el repositorio**

```bash
git clone <repository-url>
cd loopi-api
```

2. **Configurar variables de entorno**

```bash
cp .env.example .env
# Editar .env con tus configuraciones
```

3. **Instalar dependencias**

```bash
go mod download
```

4. **Configurar base de datos**

```bash
# Crear base de datos
mysql -u root -p -e "CREATE DATABASE loopi_db;"

# Ejecutar scripts de tablas (en orden)
mysql -u root -p loopi_db < scripts/tables/01.\ franchises.sql
mysql -u root -p loopi_db < scripts/tables/02.\ roles.sql
# ... (ejecutar todos los scripts en orden numérico)
```

5. **Ejecutar la aplicación**

```bash
go run cmd/server/main.go
```

La API estará disponible en `http://localhost:8080`

### 🧪 Testing

```bash
# Ejecutar tests unitarios
go test ./...

# Ejecutar tests con coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Linting
golangci-lint run
```

### 🐳 Docker Deployment

```bash
# Build imagen
docker build -t loopi-api .

# Ejecutar con Docker Compose
docker-compose up -d
```

### 📊 Monitoreo y Observabilidad

El sistema incluye endpoints para:

- **Health Check**: `GET /health`
- **Metrics**: `GET /metrics` (Prometheus)
- **Ready Check**: `GET /ready`

## 🚀 Deployment

Ver [DEPLOYMENT.md](./DEPLOYMENT.md) para guías completas de:

- **Docker & Docker Compose**
- **Kubernetes**
- **Nginx Configuration**
- **CI/CD Pipelines**
- **Monitoring Stack**
