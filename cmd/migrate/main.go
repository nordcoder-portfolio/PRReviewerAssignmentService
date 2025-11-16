package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return errors.New("DB_DSN is required")
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}

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

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	log.Printf("migrations applied successfully from %s", migrationsDir)
	return nil
}
