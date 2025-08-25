# loopi-api

/go-backend/
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
├── migrations/             # Scripts SQL o herramientas como `golang-migrate`
├── scripts/                # Dev scripts o tooling
├── go.mod
└── go.sum

cmd/server: Punto de entrada main.go, carga de config, setup de router
internal/domain: Entidades puras (User, Shift, Franchise, etc.)
internal/usecase: Lógica de aplicación (AssignShiftService, LoginService, etc.)
internal/repository: Interfaces tipo UserRepository, ShiftRepository
internal/delivery/http: Controladores HTTP (authHandler.go, employeeHandler.go)
pkg/: Utilidades generales: logger, response JSON, etc.
config/: Archivos de configuración: .env, config.go, config.yaml
