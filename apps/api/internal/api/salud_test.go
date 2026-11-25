package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
)

// bdFalsa simula PostgreSQL: responde bien o con error según el campo err.
type bdFalsa struct{ err error }

func (b bdFalsa) Ping(context.Context) error { return b.err }

// pedir hace una petición GET al servidor y devuelve la respuesta grabada.
func pedir(t *testing.T, dep Dependencias, ruta string) *httptest.ResponseRecorder {
	t.Helper()
	servidor, _ := NuevoServidor(dep)
	rec := httptest.NewRecorder()
	servidor.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
	return rec
}

func TestHealthzRespondeOK(t *testing.T) {
	rec := pedir(t, Dependencias{}, "/healthz")
	if rec.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, llegó %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"estado":"ok"`) {
		t.Fatalf("cuerpo inesperado: %s", rec.Body.String())
	}
}

func TestReadyzConBaseDisponible(t *testing.T) {
	if rec := pedir(t, Dependencias{BD: bdFalsa{}}, "/readyz"); rec.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, llegó %d", rec.Code)
	}
}

// MF-01: /readyz responde 503 en problem+json si PostgreSQL no contesta.
func TestReadyzConBaseCaida(t *testing.T) {
	rec := pedir(t, Dependencias{BD: bdFalsa{err: errors.New("sin conexión")}}, "/readyz")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("se esperaba 503, llegó %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "application/problem+json") {
		t.Fatalf("el error debe ser problem+json, llegó %q", rec.Header().Get("Content-Type"))
	}
	if strings.Contains(rec.Body.String(), "sin conexión") {
		t.Fatal("el error no debe exponer el detalle interno de la base")
	}
}

func TestReadyzSinBaseConfigurada(t *testing.T) {
	if rec := pedir(t, Dependencias{}, "/readyz"); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("se esperaba 503, llegó %d", rec.Code)
	}
}

// MF-05: la documentación interactiva no se publica en prod.
func TestDocsSoloFueraDeProduccion(t *testing.T) {
	if rec := pedir(t, Dependencias{PublicarDocs: true}, "/api/docs"); rec.Code != http.StatusOK {
		t.Fatalf("en dev /api/docs debe responder 200, llegó %d", rec.Code)
	}
	if rec := pedir(t, Dependencias{PublicarDocs: false}, "/api/docs"); rec.Code != http.StatusNotFound {
		t.Fatalf("en prod /api/docs debe responder 404, llegó %d", rec.Code)
	}
}

// MF-20: toda operación del contrato declara su permiso en x-permiso.
func TestCadaOperacionDeclaraSuPermiso(t *testing.T) {
	_, humaAPI := NuevoServidor(Dependencias{PublicarDocs: true})
	for ruta, item := range humaAPI.OpenAPI().Paths {
		operaciones := map[string]*huma.Operation{
			http.MethodGet: item.Get, http.MethodPost: item.Post, http.MethodPut: item.Put,
			http.MethodPatch: item.Patch, http.MethodDelete: item.Delete,
		}
		for metodo, op := range operaciones {
			if op != nil && op.Extensions[extensionPermiso] == nil {
				t.Errorf("%s %s no declara %s", metodo, ruta, extensionPermiso)
			}
		}
	}
}

func TestRequestIDSePropagaOSeGenera(t *testing.T) {
	servidor, _ := NuevoServidor(Dependencias{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(cabeceraRequestID, "abc-123")
	servidor.ServeHTTP(rec, req)
	if got := rec.Header().Get(cabeceraRequestID); got != "abc-123" {
		t.Fatalf("debe propagar el X-Request-Id recibido, llegó %q", got)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(cabeceraRequestID, "inválido con espacios")
	servidor.ServeHTTP(rec, req)
	if got := rec.Header().Get(cabeceraRequestID); len(got) != 32 {
		t.Fatalf("debe generar un id nuevo si el recibido es inválido, llegó %q", got)
	}
}
