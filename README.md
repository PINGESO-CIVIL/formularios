# Motor de Formularios (PT4)

La API Go vive en `apps/api`. La app React para desarrollar los microcomponentes vive en `apps/web`. El contrato compartido está en `contracts/openapi.yaml`.

## Desarrollo local

Instala Go 1.26.5, Node 22.22.0, pnpm 10.13.1 y Gitleaks. Desde la raíz:

```sh
pnpm install --frozen-lockfile
pnpm exec lefthook install
pnpm typecheck
pnpm verify
pnpm test
pnpm test:api
pnpm test:e2e:api
```

Ejecuta `pnpm --filter @pingeso-civil/formularios-ui dev` para abrir la web. Con `VITE_MOCKS=true` la app inicia MSW en desarrollo. Para la API, levanta PostgreSQL con `docker compose up -d postgres` y ejecuta `make run` desde `apps/api`. La API expone salud en `/healthz`, disponibilidad en `/readyz` y métricas en `/metrics`.

Consulta [el DAS](docs/referencia/das.md) y [ADR-0004](docs/adr/ADR-0004-adopcion-das.md) para la estructura y los criterios del curso.
