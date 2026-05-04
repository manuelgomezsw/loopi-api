# Detection Report

**Generated**: 2026-03-26T20:30:00Z
**Repository**: loopi (monorepo)

## Extraction Scope

**Mode**: FULL EXTRACTION
**Focus Component**: Full Repository

## Detected Frameworks

| Framework | Confidence | Files Found |
|-----------|------------|-------------|
| Meli SDD Kit | 🟢 High | `meli/PROJECT.md`, `meli/specs/`, `meli/wip/`, `meli/extracted/` (vacíos) |
| Claude Code | 🟡 Medium | `.claude/settings.local.json`, `.claude/skills/` |
| Plain Docs | 🟢 High | `ARCHITECTURE.md` (112 líneas), `docs/` (8 documentos) |

## Selected Strategy

**Strategy**: ASSISTED

**Rationale**: No hay specs formales previas, pero existe documentación rica en texto plano (`ARCHITECTURE.md` + 8 documentos de análisis en `docs/`). El código fuente está disponible en los submodulos de git. No es un servicio Fury, por lo que FuryMCP no aplica.

## Detected Specs Summary

| Spec Type | Location | Last Modified |
|-----------|----------|---------------|
| Architecture (Plain Doc) | `ARCHITECTURE.md` | - |
| Inventory Analysis | `docs/ANALISIS_INVENTARIO_19_20_Y_PLAN.md` | - |
| Discrepancy Self-critique | `docs/AUTOCRITICA_DISCREPANCIAS_Y_PLAN.md` | - |
| Root Cause: Expected Values | `docs/CAUSA_RAIZ_ESPERADO_Y_SOLUCIONES.md` | - |
| Root Cause: News & Admin | `docs/CAUSA_RAIZ_NOVEDADES_Y_LISTA_ADMIN.md` | - |
| Root Cause: Summary Diff | `docs/CAUSA_RAIZ_SUMMARY_DIFERENCIAS.md` | - |
| Shifts Integration Plan | `docs/INTEGRACION_TURNOS_EN_REFACTOR.md` | - |
| Structural Adjustments Plan | `docs/PLAN_AJUSTES_ESTRUCTURALES_SENIOR.md` | - |
| Shrinkage & Shifts Plan | `docs/PLAN_MERMAS_Y_TURNOS.md` | - |

## Project Profile

| Atributo | Valor |
|----------|-------|
| **Nombre** | Loopi |
| **Tipo** | Aplicación de gestión de inventario |
| **Plataforma** | Personal / Independiente (no Fury/MeLi) |
| **Backend** | Go 1.24.0, Chi v5, MySQL, Google App Engine |
| **Frontend** | Angular 20.x, Firebase Hosting |
| **Arquitectura** | Clean Architecture (4 capas) |
| **Autenticación** | JWT |

## Extraction History

| Date | Mode | Focus | Summary |
|------|------|-------|---------|
| 2026-03-26 | FULL | - | Primera extracción completa desde docs y código |

## Recommendations

- Extraer specs del ARCHITECTURE.md y documentos de docs/
- Analizar código en `/Users/mangomez/Repos/Manu/loopi/loopi-api/` (submodulo backend)
- Analizar código en `/Users/mangomez/Repos/Manu/loopi/loopi-web/` (submodulo frontend)
- No se usa FuryMCP (proyecto no-Fury)
- Los documentos en `docs/` contienen análisis detallado del dominio Inventario que debe capturarse en functional-spec.md
