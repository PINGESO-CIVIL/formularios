package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func newTestRouter(checks ...ReadinessCheck) http.Handler {
	return NewRouter(RouterDeps{Readiness: checks, Metrics: prometheus.NewRegistry()})
}

func do(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil))
	return rec
}

func TestHealthz_ReturnsOK(t *testing.T) {
	rec := do(newTestRouter(), "/healthz")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func TestReadyz_ReturnsOKWhenChecksPass(t *testing.T) {
	ok := ReadinessCheck{Name: "db", Check: func(context.Context) error { return nil }}
	rec := do(newTestRouter(ok), "/readyz")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok","checks":{"db":"ok"}}`, rec.Body.String())
}

func TestReadyz_Returns503WhenACheckFails(t *testing.T) {
	failing := ReadinessCheck{Name: "db", Check: func(context.Context) error { return errors.New("down") }}
	rec := do(newTestRouter(failing), "/readyz")

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.JSONEq(t, `{"status":"unavailable","checks":{"db":"unavailable"}}`, rec.Body.String())
}

func TestMetrics_IsExposed(t *testing.T) {
	rec := do(newTestRouter(), "/metrics")

	assert.Equal(t, http.StatusOK, rec.Code)
}
