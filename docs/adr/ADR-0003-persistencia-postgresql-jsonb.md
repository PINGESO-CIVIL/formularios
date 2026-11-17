# ADR-0003: Persistencia en PostgreSQL con JSONB para esquemas y respuestas dinámicas

- Estado: aceptado
- Fecha: 2026-11-16
- Reemplaza parcialmente a: ADR-0002 (sección de base de datos e identificadores)

## Contexto

Aunque los lineamientos iniciales del curso mencionaban MongoDB como base de
datos de referencia (`ADR-0002`), la base técnica del repositorio (`docker-compose.yml`,
`backend/Makefile`, `backend/sqlc.yaml`, `backend/db/migrations`, `pgx/v5`) fue
construida sobre **PostgreSQL 16**.

`ms-formularios` (PT4) necesita combinar dos mundos:
1. **Metadatos relacionales estrictos e inmutabilidad transaccional:** relación
   entre formularios (`form_templates`), sus versiones (`form_versions`),
   respuestas (`submissions`) y configuración por tenant (`tenant_features`),
   garantizando integridad referencial y aislamiento por `tenant_id`.
2. **Estructuras dinámicas por formulario y tenant:** cada versión define un
   `schema` (JSON Schema 2020-12) y un `ui_schema` arbitrarios, y cada respuesta
   almacena un conjunto variable de valores (`answers`).

## Decisión

- Se adopta oficialmente **PostgreSQL 16** como motor de base de datos del
  microservicio.
- Las estructuras dinámicas (`schema`, `ui_schema`, `answers`, `features`) se
  almacenan en columnas nativas **`JSONB`**, combinando flexibilidad documental
  con garantías ACID e integridad referencial entre plantillas, versiones y
  respuestas.
- El acceso a datos en Go se implementa en `backend/internal/adapters/postgres/`
  usando `github.com/jackc/pgx/v5` (`pgxpool`) y consultas tipadas generadas con
  `sqlc` (`backend/sqlc.yaml`).
- El esquema relacional se gestiona mediante migraciones SQL versionadas y
  reversibles (`up`/`down`) en `backend/db/migrations/` ejecutadas con
  `golang-migrate`.
- Los identificadores públicos y primarios de las entidades de negocio pasan a
  ser **UUID** (`format: uuid` en OpenAPI) en lugar de `ObjectID` de MongoDB.

## Modelo relacional base (`0001_init`)

La migración inicial (`backend/db/migrations/0001_init.up.sql`) define las 4
tablas del dominio con las siguientes claves, restricciones e índices:

| Tabla | Clave Primaria (PK) | Claves Foráneas (FK) y Restricciones | Columnas `JSONB` | Índices compuestos (multi-tenant) |
| --- | --- | --- | --- | --- |
| `tenant_features` | `tenant_id TEXT` | — | `features` (features LPS por tenant) | PK por `tenant_id` |
| `form_templates` | `id UUID` | — | — | `(tenant_id, created_at DESC)` |
| `form_versions` | `id UUID` | `template_id` → `form_templates(id)` (`ON DELETE RESTRICT`)<br>`UNIQUE (tenant_id, template_id, version)`<br>`CHECK (version >= 1)`<br>`CHECK (status IN ('draft', 'published', 'archived'))` | `schema` (JSON Schema 2020-12)<br>`ui_schema` | `(tenant_id, template_id, status)` |
| `submissions` | `id UUID` | `template_id` → `form_templates(id)` (`ON DELETE RESTRICT`)<br>`version_id` → `form_versions(id)` (`ON DELETE RESTRICT`)<br>`CHECK (status IN ('draft', 'submitted'))` | `answers` (respuestas enviadas/borrador) | `(tenant_id, template_id, created_at DESC)`<br>`(tenant_id, subject_type, subject_id)` parcial |

Justificación de las restricciones clave:
- **`ON DELETE RESTRICT` en todas las FKs:** impide el borrado físico en cascada
  de plantillas o versiones que tengan historial asociado, protegiendo las
  reglas de inmutabilidad (`RN-FOR-01`, `RN-FOR-04`).
- **`UNIQUE (tenant_id, template_id, version)`:** garantiza que la numeración de
  versiones de un formulario sea correlativa y única dentro de cada tenant.
- **`subject_type` y `subject_id` opacos en `submissions`:** permiten asociar
  una respuesta a cualquier entidad de un producto derivado de la LPS (paciente,
  vehículo, contrato, etc.) sin acoplar el esquema a otros microservicios.

## Consecuencias

- Se eliminan las contradicciones entre la documentación normativa (`AGENTS.md`,
  `backend/AGENTS.md`, `README.md`) y la infraestructura existente en el repo.
- Toda tabla de negocio incluye `tenant_id` e índices compuestos encabezados por
  `tenant_id` para asegurar el filtrado multi-tenant eficiente.
- Toda migración `.up.sql` debe ir acompañada obligatoriamente de su `.down.sql`.
