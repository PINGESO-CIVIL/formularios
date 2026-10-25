package config

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func envFrom(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoad_AppliesDefaults(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{"DATABASE_URL": "postgres://x"}))
	require.NoError(t, err)

	assert.Equal(t, EnvDev, cfg.Env)
	assert.Equal(t, ":8080", cfg.HTTPAddr)
	assert.Equal(t, slog.LevelInfo, cfg.LogLevel)
	assert.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
}

func TestLoad_RejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{"sin DATABASE_URL", map[string]string{}},
		{"APP_ENV desconocido", map[string]string{"DATABASE_URL": "x", "APP_ENV": "staging"}},
		{"LOG_LEVEL inválido", map[string]string{"DATABASE_URL": "x", "LOG_LEVEL": "verbose"}},
		{"SHUTDOWN_TIMEOUT inválido", map[string]string{"DATABASE_URL": "x", "SHUTDOWN_TIMEOUT": "pronto"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(envFrom(tt.env))
			assert.Error(t, err)
		})
	}
}
