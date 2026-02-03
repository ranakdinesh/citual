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
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// --- THE MAGIC GLUE (RLS) ---
	// Extract Tenant ID from context (assuming you have a domain.GetTenantID(ctx))
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		// Enforce isolation at the Postgres level
		// All subsequent queries in this TX will fail if they try to touch other tenants' rows
		_, err := tx.Exec(ctx, "SELECT set_config('citual.current_tenant', $1, true)", tenantID)
		if err != nil {
			return fmt.Errorf("failed to set tenant context: %w", err)
		}
	}
	// ----------------------------

	txCtx := context.WithValue(ctx, txContextKey{}, tx)

	if err := fn(txCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetTx retrieves the transaction from context (used by Store.getQueries)
func GetTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txContextKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}
