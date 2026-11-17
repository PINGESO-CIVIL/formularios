-- Esquema base multi-tenant para ms-formularios (PT4) sobre PostgreSQL + JSONB.
-- Documentado en:
--   * docs/adr/ADR-0003-persistencia-postgresql-jsonb.md (tablas, claves, restricciones e índices)
--   * backend/AGENTS.md (sección "Base de datos (PostgreSQL)")

-- 1. Configuración de features habilitadas por tenant en la LPS.
CREATE TABLE IF NOT EXISTS tenant_features (
    tenant_id   TEXT        PRIMARY KEY,
    features    JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Definición maestra de un formulario dentro de un tenant.
CREATE TABLE IF NOT EXISTS form_templates (
    id          UUID        PRIMARY KEY,
    tenant_id   TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    created_by  TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_form_templates_tenant_created
    ON form_templates (tenant_id, created_at DESC);

-- 3. Versiones de un formulario (borrador -> publicado -> archivado).
--    * ON DELETE RESTRICT evita borrar plantillas con versiones existentes (RN-FOR-01).
--    * UNIQUE (tenant_id, template_id, version) asegura numeración correlativa por formulario.
CREATE TABLE IF NOT EXISTS form_versions (
    id           UUID        PRIMARY KEY,
    tenant_id    TEXT        NOT NULL,
    template_id  UUID        NOT NULL REFERENCES form_templates(id) ON DELETE RESTRICT,
    version      INTEGER     NOT NULL CHECK (version >= 1),
    status       TEXT        NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
    schema       JSONB       NOT NULL,
    ui_schema    JSONB       NOT NULL DEFAULT '{}'::jsonb,
    created_by   TEXT        NOT NULL,
    published_at TIMESTAMPTZ,
    archived_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_form_versions_tenant_template_version UNIQUE (tenant_id, template_id, version)
);

CREATE INDEX IF NOT EXISTS idx_form_versions_tenant_template_status
    ON form_versions (tenant_id, template_id, status);

-- 4. Respuestas de usuarios sobre una versión de formulario (borrador -> enviada).
--    * ON DELETE RESTRICT protege la inmutabilidad del historial (RN-FOR-04).
--    * subject_type y subject_id son referencias opacas a entidades externas de la LPS.
CREATE TABLE IF NOT EXISTS submissions (
    id           UUID        PRIMARY KEY,
    tenant_id    TEXT        NOT NULL,
    template_id  UUID        NOT NULL REFERENCES form_templates(id) ON DELETE RESTRICT,
    version_id   UUID        NOT NULL REFERENCES form_versions(id) ON DELETE RESTRICT,
    status       TEXT        NOT NULL CHECK (status IN ('draft', 'submitted')),
    subject_type TEXT,
    subject_id   TEXT,
    answers      JSONB       NOT NULL DEFAULT '{}'::jsonb,
    content_hash TEXT,
    submitted_by TEXT        NOT NULL,
    submitted_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_submissions_tenant_template_created
    ON submissions (tenant_id, template_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_submissions_tenant_subject
    ON submissions (tenant_id, subject_type, subject_id)
    WHERE subject_type IS NOT NULL AND subject_id IS NOT NULL;
