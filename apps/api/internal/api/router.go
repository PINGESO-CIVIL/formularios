// Package api contiene el router, handlers Huma y structs del contrato HTTP.
// Los handlers no tienen lógica de negocio: parsean, llaman al service y traducen errores.
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/service"
)

// ReadinessCheck verifica una dependencia (ej. base de datos) para /readyz.
type ReadinessCheck struct {
	Name  string
	Check func(ctx context.Context) error
}

// RouterDeps agrupa las dependencias inyectadas al router HTTP desde cmd/api.
type RouterDeps struct {
	Readiness    []ReadinessCheck
	Metrics      prometheus.Gatherer
	PublicarDocs bool
	Respuestas   service.ServicioRespuestas
}

// NewRouter construye el manejador HTTP completo del servicio.
func NewRouter(deps RouterDeps) http.Handler {
	r, _ := servidor(deps)
	return r
}

// Contrato genera la especificación OpenAPI 3.1 (YAML) a partir de las mismas rutas que atiende la API.
func Contrato() ([]byte, error) {
	_, api := servidor(RouterDeps{})
	return api.OpenAPI().YAML()
}

func servidor(deps RouterDeps) (*chi.Mux, huma.API) {
	r := chi.NewRouter()
	cfg := huma.DefaultConfig("ms-formularios", "0.1.0")
	cfg.Info.Description = "Servicio de formularios de la LPS (PT4): definición de formularios " +
		"dinámicos, versiones publicadas y recolección de respuestas.\n\n" +
		"Cada operación declara en `x-permiso` lo que exige (`publico`, `autenticado` o `<recurso>:<accion>`). " +
		"Los errores siguen RFC 9457 (`application/problem+json`)."
	cfg.DocsPath = ""
	cfg.OpenAPIPath = ""
	cfg.SchemasPath = ""
	if deps.PublicarDocs {
		cfg.DocsPath = "/api/docs"
		cfg.OpenAPIPath = "/api/openapi"
	}
	api := humachi.New(r, cfg)

	salud := api.OpenAPI().Components.Schemas.Schema(reflect.TypeFor[HealthStatus](), true, "HealthStatus")
	respuesta := func(descripcion, tipo string, esquema *huma.Schema) *huma.Response {
		return &huma.Response{Description: descripcion, Content: map[string]*huma.MediaType{tipo: {Schema: esquema}}}
	}
	registrar := func(ruta, id, resumen string, respuestas map[string]*huma.Response, h http.Handler) {
		op := &huma.Operation{
			Method:      http.MethodGet,
			Path:        ruta,
			OperationID: id,
			Summary:     resumen,
			Responses:   respuestas,
			Extensions:  map[string]any{extensionPermiso: permisoPublico},
		}
		api.OpenAPI().AddOperation(op)
		api.Adapter().Handle(op, func(ctx huma.Context) { req, w := humachi.Unwrap(ctx); h.ServeHTTP(w, req) })
	}

	ok := respuesta("Disponible", "application/json", salud)
	registrar("/healthz", "getHealth", "El proceso está vivo", map[string]*huma.Response{"200": ok}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, HealthStatus{Status: "ok"}) }))
	registrar("/readyz", "getReadiness", "Disponibilidad del servicio", map[string]*huma.Response{"200": ok, "503": respuesta("Dependencia no disponible", "application/json", salud)}, readinessHandler(deps.Readiness))
	registrar("/metrics", "getMetrics", "Métricas de Prometheus", map[string]*huma.Response{"200": respuesta("Métricas", "text/plain", &huma.Schema{Type: "string"})}, promhttp.HandlerFor(deps.Metrics, promhttp.HandlerOpts{}))

	registrarRespuestas(api, deps.Respuestas)

	return r, api
}

// HealthStatus describe la respuesta de los endpoints operativos de salud.
type HealthStatus struct {
	Status string            `json:"status" enum:"ok,unavailable"`
	Checks map[string]string `json:"checks,omitempty"`
}

func readinessHandler(checks []ReadinessCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		res := HealthStatus{Status: "ok", Checks: map[string]string{}}
		code := http.StatusOK
		for _, c := range checks {
			if err := c.Check(ctx); err != nil {
				res.Status = "unavailable"
				res.Checks[c.Name] = "unavailable"
				code = http.StatusServiceUnavailable
				continue
			}
			res.Checks[c.Name] = "ok"
		}
		writeJSON(w, code, res)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
