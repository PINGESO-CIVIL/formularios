# ms-formularios

Microservicio de **Formularios** de la Línea de Productos de Software (LPS) de
gestión de trazabilidad/inventario (PINGESO, USACH).

> **Fase 0:** solo estructura base. Las atribuciones de este MS frente a los
> demás aún se están definiendo (ver [`docs/preguntas-abiertas.md`](docs/preguntas-abiertas.md)).

## Estructura

| Carpeta | Contenido |
| --- | --- |
| `backend/` | Servicio Go (API + worker). Ver `backend/AGENTS.md` |
| `frontend/` | Monorepo Vue (pnpm): paquetes `forms-*` y app `forms-admin`. Ver `frontend/AGENTS.md` |
| `contracts/` | OpenAPI, formato de esquemas de formulario y eventos |
| `docs/` | Requerimientos, preguntas abiertas y ADRs |

Las reglas de trabajo están en [`AGENTS.md`](AGENTS.md).

## Inicio rápido

```bash
cp .env.example .env
docker compose up -d          # postgres + minio

# backend
cd backend && make run        # http://localhost:8080/healthz

# frontend (no requiere backend: usa MSW)
cd frontend && pnpm install && pnpm dev
```
