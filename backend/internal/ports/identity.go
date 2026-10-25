package ports

import (
	"context"
	"net/http"
)

// Principal es la identidad autenticada. El TenantID sale de aquí, nunca del cuerpo de la petición.
// TODO(integracion): formato real del token y de los permisos (docs/preguntas-abiertas.md #2, #3).
type Principal struct {
	UserID      string
	TenantID    string
	Permissions []string
}

type Authenticator interface {
	Authenticate(ctx context.Context, r *http.Request) (Principal, error)
}
