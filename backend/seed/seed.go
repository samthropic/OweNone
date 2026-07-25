package seed

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed demo.sql
var demoSQL string

func Demo(ctx context.Context, pool *pgxpool.Pool) error {
	transaction, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin demo seed: %w", err)
	}
	defer transaction.Rollback(ctx) //nolint:errcheck

	if _, err := transaction.Exec(ctx, demoSQL); err != nil {
		return fmt.Errorf("execute demo seed: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit demo seed: %w", err)
	}
	return nil
}
