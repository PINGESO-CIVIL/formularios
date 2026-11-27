package ports

import "context"

// FileStore abstrae al MS de archivos.
// TODO(integracion): API real del MS de archivos (docs/preguntas-abiertas.md #4).
type FileStore interface {
	Exists(ctx context.Context, tenantID, fileID string) (bool, error)
}
