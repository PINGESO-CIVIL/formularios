package config

import (
	"strings"
	"testing"
)

// entornoValido devuelve un conjunto completo de variables válidas.
func entornoValido() map[string]string {
	return map[string]string{
		"API_PORT":      "8080",
		"API_ENV":       "dev",
		"DATABASE_URL":  "postgres://u:c@localhost:5432/formularios",
		"AUTH_JWKS_URL": "http://localhost:5161/.well-known/jwks.json",
		"FRONTEND_URL":  "http://localhost:5140",
		"TZ":            "America/Santiago",
	}
}

func lector(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestCargarConEntornoValido(t *testing.T) {
	cfg, err := CargarDesde(lector(entornoValido()))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if cfg.Puerto != "8080" || cfg.Ambiente != "dev" || cfg.EsProduccion() {
		t.Fatalf("configuración inesperada: %+v", cfg)
	}
}

// MF-02: el proceso no arranca si falta una variable y el error la nombra.
func TestCargarFallaSiFaltanVariables(t *testing.T) {
	for nombre := range entornoValido() {
		t.Run(nombre, func(t *testing.T) {
			env := entornoValido()
			delete(env, nombre)
			_, err := CargarDesde(lector(env))
			if err == nil || !strings.Contains(err.Error(), nombre) {
				t.Fatalf("se esperaba un error que nombrara %s, llegó %v", nombre, err)
			}
		})
	}
}

func TestCargarFallaConValoresInvalidos(t *testing.T) {
	casos := map[string][2]string{
		"ambiente desconocido":   {"API_ENV", "staging"},
		"jwks sin esquema":       {"AUTH_JWKS_URL", "localhost:5161/jwks"},
		"frontend con otro tipo": {"FRONTEND_URL", "ftp://localhost"},
	}
	for nombre, caso := range casos {
		t.Run(nombre, func(t *testing.T) {
			env := entornoValido()
			env[caso[0]] = caso[1]
			if _, err := CargarDesde(lector(env)); err == nil {
				t.Fatalf("se esperaba error con %s=%s", caso[0], caso[1])
			}
		})
	}
}

func TestEsProduccion(t *testing.T) {
	env := entornoValido()
	env["API_ENV"] = "prod"
	cfg, err := CargarDesde(lector(env))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.EsProduccion() {
		t.Fatal("con API_ENV=prod EsProduccion debe ser true")
	}
}
