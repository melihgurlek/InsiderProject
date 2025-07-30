package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectDB establishes a connection pool to PostgreSQL using pgxpool.
// It returns a connected *pgxpool.Pool or an error.
func ConnectDB(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	// Configure connection pool settings
	config.MaxConns = 20                      // Maximum number of connections in the pool
	config.MinConns = 5                       // Minimum number of connections in the pool
	config.MaxConnLifetime = time.Hour        // Maximum lifetime of a connection
	config.MaxConnIdleTime = 30 * time.Minute // Maximum idle time of a connection
	config.HealthCheckPeriod = time.Minute    // How often to check connection health

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
