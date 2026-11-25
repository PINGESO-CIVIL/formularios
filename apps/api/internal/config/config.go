// Paquete config carga y valida la configuración del servicio desde variables de entorno.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
)

// Configuracion reúne todas las variables de entorno que necesita el servicio (DAS §5).
type Configuracion struct {
	Puerto      string // API_PORT: puerto interno, siempre 8080
	Ambiente    string // API_ENV: dev, test o prod
	DatabaseURL string // DATABASE_URL: conexión a la base propia
	AuthJWKSURL string // AUTH_JWKS_URL: llaves públicas del servicio usuarios-roles
	FrontendURL string // FRONTEND_URL: origen permitido y base de enlaces
	ZonaHoraria string // TZ: America/Santiago
}

// ambientesValidos son los únicos valores aceptados para API_ENV.
var ambientesValidos = []string{"dev", "test", "prod"}

// Cargar lee las variables de entorno con os.Getenv.
func Cargar() (Configuracion, error) {
	return CargarDesde(os.Getenv)
}

// CargarDesde lee la configuración usando leer (os.Getenv en producción) y devuelve
// un error que nombra cada variable faltante o inválida. El proceso debe terminar si falla.
func CargarDesde(leer func(string) string) (Configuracion, error) {
	cfg := Configuracion{
		Puerto:      leer("API_PORT"),
		Ambiente:    leer("API_ENV"),
		DatabaseURL: leer("DATABASE_URL"),
		AuthJWKSURL: leer("AUTH_JWKS_URL"),
		FrontendURL: leer("FRONTEND_URL"),
		ZonaHoraria: leer("TZ"),
	}

	obligatorias := []struct{ nombre, valor string }{
		{"API_PORT", cfg.Puerto},
		{"API_ENV", cfg.Ambiente},
		{"DATABASE_URL", cfg.DatabaseURL},
		{"AUTH_JWKS_URL", cfg.AuthJWKSURL},
		{"FRONTEND_URL", cfg.FrontendURL},
		{"TZ", cfg.ZonaHoraria},
	}
	var faltantes []string
	for _, v := range obligatorias {
		if strings.TrimSpace(v.valor) == "" {
			faltantes = append(faltantes, v.nombre)
		}
	}
	if len(faltantes) > 0 {
		return Configuracion{}, fmt.Errorf("faltan variables de entorno obligatorias: %s", strings.Join(faltantes, ", "))
	}

	if !slices.Contains(ambientesValidos, cfg.Ambiente) {
		return Configuracion{}, errors.New("API_ENV debe ser dev, test o prod")
	}
	if !esURLHTTP(cfg.AuthJWKSURL) {
		return Configuracion{}, errors.New("AUTH_JWKS_URL debe ser una URL http o https")
	}
	if !esURLHTTP(cfg.FrontendURL) {
		return Configuracion{}, errors.New("FRONTEND_URL debe ser una URL http o https")
	}
	return cfg, nil
}

// EsProduccion indica si el servicio corre en producción, donde no se publica la documentación.
func (c Configuracion) EsProduccion() bool {
	return c.Ambiente == "prod"
}

// esURLHTTP indica si valor es una URL absoluta con esquema http o https.
func esURLHTTP(valor string) bool {
	u, err := url.Parse(valor)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
