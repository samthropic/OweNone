package seed

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samfiallos/owenone/backend/internal/auth"
)

//go:embed demo.sql
var demoSQL string

const DemoPassword = "password123"

var demoEmails = []string{
	"sarah@example.com",
	"jordan@example.com",
	"mike@example.com",
	"alex@example.com",
	"sam@example.com",
	"priya@example.com",
	"chris@example.com",
	"riley@example.com",
}

func Demo(ctx context.Context, pool *pgxpool.Pool) error {
	hash, err := auth.HashPassword(DemoPassword)
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}

	transaction, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin demo seed: %w", err)
	}
	defer transaction.Rollback(ctx) //nolint:errcheck

	if _, err := transaction.Exec(ctx, demoSQL); err != nil {
		return fmt.Errorf("execute demo seed: %w", err)
	}
	if _, err := transaction.Exec(ctx,
		"UPDATE users SET password_hash = $1 WHERE email = ANY($2)",
		hash, demoEmails,
	); err != nil {
		return fmt.Errorf("set demo passwords: %w", err)
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit demo seed: %w", err)
	}
	return nil
}
