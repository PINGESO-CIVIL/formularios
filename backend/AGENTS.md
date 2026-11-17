# AGENTS.md — backend (Go)

Reglas para trabajar en el servicio Go de `ms-formularios` (PT4). Lee también
el `AGENTS.md` de la raíz: las reglas globales (lineamientos, contrato primero,
multi-tenant, inmutabilidad, sin lógica por producto, no invadir otros MS)
aplican aquí.

## Stack

| Tema | Elección |
| --- | --- |
| Lenguaje | **Go 1.23** (lineamientos). `go.mod` debe declarar `go 1.23` |
| HTTP | `net/http` + `github.com/go-chi/chi/v5` |
| Base de datos | **PostgreSQL 16** (`ADR-0003`), driver `github.com/jackc/pgx/v5`, `sqlc` y `golang-migrate` |
| Contrato | `contracts/openapi.yaml` (OpenAPI 3.1.0); Swagger UI servido en `/docs` |
| Validación de esquemas de formulario | `github.com/santhosh-tekuri/jsonschema/v6` (JSON Schema 2020-12) |
| JWT | `github.com/lestrrat-go/jwx/v2` |
| Logs | `log/slog` en JSON |
| Pruebas | `testing`, `github.com/stretchr/testify`, `github.com/testcontainers/testcontainers-go` (PostgreSQL) |
| Lint | `gofmt` + `golangci-lint` (config en `.golangci.yml`) |
| Calidad / seguridad | SonarQube (reporte de cobertura `coverage.out`) y Trivy (imagen + `go.mod`) |

> Decisión de base de datos documentada en [`../docs/adr/ADR-0003-persistencia-postgresql-jsonb.md`](../docs/adr/ADR-0003-persistencia-postgresql-jsonb.md).

No agregues dependencias nuevas sin justificarlo en el PR. Prefiere la
librería estándar cuando alcance. Toda dependencia debe pasar Trivy.

## Comandos

```bash
make run            # levanta la API en modo dev (requiere docker compose up -d)
make test           # pruebas unitarias + coverage.out
make test-int       # pruebas de integración (PostgreSQL en contenedor)
make lint           # golangci-lint
make sqlc           # genera código tipado en internal/adapters/postgres/sqlcgen
make migrate-up     # aplica migraciones SQL en db/migrations
make migrate-down   # revierte la última migración SQL
make trivy          # escaneo de vulnerabilidades de dependencias e imagen
make check          # lint + test + test-int: córrelo antes de terminar
```

Si un comando no existe todavía en el `Makefile`, créalo como parte de tu tarea.

## Arquitectura: Handlers → Services → Repositories

Los lineamientos exigen Clean Architecture con esas tres capas. En este repo
se corresponden así:

| Capa (lineamientos) | Carpeta | Responsabilidad |
| --- | --- | --- |
| Handlers | `internal/adapters/http/` | Rutas, DTOs (`snake_case`), middleware, traducción de errores al `ErrorEstandar` |
| Services | `internal/app/` (+ reglas en `internal/domain/`) | Casos de uso, uno por archivo (`publish_version.go`) |
| Repositories | `internal/adapters/postgres/` | Acceso a PostgreSQL (`pgx/v5` + `sqlc`); implementan interfaces de `internal/ports/` |
| Mocks de otros MS | `internal/adapters/{auth,files,events}/` | PT6, PT7, PT8 en fase 1 |

```
backend/
├── cmd/api/main.go           # solo wiring: config, adaptadores, servidor
├── db/
│   ├── migrations/           # migraciones SQL (.up.sql / .down.sql)
│   └── queries/              # consultas SQL para sqlc
├── internal/
│   ├── domain/               # entidades y reglas; SIN imports de infraestructura
│   │   ├── template/         # Formulario, Versión, estados, publicación
│   │   ├── submission/       # Respuesta, estados, hash de contenido
│   │   ├── signature/
│   │   └── feature/          # features LPS por tenant
│   ├── app/                  # services / casos de uso
│   ├── ports/                # interfaces hacia afuera (repos y otros MS)
│   ├── adapters/
│   │   ├── http/             # handlers
│   │   ├── postgres/         # repositories (pgx/v5 + sqlcgen)
│   │   ├── auth/             # mock PT6 (JWT quemado) → luego adaptador real
│   │   ├── files/            # mock PT7 → luego adaptador real
│   │   └── events/           # mock PT8 (consola) → luego adaptador real
│   ├── schema/               # compilación y validación de esquemas + formatos custom (rut)
│   └── config/
└── test/integration/
```

### Reglas de dependencia (obligatorias)

- `domain` no importa nada de `app`, `adapters`, `ports` ni librerías de
  infraestructura (`pgx`, `chi`, `jwx`).
- `app` depende de `domain` y `ports`. Nunca de `adapters`.
- `adapters` implementan `ports` y pueden importar `domain`.
- `cmd` es el único lugar que conoce todos los adaptadores y los conecta.
- Los handlers no tienen lógica de negocio.

## Puertos y mocks (fase 1)

Los otros MS no están disponibles. Implementa contra estas interfaces y usa los
mocks; no llames a URLs reales de otros grupos.

```go
// ports/identity.go — PT6 Identity Provider
type Principal struct {
    UserID      string
    TenantID    string
    Permissions []string // provisionales, ver más abajo
}
type Authenticator interface {
    Authenticate(ctx context.Context, r *http.Request) (Principal, error)
}

// ports/files.go — PT7 Bóveda Documental
type FileStore interface {
    Exists(ctx context.Context, tenantID, fileID string) (bool, error)
}

// ports/events.go — PT8 Central Auditoría
type EventPublisher interface {
    Publish(ctx context.Context, e Event) error
}

// ports/clock.go e ids.go: inyectar tiempo e IDs para pruebas deterministas
```

| Mock | Comportamiento (según "Mock requerido" de lineamientos) | Activación |
| --- | --- | --- |
| auth (PT6) | Acepta un JWT de prueba quemado / firmado con `DEV_JWT_SECRET` que simula un usuario emisor válido | Solo si `APP_ENV=dev`; en otro entorno el servicio no arranca con este mock |
| files (PT7) | Asume que todo `file_id` existe y ya viene autenticado | `FILESTORE=mock` |
| events (PT8) | Imprime cada evento en consola en JSON | `EVENTS=console` |

Formato real del token, claims, nombres de roles y formato de logs de PT8:
**no los supongas**. `// TODO(integracion)` + `docs/preguntas-abiertas.md`.
Los permisos (`forms:read`, `forms:design`, `forms:fill`, `forms:review`) son
provisionales y se centralizan en un solo archivo del dominio.

## Eventos (hacia PT8)

- Un evento por acción: `formulario.creado`, `version.publicada`,
  `respuesta.enviada`, etc.
- Sobre provisional: `event_id`, `tipo`, `version`, `ocurrido_en`, `actor`,
  `tenant_id`, `correlation_id`, `servicio_origen` (`ms-formularios`), `payload`.
  Se arma en un solo lugar (`adapters/events/envelope.go`).
- El payload lleva IDs y hashes, **nunca** el contenido de la respuesta.
- Cada tipo de evento tiene su JSON Schema en `contracts/events/`.

## Base de datos (PostgreSQL)

- Tablas en inglés, `snake_case` y plural (definidas en `db/migrations/0001_init.up.sql`, ver `ADR-0003`):
  - `tenant_features` (`PK: tenant_id`, `features JSONB`).
  - `form_templates` (`PK: id UUID`, `tenant_id`, metadatos del formulario).
  - `form_versions` (`PK: id UUID`, `FK: template_id` con `ON DELETE RESTRICT`, `UNIQUE (tenant_id, template_id, version)`, `status` en `draft|published|archived`, `schema JSONB`, `ui_schema JSONB`).
  - `submissions` (`PK: id UUID`, `FK: template_id` y `version_id` con `ON DELETE RESTRICT`, `status` en `draft|submitted`, `subject_type`/`subject_id` opacos, `answers JSONB`, `content_hash`).
- `id` es `UUID`; se expone en la API como `id` (`type: string`, `format: uuid`).
- Toda tabla de negocio tiene `tenant_id` y `created_at` (`TIMESTAMPTZ`, UTC).
- Esquemas (`schema`, `ui_schema`) y respuestas (`answers`) se guardan en columnas
  `JSONB`, combinando el modelo dinámico de formularios con integridad referencial SQL.
- Índices compuestos que empiezan por `tenant_id`. Todo cambio de esquema o índice
  se versiona en `db/migrations/` con su archivo `.up.sql` y `.down.sql`.
- Nunca borrado físico (`ON DELETE RESTRICT`) de versiones publicadas ni respuestas enviadas.

## API HTTP

- Se escribe contra `contracts/openapi.yaml`. Si generas tipos, usa
  `oapi-codegen` y no edites el código generado.
- Prefijo `/api/v1`, recursos en español y plural, JSON en `snake_case`.
- Errores con el `ErrorEstandar` del contrato: `{ codigo, mensaje, detalles }`
  y, en validación de respuestas, `errores[]` con `{ campo, mensaje }`.
  Mensajes en español.
- `X-Correlation-Id`: se lee si viene, se genera si no, se devuelve y va en cada log.
- `/healthz` y `/readyz` sin autenticación.

## Pruebas

- Dominio: tabla-dirigidas, sin base de datos.
- Services: con repositorios en memoria.
- Repositories: integración con PostgreSQL en testcontainers.
- HTTP: `httptest` contra el router completo con el mock de PT6.
- Nombra las pruebas por regla: `TestPublishVersion_RejectsAlreadyPublished`.
- Cada regla `RN-FOR-xx` implementada tiene al menos una prueba que la cita.
- Genera `coverage.out` para SonarQube.

## Qué construir ahora: Hito 1 (13 de noviembre)

1. Contrato `contracts/openapi.yaml` completo (formularios, versiones,
   respuestas) en el estilo del swagger de ejemplo, validado.
2. Swagger UI servido por el backend.
3. Handlers de todos los endpoints del contrato respondiendo con **mocks**
   coherentes con el contrato (incluidos errores).
4. Mock de PT6 (JWT quemado) en el middleware de autenticación + extracción de tenant.
5. CI con lint, pruebas, SonarQube y Trivy.

**Hito 2 (4 de diciembre):** dominio real, validación de esquemas, repositorios
PostgreSQL, eventos a consola/PT8, cero mocks propios.

## Qué NO hacer en esta fase

- No implementes firma, adjuntos reales, revisión, asignaciones ni exportación
  hasta que se pida.
- No generes PDF/Excel (PT1), no envíes correos (PT10), no guardes archivos (PT7),
  no emitas tokens (PT6), no hagas bitácora propia (PT8).
- No llames a otros microservicios ni supongas su URL, formato de token o roles.
- No guardes nombres ni emails de usuarios: solo el `user_id` del token.
- No metas reglas de un producto en el dominio.
- No uses el mock de PT6 fuera de `APP_ENV=dev`.
