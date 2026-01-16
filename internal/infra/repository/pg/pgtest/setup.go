// Package pgtest provides shared PostgreSQL test infrastructure
package pgtest

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	TestDB        *pgxpool.Pool
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

	// Connect to database
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to connect: %v", err))
	}

	TestDB = pool

	// Run migrations
	if err := runMigrations(ctx); err != nil {
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}

	// Run all tests
	code := m.Run()

	os.Exit(code)
}

func runMigrations(ctx context.Context) error {
	// Create all tables needed for tests
	_, err := TestDB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email TEXT UNIQUE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ
		);

		CREATE TABLE IF NOT EXISTS user_profiles (
			id SERIAL PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name TEXT,
			avatar_url TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ
		);

		CREATE TABLE IF NOT EXISTS user_providers (
			id SERIAL PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			provider TEXT NOT NULL,
			provider_user_id TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(provider, provider_user_id)
		);

		CREATE TABLE IF NOT EXISTS urls (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			short_code TEXT NOT NULL UNIQUE,
			encrypted_url TEXT NOT NULL,
			user_id UUID REFERENCES users(id),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ,
			expires_at TIMESTAMPTZ,
			deleted_at TIMESTAMPTZ
		);

		CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			refresh_token_hash TEXT NOT NULL,
			user_agent TEXT,
			ip_address TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL,
			revoked_at TIMESTAMPTZ NULL
		);
	`)
	return err
}
