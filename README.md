# Formularios (PT 4)

Servicio de **formularios dinámicos** de la Línea de Productos de Software de
PINGESO 2-2026: define formularios, los versiona y recolecta respuestas validadas,
de forma genérica para cualquier producto. Entrega una API (Go + Huma) y
microcomponentes React moldeables visualmente. Sigue el documento
**DAS — Desarrollo de Microservicios**.

## Evidencia

- Pipeline: https://github.com/PINGESO-CIVIL/formularios/actions
- SonarQube: _pendiente de los secrets del ayudante_
- Contrato: [`contracts/openapi.yaml`](contracts/openapi.yaml)
- Modos de falla: [`docs/modos-de-falla.md`](docs/modos-de-falla.md)
- Normas del curso: [`docs/referencia/das.md`](docs/referencia/das.md) ·
  [`docs/referencia/hitos-1-y-2.md`](docs/referencia/hitos-1-y-2.md)

## Puertos locales (servicio N = 4, por confirmar)

| Componente | URL                            |
| ---------- | ------------------------------ |
| Web        | http://localhost:5140          |
| API        | http://localhost:5141          |
| Docs API   | http://localhost:5141/api/docs |

## Puesta en marcha

Requisitos: Docker, Node 22.22.0 (`nvm use`), pnpm 10.13.1, Go 1.26 y Gitleaks.

```bash
pnpm install
pnpm exec lefthook install      # .npmrc tiene ignore-scripts: los hooks se instalan a mano
cp .env.example .env            # completar los valores replace-with-*
docker compose up -d --build    # postgres -> migrate -> backend -> frontend
```

Para validar tokens hace falta el servicio **usuarios-roles** corriendo en
`localhost:5161` (ver su `docs/jwt-de-prueba.md`).

## Comandos

| Comando          | Qué hace                                                                |
| ---------------- | ----------------------------------------------------------------------- |
| `pnpm codegen`   | Regenera `contracts/openapi.yaml` (desde los structs Go) y los tipos TS |
| `pnpm verify`    | Prettier, oxlint, gofmt y `go vet`                                      |
| `pnpm typecheck` | Tipos del frontend                                                      |
| `pnpm test`      | Pruebas del frontend (Vitest)                                           |
| `pnpm test:api`  | Pruebas de la API y del servidor web                                    |

## Estructura

| Carpeta      | Contenido                                                                         |
| ------------ | --------------------------------------------------------------------------------- |
| `apps/api/`  | API Go: `internal/api` → `internal/service` → `internal/store`, migraciones goose |
| `apps/web/`  | Microcomponentes React, app de desarrollo y spa-server                            |
| `contracts/` | Contrato generado, contratos externos y formato del esquema de formulario         |
| `docs/`      | Modos de falla, ADRs, preguntas abiertas y material del curso                     |

Reglas de trabajo: [`AGENTS.md`](AGENTS.md).
