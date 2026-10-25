# AGENTS.md — backend (Go)

Reglas para trabajar en el servicio Go de `ms-formularios`. Lee también el
`AGENTS.md` de la raíz: las reglas globales (multi-tenant, inmutabilidad,
contrato primero, sin lógica por producto) aplican aquí.

## Stack

| Tema | Elección |
| --- | --- |
| Lenguaje | Go 1.25 o superior (ver `go.mod`) |
| HTTP | `net/http` + `github.com/go-chi/chi/v5` |
| Base de datos | PostgreSQL 16, driver `github.com/jackc/pgx/v5` |
| Consultas | `sqlc` (SQL en `db/queries/`, código generado en `internal/adapters/postgres/sqlcgen/`) |
| Migraciones | `golang-migrate` (archivos `db/migrations/NNNN_nombre.up.sql` / `.down.sql`) |
| Validación de esquemas | `github.com/santhosh-tekuri/jsonschema/v6` (JSON Schema 2020-12) |
| JWT | `github.com/lestrrat-go/jwx/v2` |
| Logs | `log/slog` en JSON |
| Métricas | `github.com/prometheus/client_golang` |
| Pruebas | `testing`, `github.com/stretchr/testify`, `github.com/testcontainers/testcontainers-go` |
| Lint | `golangci-lint` (config en `.golangci.yml`) |

No agregues dependencias nuevas sin justificarlo en el PR. Prefiere la
librería estándar cuando alcance.

## Comandos

```bash
make run            # levanta la API en modo dev (requiere docker compose up -d)
make worker         # levanta el worker (outbox relay)
make test           # pruebas unitarias
make test-int       # pruebas de integración (Postgres en contenedor)
make lint           # golangci-lint
make sqlc           # regenera código de sqlc
make migrate-up     # aplica migraciones
make migrate-new name=create_form_templates
make check          # lint + test + test-int: córrelo antes de terminar
```

Si un comando no existe todavía en el `Makefile`, créalo como parte de tu tarea.

## Estructura

```
backend/
├── cmd/
│   ├── api/main.go           # solo wiring: config, adaptadores, servidor
│   └── worker/main.go        # outbox relay; luego recurrencias
├── internal/
│   ├── domain/               # entidades y reglas; SIN imports de infraestructura
│   │   ├── template/         # Template, Version, estados, publicación
│   │   ├── submission/       # Submission, estados, content hash
│   │   ├── signature/
│   │   └── feature/          # features LPS por tenant
│   ├── app/                  # casos de uso, uno por archivo (publish_version.go)
│   ├── ports/                # interfaces hacia afuera
│   ├── adapters/
│   │   ├── http/             # handlers, DTOs, middleware, errores
│   │   ├── postgres/         # repositorios (implementan ports)
│   │   ├── auth/devauth/     # stub de identidad (fase 0)
│   │   ├── files/localfs/    # stub de archivos (fase 0)
│   │   └── events/           # outbox + logpublisher (fase 0)
│   ├── schema/               # compilación y validación de esquemas + formatos custom (rut)
│   └── config/
├── db/
│   ├── migrations/
│   └── queries/
└── test/integration/
```

### Reglas de dependencia (obligatorias)

- `domain` no importa nada de `app`, `adapters`, `ports`, ni paquetes de terceros
  de infraestructura (pgx, chi, jwx).
- `app` depende de `domain` y `ports`. Nunca de `adapters`.
- `adapters` implementan `ports` y pueden importar `domain`.
- `cmd` es el único lugar que conoce todos los adaptadores y los conecta.
- Los handlers HTTP no tienen lógica de negocio: parsean, llaman al caso de uso
  y traducen el resultado.

## Puertos y stubs de fase 0

Los servicios externos no existen todavía. Implementa contra estas interfaces y
usa los stubs; no llames a URLs reales de otros MS.

```go
// ports/identity.go
type Principal struct {
    UserID      string
    TenantID    string
    Permissions []string // ej. "forms:design", "forms:fill"
}
type Authenticator interface {
    Authenticate(ctx context.Context, r *http.Request) (Principal, error)
}

// ports/files.go
type FileStore interface {
    Exists(ctx context.Context, tenantID, fileID string) (bool, error)
}

// ports/events.go
type Event struct {
    Type          string          // "respuesta.enviada"
    Version       int
    OccurredAt    time.Time
    Actor         Actor           // id + rol
    Context       map[string]string
    CorrelationID string
    Payload       json.RawMessage
}
type EventPublisher interface {
    Publish(ctx context.Context, e Event) error
}

// ports/clock.go e ids.go: inyectar tiempo e IDs para pruebas deterministas
```

| Stub | Comportamiento | Activación |
| --- | --- | --- |
| `devauth` | Acepta JWT HS256 firmado con `DEV_JWT_SECRET`, o headers `X-Dev-User`, `X-Dev-Tenant`, `X-Dev-Permissions` | Solo si `APP_ENV=dev`; en cualquier otro entorno el servicio no arranca con este stub |
| `localfs` | `Exists` revisa una carpeta local; incluye endpoint dev para subir archivos de prueba | `FILESTORE=local` |
| `logpublisher` | El worker lee `outbox_events` y escribe cada evento en el log en JSON | `EVENTS=log` |

Los nombres de permisos (`forms:read`, `forms:design`, `forms:fill`,
`forms:review`, `forms:assign`, `forms:export`) son **provisionales**:
centralízalos en `internal/domain/permission.go` para cambiarlos en un solo lugar.

## Eventos

- Se escriben en `outbox_events` **en la misma transacción** que el cambio de
  datos. Nunca publiques directo desde un caso de uso.
- El sobre sigue el formato propuesto para la LPS: `eventId`, `tipo`, `version`,
  `ocurridoEn`, `actor`, `contexto`, `correlationId`, `servicioOrigen`
  (`ms-formularios`), `payload`. Es provisional: arma el sobre en un solo lugar
  (`adapters/events/envelope.go`).
- El payload lleva IDs y hashes, nunca el contenido de la respuesta.
- Cada tipo de evento tiene su JSON Schema en `contracts/events/` y una prueba
  que valida un ejemplo contra él.

## Base de datos

- Nombres en `snake_case`, tablas en plural, PK `id uuid` (UUIDv7 generado en Go).
- Toda tabla de negocio: `tenant_id uuid not null`, `created_at timestamptz not null default now()`.
- Esquemas y respuestas en `jsonb`. Valores que se filtran → `submission_values`, no índices sobre JSON arbitrario.
- Índices que empiezan por `tenant_id`.
- Migraciones pequeñas, una por cambio, siempre con `down`.
- No uses `ON DELETE CASCADE` sobre respuestas ni firmas.

## API HTTP

- Prefijo `/api/v1`. JSON en `camelCase`.
- Errores con `application/problem+json` (RFC 9457): `type`, `title`, `status`,
  `detail`, y `errors[]` con `{ field, code, message }` para validación.
  Los `message` van en español.
- Paginación por cursor: `?limit=&cursor=` → `{ items, nextCursor }`.
- `X-Correlation-Id`: se lee si viene, se genera si no, se devuelve en la respuesta y va en cada log.
- Los handlers se escriben contra `contracts/openapi.yaml`. Si generas tipos desde el contrato, usa `oapi-codegen` y no edites el código generado.

## Pruebas

- Dominio: pruebas unitarias tabla-dirigidas, sin base de datos.
- Casos de uso: con repositorios en memoria (`internal/app/apptest/`).
- Repositorios y migraciones: integración con testcontainers.
- HTTP: `httptest` contra el router completo con `devauth`.
- Nombra las pruebas por regla: `TestPublishVersion_RejectsAlreadyPublished`.
- Cada regla `RN-FOR-xx` implementada tiene al menos una prueba que la cita en un comentario.

## Qué construir ahora (en este orden)

1. Esqueleto: `go.mod`, `cmd/api`, config por env, `slog`, `/healthz`, `/readyz`, `/metrics`, `Makefile`, `Dockerfile` multi-stage, `.golangci.yml`, CI.
2. Middleware: correlation id, recover, logging de requests, autenticación vía `ports.Authenticator` con `devauth`, extracción de tenant.
3. Migraciones iniciales: `form_templates`, `form_versions`, `submissions`, `tenant_features`, `outbox_events`.
4. Paquete `schema`: compilar un esquema de formulario, validarlo y devolver errores por campo; formato custom `rut` con dígito verificador.
5. Dominio `template`: crear plantilla, borrador de versión, publicar, despublicar, archivar (RF-FOR-01, 04, 05).
6. Dominio `submission`: iniciar sobre una versión publicada, guardar borrador, enviar con validación, hash del contenido, inmutabilidad (RF-FOR-10 a 13).
7. Outbox + worker con `logpublisher`.
8. Listado de respuestas con filtros básicos y paginación (RF-FOR-25).

## Qué NO hacer en esta fase

- No implementes firma, adjuntos reales, revisión, asignaciones ni exportación hasta que se pida.
- No llames a otros microservicios ni supongas su URL, su formato de token o sus roles.
- No guardes usuarios, nombres ni emails: solo `user_id` del principal.
- No metas reglas de un producto en el dominio.
- No uses `devauth` fuera de `APP_ENV=dev`.
