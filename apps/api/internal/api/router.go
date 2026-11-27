// Package http contiene el router, handlers, DTOs y middleware HTTP.
// Los handlers no tienen lÃ³gica de negocio: parsean, llaman al caso de uso y traducen.
package http

import (
	"context"
	"encoding/json"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"net/http"
	"reflect"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ReadinessCheck verifica una dependencia (ej. base de datos) para /readyz.
type ReadinessCheck struct {
	Name  string
	Check func(ctx context.Context) error
}

type RouterDeps struct {
	Readiness    []ReadinessCheck
	Metrics      prometheus.Gatherer
	PublicarDocs bool
}

func NewRouter(deps RouterDeps) http.Handler {
	r, _ := servidor(deps)
	return r
}

// Contrato genera la especificación a partir de las mismas rutas que atiende la API.
func Contrato() ([]byte, error) {
	_, api := servidor(RouterDeps{})
	return api.OpenAPI().YAML()
}

func servidor(deps RouterDeps) (*chi.Mux, huma.API) {
	r := chi.NewRouter()
	cfg := huma.DefaultConfig("ms-formularios", "0.1.0")
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
		op := &huma.Operation{Method: http.MethodGet, Path: ruta, OperationID: id, Summary: resumen, Responses: respuestas}
		api.OpenAPI().AddOperation(op)
		api.Adapter().Handle(op, func(ctx huma.Context) { req, w := humachi.Unwrap(ctx); h.ServeHTTP(w, req) })
	}
	ok := respuesta("Disponible", "application/json", salud)
	registrar("/healthz", "getHealth", "El proceso está vivo", map[string]*huma.Response{"200": ok}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, HealthStatus{Status: "ok"}) }))
	registrar("/readyz", "getReadiness", "Disponibilidad del servicio", map[string]*huma.Response{"200": ok, "503": respuesta("Dependencia no disponible", "application/json", salud)}, readinessHandler(deps.Readiness))
	registrar("/metrics", "getMetrics", "Métricas de Prometheus", map[string]*huma.Response{"200": respuesta("Métricas", "text/plain", &huma.Schema{Type: "string"})}, promhttp.HandlerFor(deps.Metrics, promhttp.HandlerOpts{}))
	return r, api
}

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
