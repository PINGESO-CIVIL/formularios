package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool valida la cadena de conexión y crea un pool de conexiones hacia PostgreSQL.
// La conexión física se verifica bajo demanda (por ejemplo, en /readyz mediante pool.Ping).
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsear DATABASE_URL de postgres: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("crear pool de postgres: %w", err)
	}

	return pool, nil
}
