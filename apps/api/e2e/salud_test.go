package e2e

import (
	api "github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/api"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServidorExponeSalud(t *testing.T) {
	servidor := httptest.NewServer(api.NewRouter(api.RouterDeps{Metrics: prometheus.NewRegistry()}))
	defer servidor.Close()
	resp, err := servidor.Client().Get(servidor.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("estado inesperado: %d", resp.StatusCode)
	}
}
