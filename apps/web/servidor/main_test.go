package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// prepararRaiz crea un directorio con index.html y un asset.
func prepararRaiz(t *testing.T) string {
	t.Helper()
	raiz := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raiz, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raiz, "index.html"), []byte("<html>spa</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raiz, "assets", "app.js"), []byte("console.log(1)"), 0o644); err != nil {
		t.Fatal(err)
	}
	return raiz
}

func pedir(t *testing.T, h http.Handler, ruta string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
	return rec
}

func TestHealthzYCabecerasSeguras(t *testing.T) {
	h, err := nuevoManejador(prepararRaiz(t), "http://backend:8080")
	if err != nil {
		t.Fatal(err)
	}
	rec := pedir(t, h, "/healthz")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"estado":"ok"`) {
		t.Fatalf("healthz inesperado: %d %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" || rec.Header().Get("Referrer-Policy") == "" {
		t.Fatal("faltan cabeceras de seguridad")
	}
}

func TestAssetsInmutablesYRutasDeLaSPA(t *testing.T) {
	h, err := nuevoManejador(prepararRaiz(t), "http://backend:8080")
	if err != nil {
		t.Fatal(err)
	}
	if rec := pedir(t, h, "/assets/app.js"); !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("los assets deben ser inmutables, llegó %q", rec.Header().Get("Cache-Control"))
	}
	rec := pedir(t, h, "/formularios/123")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "spa") {
		t.Fatalf("una ruta de la SPA debe devolver index.html, llegó %d %s", rec.Code, rec.Body.String())
	}
}

func TestReenviaApiAlBackend(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("backend:" + r.URL.Path))
	}))
	defer backend.Close()

	h, err := nuevoManejador(prepararRaiz(t), backend.URL)
	if err != nil {
		t.Fatal(err)
	}
	if rec := pedir(t, h, "/api/formularios"); rec.Body.String() != "backend:/api/formularios" {
		t.Fatalf("no se reenvió al backend: %s", rec.Body.String())
	}
}

func TestBackendURLInvalida(t *testing.T) {
	if _, err := nuevoManejador(t.TempDir(), "backend-sin-esquema"); err == nil {
		t.Fatal("se esperaba error con una BACKEND_URL relativa")
	}
}
