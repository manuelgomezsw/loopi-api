# Metadatos de Extracción

**Fecha**: 2026-03-26
**Modo**: FULL EXTRACTION
**Aplicación**: Loopi
**Estrategia**: ASSISTED (documentación rica + código disponible, no-Fury)

## Fuentes

| Fuente | Estado | Notas |
|--------|--------|-------|
| **Código Go (loopi-api)** | ✅ Completo | `/Users/mangomez/Repos/Manu/loopi/loopi-api/` |
| **Código Angular (loopi-web)** | ✅ Completo | `/Users/mangomez/Repos/Manu/loopi/loopi-web/` |
| **ARCHITECTURE.md** | ✅ Completo | Principios y dominios |
| **docs/ (8 documentos)** | ✅ Completo | Análisis de bugs, planes de refactor |
| **FuryMCP** | ⛔ N/A | No es un servicio Fury/MeLi |
| **MeliSystemMCP** | ⛔ N/A | No es un servicio Fury/MeLi |

## Estadísticas

- Endpoints documentados: 37 (backend) / ~40 (consumidos por frontend)
- Entidades de dominio: 7 (Employee, Item, Inventory, InventoryDetail, Category, Supplier, MeasurementUnit)
- Casos de uso identificados: 18
- Reglas de negocio consolidadas: 16
- Features pendientes identificadas: 14
- Tablas de base de datos: 6 activas (+ 1 eliminada en migración 011)

## Archivos Generados

```
raw/
├── existing-specs/DETECTION_REPORT.md    ← Framework detection
├── mcpfury/NOT_APPLICABLE.md             ← FuryMCP no aplica
├── code-analysis/
│   ├── architecture/backend.md           ← Arquitectura Go
│   ├── architecture/frontend.md          ← Arquitectura Angular
│   ├── api-specs/endpoints.md            ← Todos los endpoints
│   ├── database/schema.md                ← Esquema MySQL
│   ├── deployment/infrastructure.md      ← GAE + Firebase
│   └── fury-services/NOT_APPLICABLE.md   ← No usa Fury services
└── README.md                             ← Este archivo
```
