-- name: GetTenantFeatures :one
SELECT tenant_id, features, created_at, updated_at
FROM tenant_features
WHERE tenant_id = $1;

