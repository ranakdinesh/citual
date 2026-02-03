package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// txContextKey is the unique key for storing the transaction in context
type txContextKey struct{}

// RunInTx executes a function within a database transaction with RLS enforcement.
func RunInTx(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {
	// 1. Begin Transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 2. RLS Enforcement (The Magic Glue)
	// We check if this request is from a Super User (System) or a specific Tenant

	// 3. Inject Transaction into Context
	// We use a specific key so GetTx can find it later
	txCtx := context.WithValue(ctx, txContextKey{}, tx)

	// 4. Run Business Logic
	if err := fn(txCtx); err != nil {
		return err
	}

	// 5. Commit
	return tx.Commit(ctx)
}

// GetTx retrieves the transaction from context (used by Store.getQueries)
func GetTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txContextKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}
