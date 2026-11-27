// Package config carga la configuración del servicio desde variables de entorno.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Env string

const (
	EnvDev  Env = "dev"
	EnvTest Env = "test"
	EnvProd Env = "prod"
)

type Config struct {
	Env             Env
	HTTPAddr        string
	LogLevel        slog.Level
	DatabaseURL     string
	ShutdownTimeout time.Duration
}

// Load lee la configuración desde el entorno usando getenv (os.Getenv en producción).
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Env:             Env(withDefault(getenv("APP_ENV"), string(EnvDev))),
		HTTPAddr:        withDefault(getenv("HTTP_ADDR"), ":8080"),
		DatabaseURL:     getenv("DATABASE_URL"),
		ShutdownTimeout: 10 * time.Second,
	}

	var errs []error

	switch cfg.Env {
	case EnvDev, EnvTest, EnvProd:
	default:
		errs = append(errs, fmt.Errorf("APP_ENV inválido: %q (dev|test|prod)", cfg.Env))
	}

	if err := cfg.LogLevel.UnmarshalText([]byte(withDefault(getenv("LOG_LEVEL"), "info"))); err != nil {
		errs = append(errs, fmt.Errorf("LOG_LEVEL inválido: %w", err))
	}

	if raw := getenv("SHUTDOWN_TIMEOUT"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT inválido: %w", err))
		}
		cfg.ShutdownTimeout = d
	}

	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL es obligatorio"))
	}

	return cfg, errors.Join(errs...)
}

// FromOS es un atajo para Load(os.Getenv).
func FromOS() (Config, error) { return Load(os.Getenv) }

func withDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
