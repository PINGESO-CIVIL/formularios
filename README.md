# ms-formularios

**PT4 · Motor de Formularios** de la Línea de Productos de Software (LPS) del
curso Proyecto de Ingeniería en Software (PINGESO, USACH).

API para crear formularios dinámicos, versionarlos y recolectar las respuestas
de los usuarios, de forma genérica para cualquier producto derivado de la LPS.

- Lineamientos técnicos del curso: [`docs/referencia/lineamientos.pdf`](docs/referencia/lineamientos.pdf)
- Reglas de trabajo y estándares: [`AGENTS.md`](AGENTS.md)
- Decisiones: [`docs/adr/`](docs/adr/) · Dudas con otros PT: [`docs/preguntas-abiertas.md`](docs/preguntas-abiertas.md)

## Stack

OpenAPI 3.1.0 · Go 1.23 · Vue 3.5 (Composition API) · PostgreSQL 16 (`JSONB`) · SonarQube · Trivy

> Ver [`docs/adr/ADR-0003-persistencia-postgresql-jsonb.md`](docs/adr/ADR-0003-persistencia-postgresql-jsonb.md) sobre la adopción de PostgreSQL con `JSONB`.

## Estructura

| Carpeta | Contenido |
| --- | --- |
| `contracts/` | Contrato OpenAPI (fuente de verdad), formato de esquemas de formulario y eventos |
| `backend/` | Servicio Go (Handlers → Services → Repositories). Ver `backend/AGENTS.md` |
| `frontend/` | Monorepo Vue (pnpm): paquetes `forms-*` y app `forms-admin`. Ver `frontend/AGENTS.md` |
| `docs/` | Requerimientos, preguntas abiertas, ADRs y material de referencia del curso |

## Hitos

| Hito | Fecha | Entregable |
| --- | --- | --- |
| Hito 1 | 13 de noviembre | Contrato publicado y Swagger validado; endpoints con mocks |
| Hito 2 | 4 de diciembre | Lógica real, persistencia, umbrales de Trivy y SonarQube |
| Fase 3 | dic – ene | Derivación LPS, despliegue cloud por cliente, validación |

## Inicio rápido

```bash
cp .env.example .env

# backend
cd backend && make run        # http://localhost:8080/healthz

# frontend (no requiere backend: usa MSW)
cd frontend && pnpm install && pnpm dev
```
