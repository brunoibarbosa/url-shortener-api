//go:build test

// Package pg_test provides shared PostgreSQL test infrastructure
package pg_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	TestDB        *sql.DB
	TestContainer *postgres.PostgresContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create PostgreSQL container (shared across all PG tests)
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to start container: %v", err))
	}

	TestContainer = container

	// Cleanup guaranteed with defer
	defer func() {
		if TestDB != nil {
			TestDB.Close()
		}
		if TestContainer != nil {
			TestContainer.Terminate(context.Background())
		}
	}()

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("failed to get connection string: %v", err))
	}

	// Connect to database using database/sql (required by goose)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to connect: %v", err))
	}

	TestDB = db

	// Run real migrations from files
	if err := runMigrations(db); err != nil {
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}

	// Run all tests
	code := m.Run()

	os.Exit(code)
}

func runMigrations(db *sql.DB) error {
	// Get the absolute path to migrations directory using runtime.Caller
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get current file path")
	}

	// From: internal/infra/repository/pg/testing.go
	// To:   internal/infra/database/pg/migrations/
	baseDir := filepath.Dir(filename)
	migrationsDir := filepath.Join(baseDir, "..", "..", "database", "pg", "migrations")

	// Validate migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found: %s", migrationsDir)
	}

	// Set goose dialect
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Run all migrations up
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// BeginTestTx starts a transaction for test isolation.
// The transaction should be rolled back after the test to avoid cleanup overhead.
func BeginTestTx(t *testing.T) (*sql.Tx, func()) {
	t.Helper()

	tx, err := TestDB.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	cleanup := func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Errorf("failed to rollback transaction: %v", err)
		}
	}

	return tx, cleanup
}
