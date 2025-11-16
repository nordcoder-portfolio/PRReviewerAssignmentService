package postgrestest

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // for tests (lint)
	"github.com/pressly/goose/v3"
)

var TestPool *pgxpool.Pool

func TestMain(m *testing.M) {
	code := Run(m)
	os.Exit(code)
}

func Run(m *testing.M) int {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = os.Getenv("DB_DSN")
	}

	if dsn == "" {
		log.Println("set dsn")
		return m.Run()
	}

	if err := runMigrations(dsn); err != nil {
		log.Printf("migrations failed: %v\n", err)
		return 1
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Printf("pgxpool.New failed: %v\n", err)
		return 1
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Printf("db ping failed: %v\n", err)
		return 1
	}

	TestPool = pool

	return m.Run()
}

func runMigrations(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

func ResetDB(t *testing.T, tables ...string) {
	t.Helper()

	if TestPool == nil {
		t.Skip("postgres test pool is not initialized, skipping")
		return
	}

	if len(tables) == 0 {
		tables = []string{
			"pr_reviewers",
			"pull_requests",
			"users",
			"teams",
		}
	}

	ctx := t.Context()
	query := fmt.Sprintf(
		"TRUNCATE %s RESTART IDENTITY CASCADE;",
		strings.Join(tables, ", "),
	)

	if _, err := TestPool.Exec(ctx, query); err != nil {
		t.Fatalf("truncate tables %v: %v", tables, err)
	}
}
