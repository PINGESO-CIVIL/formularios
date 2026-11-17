-- Reversa de la migración inicial 0001_init.

DROP INDEX IF EXISTS idx_submissions_tenant_subject;
DROP INDEX IF EXISTS idx_submissions_tenant_template_created;
DROP TABLE IF EXISTS submissions;

DROP INDEX IF EXISTS idx_form_versions_tenant_template_status;
DROP TABLE IF EXISTS form_versions;

DROP INDEX IF EXISTS idx_form_templates_tenant_created;
DROP TABLE IF EXISTS form_templates;

DROP TABLE IF EXISTS tenant_features;

