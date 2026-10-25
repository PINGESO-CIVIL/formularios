// Package http contiene el router, handlers, DTOs y middleware HTTP.
// Los handlers no tienen lógica de negocio: parsean, llaman al caso de uso y traducen.
package http

import (
	"context"
	"encoding/json"
	"net/http"
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
	Readiness []ReadinessCheck
	Metrics   prometheus.Gatherer
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthStatus{Status: "ok"})
	})
	r.Get("/readyz", readinessHandler(deps.Readiness))
	r.Method(http.MethodGet, "/metrics", promhttp.HandlerFor(deps.Metrics, promhttp.HandlerOpts{}))

	r.Route("/api/v1", func(_ chi.Router) {
		// Los recursos de negocio se agregan aquí según contracts/openapi.yaml.
	})

	return r
}

type healthStatus struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

func readinessHandler(checks []ReadinessCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		res := healthStatus{Status: "ok", Checks: map[string]string{}}
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
