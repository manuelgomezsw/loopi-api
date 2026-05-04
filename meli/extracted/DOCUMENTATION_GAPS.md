# Reporte de Brechas de Documentación (Phase 2)

**Generado**: 2026-03-26
**Aplicación**: Loopi

## Resumen de Cobertura por Fuente

| Fuente | Cobertura | Notas |
|--------|-----------|-------|
| **Código Go (backend)** | 100% | Fuente de verdad. 37 endpoints, 7 entidades, esquema completo. |
| **Código Angular (frontend)** | 100% | ~40 endpoints consumidos, modelos TypeScript completos. |
| **ARCHITECTURE.md** | 85% | Cubre dominios y fórmulas correctamente, pero no documenta todos los endpoints admin ni el flujo de turnos futuro. |
| **docs/ (8 documentos)** | N/A | Análisis técnico de bugs y planes. No son specs, son historial de decisiones. |
| **FuryMCP** | N/A | No aplica (proyecto no-Fury). |

**Prioridad de fuentes (conflictos):** CÓDIGO > ARCHITECTURE.md > docs/

## Brechas Identificadas

### 🔴 Alta Prioridad

| ID | Brecha | Impacto |
|----|--------|---------|
| GAP-001 | README.md del backend solo documenta ~10 endpoints de los 37 existentes | Onboarding engañoso |
| GAP-002 | No hay tests automatizados (unitarios ni de integración) documentados como existentes | Sin cobertura de tests |
| GAP-003 | La feature de "4 turnos" (morning_open/close, afternoon_open/close) está planificada pero no implementada. El sistema actual usa 3 turnos (opening/noon/closing). | Roadmap de producto |
| GAP-004 | No hay documentación de errores HTTP (qué errores retorna cada endpoint y en qué casos) | Integración frontend/API |

### 🟡 Media Prioridad

| ID | Brecha | Impacto |
|----|--------|---------|
| GAP-005 | `GetRecentDiscrepancies` y `GetDashboardStats` en repos aún usan definición antigua (`suggested != real`) según docs. Código real podría estar actualizado. | Consistencia de datos |
| GAP-006 | No hay documentación de rate limiting, timeouts, o políticas de retry | Resiliencia |
| GAP-007 | Los docs mencionan `inventory_issues` (entidad eliminada). Ya no existe en el código. | Deuda de documentación |
| GAP-008 | El campo `shrinkage` es editable en el panel admin, pero no hay documentación de quién puede registrar mermas y en qué condiciones | Regla de negocio parcialmente documentada |

### 🟢 Baja Prioridad

| ID | Brecha | Impacto |
|----|--------|---------|
| GAP-009 | No hay documentación de pagination defaults y límites en todos los endpoints | Comportamiento por defecto |
| GAP-010 | No hay documentación de la lógica de `suggested-schedule` (horarios exactos para cada turno) | UX en frontend |
| GAP-011 | No hay changelog o historial de versiones del API | Trazabilidad |

## Actor Discovery

| Método | Resultado |
|--------|-----------|
| MeliSystemMCP | No aplica |
| FuryMCP | No aplica |
| **Código (CORS config)** | **Frontend Angular** (`loopi-c048d.web.app`) + **localhost:4200** |
| **Código (router)** | Actores identificados: Empleado (rol "employee"), Administrador (rol "admin") |

**Actores identificados**: 2 internos (Empleado, Administrador) + 1 sistema (Frontend Angular).

No hay integraciones con sistemas externos (no webhooks, no colas de mensajes, no APIs de terceros detectadas).
