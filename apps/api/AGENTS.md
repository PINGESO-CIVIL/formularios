# AGENTS.md — apps/api (Go + Huma)

Reglas específicas de la API. Aplican también las del `AGENTS.md` de la raíz y el
DAS (`docs/referencia/das.md`), que mandan sobre este archivo.

## Capas (DAS §3)

| Capa    | Carpeta                | Qué hace                                                                 | Qué no hace                      |
| ------- | ---------------------- | ------------------------------------------------------------------------ | -------------------------------- |
| Handler | `internal/api/`        | Structs del contrato (tags `json`, validaciones, ejemplos), traduce HTTP | Reglas de negocio, SQL           |
| Service | `internal/service/`    | Reglas de negocio; define los puertos de almacenamiento que consume      | Importar `net/http`, Huma ni pgx |
| Store   | `internal/store/`      | Implementa los puertos: `memoria` (Hito 1) y PostgreSQL (Hito 2)         | Reglas de negocio                |
| —       | `internal/repository/` | Generado por sqlc desde `db/queries/`                                    | Se edita a mano: **nunca**       |
| Armado  | `cmd/api/`             | Configuración, conexión, señales, subcomando `openapi`                   | Lógica                           |

- `internal/api` depende de **interfaces** del service, nunca del store.
- `cmd/api` es el único lugar que conoce todas las piezas y las conecta.

## Contrato

- Cada operación se registra con `huma.Register` y declara `x-permiso`
  (`publico`, `autenticado` o `<recurso>:<accion>`). Hay una prueba que falla si falta.
- Toda entrada se valida con tags de Huma: `minLength`, `maxLength`, `pattern`,
  `minimum`, `maximum`, `enum`, `format:"uuid"`. Lo inválido responde 422.
- Después de tocar un struct o una ruta: `pnpm codegen` y commitear
  `contracts/openapi.yaml` y `apps/web/src/api/esquema.d.ts`.
- Errores con `huma.ErrorNNN...` (RFC 9457). Mensajes en español y sin detalles
  internos (SQL, trazas, rutas).

## Datos

- Migraciones nuevas: `make migracion-nueva nombre=<algo>` (goose, archivo `.sql`
  con `-- +goose Up` y `-- +goose Down`). Una migración ya publicada no se edita.
- Queries en `db/queries/*.sql` y `make generate` (sqlc v1.29.0).
- Fechas `timestamptz`, ids `uuid`, estructuras dinámicas en `jsonb`.
- Toda operación que cambia datos escribe en `auditoria` en la misma transacción.

## Comandos (desde la raíz)

```bash
pnpm test:api          # go test de la API y del spa-server
pnpm verify:api        # gofmt + go vet
pnpm codegen           # regenera el contrato y los tipos TypeScript
docker compose up -d --build
```

## Qué NO hacer

- No llamar a `internal/store` desde `internal/api`.
- No migrar al arrancar la API.
- No loguear cuerpos, tokens, cookies ni el contenido de las respuestas.
- No agregar dependencias sin fijar versión exacta ni sin pasar govulncheck y Trivy.
