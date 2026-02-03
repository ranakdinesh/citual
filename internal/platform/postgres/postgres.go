package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool represents the PostgreSQL connection pool
type Pool struct {
	*pgxpool.Pool
}

// NewPool creates and returns a new PostgreSQL connection pool
func NewPool(ctx context.Context) (*Pool, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse DATABASE_URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Ping the database to verify connection
	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")
	return &Pool{pool}, nil
}

// Close closes the connection pool
func (p *Pool) Close() {
	p.Pool.Close()
	fmt.Println("PostgreSQL connection pool closed.")
}
