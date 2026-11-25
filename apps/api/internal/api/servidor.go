// Paquete api contiene los handlers HTTP (Huma) y los structs que forman el contrato.
// Solo traduce HTTP: las reglas de negocio viven en internal/service y nunca se llama
// a internal/store desde aquí.
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

// Titulo y Version identifican el contrato publicado en contracts/openapi.yaml.
const (
	Titulo  = "API Formularios"
	Version = "0.1.0"
)

// Valores de la extensión x-permiso del contrato.
const (
	extensionPermiso = "x-permiso"
	permisoPublico   = "publico"
)

// VerificadorBD comprueba que la base de datos responde. pgxpool.Pool lo cumple.
type VerificadorBD interface {
	Ping(ctx context.Context) error
}

// Dependencias agrupa lo que los handlers necesitan para funcionar.
type Dependencias struct {
	BD           VerificadorBD
	PublicarDocs bool         // true en dev y test: publica /api/docs y el spec
	Logger       *slog.Logger // si es nil, no se registran las peticiones
}

// NuevoServidor arma el router con todas las rutas y devuelve también la API de Huma,
// que se usa para generar el contrato OpenAPI.
func NuevoServidor(dep Dependencias) (http.Handler, huma.API) {
	mux := http.NewServeMux()

	cfg := huma.DefaultConfig(Titulo, Version)
	cfg.Info.Description = "Servicio de formularios de la LPS (PT 4): definición de formularios " +
		"dinámicos, versiones publicadas y recolección de respuestas.\n\n" +
		"Cada operación declara en `x-permiso` lo que exige: `publico`, `autenticado` o un permiso " +
		"`<recurso>:<accion>`. Sin token válido se responde 401; con token pero sin permiso, 403.\n\n" +
		"Los tokens los emite el servicio usuarios-roles y se validan con su JWKS (`AUTH_JWKS_URL`)."
	if dep.PublicarDocs {
		cfg.DocsPath = "/api/docs"
		cfg.OpenAPIPath = "/api/openapi"
	} else {
		cfg.DocsPath = ""
		cfg.OpenAPIPath = ""
		cfg.SchemasPath = ""
	}

	api := humago.New(mux, cfg)
	registrarSalud(api, dep.BD)

	return conObservabilidad(mux, dep.Logger), api
}

// publica marca una operación como accesible sin token.
func publica(op huma.Operation) huma.Operation {
	if op.Extensions == nil {
		op.Extensions = map[string]any{}
	}
	op.Extensions[extensionPermiso] = permisoPublico
	return op
}
