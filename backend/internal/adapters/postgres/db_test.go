package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DiegoRaAr/ms-formularios-pingeso/backend/internal/adapters/postgres"
)

func TestNewPool_CreatesPoolForValidConfig(t *testing.T) {
	pool, err := postgres.NewPool(
		context.Background(),
		"postgres://formularios:formularios@localhost:5432/formularios?sslmode=disable",
	)
	require.NoError(t, err)
	require.NotNil(t, pool)
	pool.Close()
}

func TestNewPool_RejectsInvalidDatabaseURL(t *testing.T) {
	pool, err := postgres.NewPool(context.Background(), "://url-invalida")
	require.Error(t, err)
	require.Nil(t, pool)
}
