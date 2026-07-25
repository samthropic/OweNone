package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samfiallos/owenone/backend/migrations"
	"github.com/samfiallos/owenone/backend/seed"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	if err := migrations.Up(ctx, pool); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	if err := seed.Demo(ctx, pool); err != nil {
		return err
	}
	fmt.Println("demo data seeded")
	return nil
}
