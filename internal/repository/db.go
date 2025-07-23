package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// ConnectDB establishes a connection to PostgreSQL using pgx.
// It returns a connected *pgx.Conn or an error.
func ConnectDB(ctx context.Context, dbURL string) (*pgx.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
