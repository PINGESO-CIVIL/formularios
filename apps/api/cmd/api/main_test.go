package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImprimirContratoGeneraOpenAPI31(t *testing.T) {
	var salida bytes.Buffer
	if err := ejecutar(context.Background(), []string{"openapi"}, &salida); err != nil {
		t.Fatal(err)
	}
	contrato := salida.String()
	for _, esperado := range []string{"openapi: 3.1", "/healthz", "/readyz", "x-permiso"} {
		if !strings.Contains(contrato, esperado) {
			t.Errorf("el contrato no contiene %q", esperado)
		}
	}
}

func TestEscribirContratoEnArchivo(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "openapi.yaml")
	if err := ejecutar(context.Background(), []string{"openapi", ruta}, nil); err != nil {
		t.Fatal(err)
	}
	contenido, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(contenido, []byte("components:")) && !bytes.Contains(contenido, []byte("openapi: 3.1")) {
		t.Fatalf("archivo inesperado: %.80s", contenido)
	}
}

// MF-02: sin configuración el servidor no arranca.
func TestLevantarServidorFallaSinConfiguracion(t *testing.T) {
	for _, v := range []string{"API_PORT", "API_ENV", "DATABASE_URL", "AUTH_JWKS_URL", "FRONTEND_URL", "TZ"} {
		t.Setenv(v, "")
	}
	err := ejecutar(context.Background(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "faltan variables") {
		t.Fatalf("se esperaba error de configuración, llegó %v", err)
	}
}
