package pg_repo_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	url_domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	pg_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/url"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDB        *pgxpool.Pool
	testContainer *postgres.PostgresContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

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
		panic(err)
	}

	testContainer = container
	defer func() {
		if testDB != nil {
			testDB.Close()
		}
		if testContainer != nil {
			testContainer.Terminate(context.Background())
		}
	}()

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		panic(err)
	}

	testDB = pool

	if err := runMigrations(ctx); err != nil {
		panic(err)
	}

	code := m.Run()
	os.Exit(code)
}

func runMigrations(ctx context.Context) error {
	_, err := testDB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email TEXT UNIQUE,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ
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
	`)
	return err
}

func cleanDB(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.Exec(ctx, "TRUNCATE urls, users CASCADE")
	require.NoError(t, err)
}

func createTestUser(t *testing.T, ctx context.Context) uuid.UUID {
	var userID uuid.UUID
	err := testDB.QueryRow(ctx, "INSERT INTO users (email) VALUES ($1) RETURNING id", "test@example.com").Scan(&userID)
	require.NoError(t, err)
	return userID
}

func TestURLRepository_Exists_True(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	// Create URL
	url := &url_domain.URL{
		ShortCode:    "test123",
		EncryptedURL: "encrypted-data",
		UserID:       nil,
		ExpiresAt:    nil,
	}
	err := repo.Save(ctx, url)
	require.NoError(t, err)

	// Check if exists
	exists, err := repo.Exists(ctx, "test123")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestURLRepository_Exists_False(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	exists, err := repo.Exists(ctx, "nonexistent")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestURLRepository_Save_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	url := &url_domain.URL{
		ShortCode:    "abc123",
		EncryptedURL: "encrypted-url-data",
		UserID:       nil,
		ExpiresAt:    nil,
	}

	err := repo.Save(ctx, url)

	require.NoError(t, err)
}

func TestURLRepository_Save_WithUser(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewURLRepository(testDB)

	url := &url_domain.URL{
		ShortCode:    "user123",
		EncryptedURL: "encrypted-url",
		UserID:       &userID,
		ExpiresAt:    nil,
	}

	err := repo.Save(ctx, url)

	require.NoError(t, err)
}

func TestURLRepository_Save_WithExpiration(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	expiration := time.Now().Add(24 * time.Hour).UTC()
	expiresAt := &expiration
	url := &url_domain.URL{
		ShortCode:    "exp123",
		EncryptedURL: "encrypted-url",
		UserID:       nil,
		ExpiresAt:    expiresAt,
	}

	err := repo.Save(ctx, url)

	require.NoError(t, err)
}

func TestURLRepository_FindByShortCode_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	// Save URL
	url := &url_domain.URL{
		ShortCode:    "find123",
		EncryptedURL: "encrypted-test-data",
		UserID:       nil,
		ExpiresAt:    nil,
	}
	err := repo.Save(ctx, url)
	require.NoError(t, err)

	// Find URL
	found, err := repo.FindByShortCode(ctx, "find123")

	require.NoError(t, err)
	assert.Equal(t, "find123", found.ShortCode)
	assert.Equal(t, "encrypted-test-data", found.EncryptedURL)
	assert.Nil(t, found.UserID)
	assert.Nil(t, found.ExpiresAt)
	assert.Nil(t, found.DeletedAt)
}

func TestURLRepository_FindByShortCode_NotFound(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	found, err := repo.FindByShortCode(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestURLRepository_SoftDelete_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewURLRepository(testDB)

	// Create URL with user
	var urlID uuid.UUID
	err := testDB.QueryRow(ctx,
		"INSERT INTO urls (short_code, encrypted_url, user_id) VALUES ($1, $2, $3) RETURNING id",
		"del123", "encrypted-data", userID).Scan(&urlID)
	require.NoError(t, err)

	// Soft delete
	shortCode, err := repo.SoftDelete(ctx, urlID, userID)

	require.NoError(t, err)
	assert.Equal(t, "del123", shortCode)

	// Verify deleted
	found, err := repo.FindByShortCode(ctx, "del123")
	assert.Error(t, err) // Should not find deleted URLs
	assert.Nil(t, found)
}

func TestURLRepository_SoftDelete_AlreadyDeleted(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewURLRepository(testDB)

	// Create and soft delete URL
	var urlID uuid.UUID
	err := testDB.QueryRow(ctx,
		"INSERT INTO urls (short_code, encrypted_url, user_id, deleted_at) VALUES ($1, $2, $3, NOW()) RETURNING id",
		"del789", "encrypted-data", userID).Scan(&urlID)
	require.NoError(t, err)

	// Try to delete again
	shortCode, err := repo.SoftDelete(ctx, urlID, userID)

	assert.Error(t, err)
	assert.Empty(t, shortCode)
}

func TestURLRepository_DuplicateShortCode(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	// Save first URL
	url1 := &url_domain.URL{
		ShortCode:    "dup123",
		EncryptedURL: "encrypted-1",
	}
	err := repo.Save(ctx, url1)
	require.NoError(t, err)

	// Try to save with same short code
	url2 := &url_domain.URL{
		ShortCode:    "dup123",
		EncryptedURL: "encrypted-2",
	}
	err = repo.Save(ctx, url2)

	assert.Error(t, err) // Should fail due to unique constraint
}

func TestURLRepository_MultipleURLs(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	// Create multiple URLs
	for i := 0; i < 5; i++ {
		url := &url_domain.URL{
			ShortCode:    fmt.Sprintf("multi%d", i),
			EncryptedURL: fmt.Sprintf("encrypted-%d", i),
		}
		err := repo.Save(ctx, url)
		require.NoError(t, err)
	}

	// Verify all exist
	for i := 0; i < 5; i++ {
		exists, err := repo.Exists(ctx, fmt.Sprintf("multi%d", i))
		require.NoError(t, err)
		assert.True(t, exists)
	}
}

func TestURLRepository_ExpiredURL(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	// Create expired URL
	expiration := time.Now().Add(-1 * time.Hour).UTC()
	expiresAt := &expiration
	url := &url_domain.URL{
		ShortCode:    "expired",
		EncryptedURL: "encrypted-url",
		ExpiresAt:    expiresAt,
	}
	err := repo.Save(ctx, url)
	require.NoError(t, err)

	// Can still find expired URL (expiration handled by application layer)
	found, err := repo.FindByShortCode(ctx, "expired")
	require.NoError(t, err)
	assert.NotNil(t, found.ExpiresAt)
	assert.True(t, found.ExpiresAt.Before(time.Now()))
}

func TestURLRepository_FindByShortCode_WithUser(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewURLRepository(testDB)

	// Save URL with user
	url := &url_domain.URL{
		ShortCode:    "userurl",
		EncryptedURL: "encrypted-user-data",
		UserID:       &userID,
	}
	err := repo.Save(ctx, url)
	require.NoError(t, err)

	// Find URL
	found, err := repo.FindByShortCode(ctx, "userurl")

	require.NoError(t, err)
	assert.Equal(t, "userurl", found.ShortCode)
	assert.NotNil(t, found.UserID)
	assert.Equal(t, userID, *found.UserID)
}

func TestURLRepository_Exists_Deleted(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewURLRepository(testDB)

	// Create and delete URL
	url := &url_domain.URL{
		ShortCode:    "deleted123",
		EncryptedURL: "encrypted-data",
		UserID:       &userID,
	}
	err := repo.Save(ctx, url)
	require.NoError(t, err)

	var urlID uuid.UUID
	err = testDB.QueryRow(ctx, "SELECT id FROM urls WHERE short_code = $1", "deleted123").Scan(&urlID)
	require.NoError(t, err)

	_, err = repo.SoftDelete(ctx, urlID, userID)
	require.NoError(t, err)

	// Exists should still return true (deleted URLs still exist in DB)
	exists, err := repo.Exists(ctx, "deleted123")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestURLRepository_Save_NilFields(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	url := &url_domain.URL{
		ShortCode:    "minimal",
		EncryptedURL: "encrypted",
		UserID:       nil,
		ExpiresAt:    nil,
	}

	err := repo.Save(ctx, url)
	require.NoError(t, err)

	// Verify it was saved
	exists, err := repo.Exists(ctx, "minimal")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestURLRepository_FindByShortCode_Deleted(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewURLRepository(testDB)

	// Create URL
	url := &url_domain.URL{
		ShortCode:    "todelete",
		EncryptedURL: "encrypted-data",
		UserID:       &userID,
	}
	err := repo.Save(ctx, url)
	require.NoError(t, err)

	var urlID uuid.UUID
	err = testDB.QueryRow(ctx, "SELECT id FROM urls WHERE short_code = $1", "todelete").Scan(&urlID)
	require.NoError(t, err)

	// Delete it
	_, err = repo.SoftDelete(ctx, urlID, userID)
	require.NoError(t, err)

	// FindByShortCode should not find deleted URLs
	found, err := repo.FindByShortCode(ctx, "todelete")
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestURLRepository_SoftDelete_WrongUser(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	// Create second user
	var otherUserID uuid.UUID
	err := testDB.QueryRow(ctx, "INSERT INTO users (email) VALUES ($1) RETURNING id", "other@example.com").Scan(&otherUserID)
	require.NoError(t, err)

	repo := pg_repo.NewURLRepository(testDB)

	// Create URL with first user
	url := &url_domain.URL{
		ShortCode:    "useronly",
		EncryptedURL: "encrypted-data",
		UserID:       &userID,
	}
	err = repo.Save(ctx, url)
	require.NoError(t, err)

	var urlID uuid.UUID
	err = testDB.QueryRow(ctx, "SELECT id FROM urls WHERE short_code = $1", "useronly").Scan(&urlID)
	require.NoError(t, err)

	// Try to delete with different user
	shortCode, err := repo.SoftDelete(ctx, urlID, otherUserID)
	assert.Error(t, err)
	assert.Empty(t, shortCode)
}

func TestURLRepository_SoftDelete_NonExistent(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewURLRepository(testDB)

	shortCode, err := repo.SoftDelete(ctx, uuid.New(), userID)
	assert.Error(t, err)
	assert.Empty(t, shortCode)
}

func TestURLRepository_Save_VeryLongURL(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	// Very long encrypted URL
	// Create a very long URL (5000 characters)
	longURL := strings.Repeat("A", 5000)

	url := &url_domain.URL{
		ShortCode:    "longurl",
		EncryptedURL: longURL,
	}

	err := repo.Save(ctx, url)
	require.NoError(t, err)

	// Verify it can be retrieved
	found, err := repo.FindByShortCode(ctx, "longurl")
	require.NoError(t, err)
	assert.NotNil(t, found)
}

func TestURLRepository_FindByShortCode_SpecialCharacters(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	// ShortCode with special characters (URL-safe)
	specialCode := "abc-123_XYZ"

	url := &url_domain.URL{
		ShortCode:    specialCode,
		EncryptedURL: "encrypted-data",
	}
	err := repo.Save(ctx, url)
	require.NoError(t, err)

	found, err := repo.FindByShortCode(ctx, specialCode)
	require.NoError(t, err)
	assert.Equal(t, specialCode, found.ShortCode)
}

func TestURLRepository_Save_WithExpiredDate(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	// URL that's already expired
	expiresAt := time.Now().Add(-24 * time.Hour)
	url := &url_domain.URL{
		ShortCode:    "expired",
		EncryptedURL: "encrypted-data",
		ExpiresAt:    &expiresAt,
	}

	err := repo.Save(ctx, url)
	require.NoError(t, err)

	// Should still be able to find it (expiration logic is in app layer)
	found, err := repo.FindByShortCode(ctx, "expired")
	require.NoError(t, err)
	assert.NotNil(t, found.ExpiresAt)
	assert.True(t, found.ExpiresAt.Before(time.Now()))
}

func TestURLRepository_Exists_CaseSensitive(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	// Save with lowercase
	url := &url_domain.URL{
		ShortCode:    "abc123",
		EncryptedURL: "encrypted-data",
	}
	err := repo.Save(ctx, url)
	require.NoError(t, err)

	// Check with different case
	exists1, err := repo.Exists(ctx, "abc123")
	require.NoError(t, err)
	assert.True(t, exists1)

	exists2, err := repo.Exists(ctx, "ABC123")
	require.NoError(t, err)
	// Depends on DB collation, but typically case-sensitive
	assert.False(t, exists2)
}

func TestURLRepository_SoftDelete_CheckDeletedAt(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewURLRepository(testDB)

	url := &url_domain.URL{
		ShortCode:    "checkdelete",
		EncryptedURL: "encrypted-data",
		UserID:       &userID,
	}
	err := repo.Save(ctx, url)
	require.NoError(t, err)

	var urlID uuid.UUID
	err = testDB.QueryRow(ctx, "SELECT id FROM urls WHERE short_code = $1", "checkdelete").Scan(&urlID)
	require.NoError(t, err)

	// Soft delete
	shortCode, err := repo.SoftDelete(ctx, urlID, userID)
	require.NoError(t, err)
	assert.Equal(t, "checkdelete", shortCode)

	// Verify deleted_at is set
	var deletedAt *time.Time
	err = testDB.QueryRow(ctx, "SELECT deleted_at FROM urls WHERE id = $1", urlID).Scan(&deletedAt)
	require.NoError(t, err)
	assert.NotNil(t, deletedAt)
	assert.True(t, deletedAt.Before(time.Now().Add(1*time.Second)))
}

func TestURLRepository_MultipleURLsSameUser(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewURLRepository(testDB)

	// Create multiple URLs for same user
	for i := 0; i < 5; i++ {
		url := &url_domain.URL{
			ShortCode:    fmt.Sprintf("user-url-%d", i),
			EncryptedURL: fmt.Sprintf("encrypted-%d", i),
			UserID:       &userID,
		}
		err := repo.Save(ctx, url)
		require.NoError(t, err)
	}

	// Verify all exist
	for i := 0; i < 5; i++ {
		exists, err := repo.Exists(ctx, fmt.Sprintf("user-url-%d", i))
		require.NoError(t, err)
		assert.True(t, exists)
	}
}

// Transaction Tests - Cover BaseRepository.Q()

func TestURLRepository_WithTransaction_Commit(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	// Create user first
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "txurl1@test.com")
	require.NoError(t, err)

	// Execute operations within transaction
	err = txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		url1 := &url_domain.URL{
			UserID:       &userID,
			ShortCode:    "tx1",
			EncryptedURL: "https://tx-commit-1.com",
		}
		if err := repo.Save(txCtx, url1); err != nil {
			return err
		}

		url2 := &url_domain.URL{
			UserID:       &userID,
			ShortCode:    "tx2",
			EncryptedURL: "https://tx-commit-2.com",
		}
		return repo.Save(txCtx, url2)
	})

	require.NoError(t, err)

	// Verify both URLs were committed
	found1, err := repo.FindByShortCode(ctx, "tx1")
	require.NoError(t, err)
	assert.Equal(t, "https://tx-commit-1.com", found1.EncryptedURL)

	found2, err := repo.FindByShortCode(ctx, "tx2")
	require.NoError(t, err)
	assert.Equal(t, "https://tx-commit-2.com", found2.EncryptedURL)
}

func TestURLRepository_WithTransaction_Rollback(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	// Create user first
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "txurl2@test.com")
	require.NoError(t, err)

	// Execute operations within transaction that fails
	err = txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		url := &url_domain.URL{
			UserID:       &userID,
			ShortCode:    "rollback",
			EncryptedURL: "https://rollback.com",
		}
		if err := repo.Save(txCtx, url); err != nil {
			return err
		}

		// Force rollback
		return errors.New("rollback test")
	})

	require.Error(t, err)

	// Verify URL was NOT persisted
	exists, err := repo.Exists(ctx, "rollback")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestURLRepository_WithTransaction_MultipleCreates(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	// Create user first
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "txurl3@test.com")
	require.NoError(t, err)

	// Execute multiple operations within transaction
	err = txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Create first URL
		url1 := &url_domain.URL{
			UserID:       &userID,
			ShortCode:    "multi1",
			EncryptedURL: "https://multiple-1.com",
		}
		if err := repo.Save(txCtx, url1); err != nil {
			return err
		}

		// Create second URL
		url2 := &url_domain.URL{
			UserID:       &userID,
			ShortCode:    "multi2",
			EncryptedURL: "https://multiple-2.com",
		}
		if err := repo.Save(txCtx, url2); err != nil {
			return err
		}

		// Check existence within transaction
		exists, err := repo.Exists(txCtx, "multi1")
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("url1 should exist")
		}

		return nil
	})

	require.NoError(t, err)

	// Verify both URLs were committed
	found1, err := repo.FindByShortCode(ctx, "multi1")
	require.NoError(t, err)
	assert.Equal(t, "https://multiple-1.com", found1.EncryptedURL)

	found2, err := repo.FindByShortCode(ctx, "multi2")
	require.NoError(t, err)
	assert.Equal(t, "https://multiple-2.com", found2.EncryptedURL)
}

// High Priority Tests - Database Constraints

func TestURLRepository_ShortCode_UniqueConstraint(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	url1 := &url_domain.URL{
		ShortCode:    "duplicate",
		EncryptedURL: "https://first.com",
	}

	err := repo.Save(ctx, url1)
	require.NoError(t, err)

	// Try to create another URL with same shortcode
	url2 := &url_domain.URL{
		ShortCode:    "duplicate",
		EncryptedURL: "https://second.com",
	}

	err = repo.Save(ctx, url2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate key")
}

func TestURLRepository_UserID_ForeignKeyConstraint(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	// Try to create URL with non-existent user
	nonExistentUserID := uuid.New()
	url := &url_domain.URL{
		ShortCode:    "test123",
		EncryptedURL: "https://test.com",
		UserID:       &nonExistentUserID,
	}

	err := repo.Save(ctx, url)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "violates foreign key constraint")
}

func TestURLRepository_Save_ValidatesTimestamps(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewURLRepository(testDB)

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	url := &url_domain.URL{
		ShortCode:    "timestamp",
		EncryptedURL: "https://test.com",
		ExpiresAt:    &expiresAt,
	}

	err := repo.Save(ctx, url)
	require.NoError(t, err)

	// Verify timestamps
	found, err := repo.FindByShortCode(ctx, "timestamp")
	require.NoError(t, err)

	assert.NotNil(t, found.ExpiresAt)
	assert.True(t, found.ExpiresAt.After(time.Now()), "ExpiresAt should be in the future")
	assert.Nil(t, found.DeletedAt, "DeletedAt should be nil for new URL")
}

func TestURLRepository_FindByShortCode_ContextCanceled(t *testing.T) {
	cleanDB(t)

	repo := pg_repo.NewURLRepository(testDB)

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.FindByShortCode(ctx, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestURLRepository_Transaction_PersistenceValidation(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// Create user first
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "txurl@test.com")
	require.NoError(t, err)

	repo := pg_repo.NewURLRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	// Execute within transaction
	err = txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		url := &url_domain.URL{
			UserID:       &userID,
			ShortCode:    "txpersist",
			EncryptedURL: "https://tx-persist.com",
		}
		return repo.Save(txCtx, url)
	})

	require.NoError(t, err)

	// Verify persistence OUTSIDE transaction with fresh context
	freshCtx := context.Background()
	found, err := repo.FindByShortCode(freshCtx, "txpersist")
	require.NoError(t, err, "URL should persist after transaction commit")
	assert.Equal(t, "https://tx-persist.com", found.EncryptedURL)
	assert.Equal(t, userID, *found.UserID)
}

func TestURLRepository_SoftDelete_ChecksDeletedAt(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// Create user
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "delete@test.com")
	require.NoError(t, err)

	repo := pg_repo.NewURLRepository(testDB)

	// Create URL
	url := &url_domain.URL{
		ShortCode:    "softdel",
		EncryptedURL: "https://todelete.com",
		UserID:       &userID,
	}

	err = repo.Save(ctx, url)
	require.NoError(t, err)

	// Get ID from database
	var urlID uuid.UUID
	err = testDB.QueryRow(ctx, "SELECT id FROM urls WHERE short_code = $1", "softdel").Scan(&urlID)
	require.NoError(t, err)

	// Soft delete
	beforeDelete := time.Now().Add(-1 * time.Second) // Add 1 second buffer
	_, err = repo.SoftDelete(ctx, urlID, userID)
	require.NoError(t, err)
	afterDelete := time.Now().Add(1 * time.Second) // Add 1 second buffer

	// Verify DeletedAt is set
	var deletedAt *time.Time
	err = testDB.QueryRow(ctx, "SELECT deleted_at FROM urls WHERE id = $1", urlID).Scan(&deletedAt)
	require.NoError(t, err)
	require.NotNil(t, deletedAt, "DeletedAt should be set after soft delete")
	assert.True(t, deletedAt.After(beforeDelete) && deletedAt.Before(afterDelete), "DeletedAt should be between deletion time")
}

// TestURLRepository_Concurrent_SaveURLs tests concurrent URL creation
// to verify thread safety and detect potential race conditions
func TestURLRepository_Concurrent_SaveURLs(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// Create test user
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "concurrent@test.com")
	require.NoError(t, err)

	repo := pg_repo.NewURLRepository(testDB)

	const numGoroutines = 10
	errors := make(chan error, numGoroutines)
	shortCodes := make(chan string, numGoroutines)

	// Create URLs concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			url := &url_domain.URL{
				ShortCode:    fmt.Sprintf("concurrent%d", index),
				EncryptedURL: fmt.Sprintf("https://concurrent%d.com", index),
				UserID:       &userID,
			}

			err := repo.Save(ctx, url)
			if err != nil {
				errors <- err
				return
			}

			shortCodes <- url.ShortCode
			errors <- nil
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		require.NoError(t, err, "Concurrent save should not fail")
	}

	close(shortCodes)
	uniqueCodes := make(map[string]bool)
	for code := range shortCodes {
		uniqueCodes[code] = true
	}

	require.Equal(t, numGoroutines, len(uniqueCodes), "All URLs should have unique short codes")
}

// TestURLRepository_FindByShortCode_ContextTimeout tests that
// FindByShortCode properly respects context timeout
func TestURLRepository_FindByShortCode_ContextTimeout(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// Create test user
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "timeout@test.com")
	require.NoError(t, err)

	repo := pg_repo.NewURLRepository(testDB)

	url := &url_domain.URL{
		ShortCode:    "timeouttest",
		EncryptedURL: "https://timeout.com",
		UserID:       &userID,
	}

	err = repo.Save(ctx, url)
	require.NoError(t, err)

	// Create context with very short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout has passed

	_, err = repo.FindByShortCode(timeoutCtx, "timeouttest")
	require.Error(t, err)
	require.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled),
		"Error should be context deadline exceeded or canceled")
}

// TestURLRepository_Save_ValidatesCreatedAt verifies that CreatedAt timestamp
// is properly set when saving a URL
func TestURLRepository_Save_ValidatesCreatedAt(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// Create test user
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "created-at@test.com")
	require.NoError(t, err)

	repo := pg_repo.NewURLRepository(testDB)

	beforeCreate := time.Now().Add(-1 * time.Second)
	url := &url_domain.URL{
		ShortCode:    "createdtest",
		EncryptedURL: "https://created.com",
		UserID:       &userID,
	}

	err = repo.Save(ctx, url)
	require.NoError(t, err)
	afterCreate := time.Now().Add(1 * time.Second)

	// Verify CreatedAt is set
	var createdAt time.Time
	err = testDB.QueryRow(ctx, "SELECT created_at FROM urls WHERE short_code = $1", "createdtest").Scan(&createdAt)
	require.NoError(t, err)
	require.NotZero(t, createdAt, "CreatedAt should be set")
	assert.True(t, createdAt.After(beforeCreate) && createdAt.Before(afterCreate),
		"CreatedAt should be between save operation time")
}

// TestURLRepository_Save_TableDriven tests various save scenarios using table-driven approach
func TestURLRepository_Save_TableDriven(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// Create test user
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "table-driven@test.com")
	require.NoError(t, err)

	repo := pg_repo.NewURLRepository(testDB)

	tests := []struct {
		name         string
		shortCode    string
		encryptedURL string
		userID       *uuid.UUID
		wantErr      bool
	}{
		{
			name:         "valid URL with user",
			shortCode:    "table1",
			encryptedURL: "https://example.com/1",
			userID:       &userID,
			wantErr:      false,
		},
		{
			name:         "valid URL without user",
			shortCode:    "table2",
			encryptedURL: "https://example.com/2",
			userID:       nil,
			wantErr:      false,
		},
		{
			name:         "valid URL with long encrypted URL",
			shortCode:    "table3",
			encryptedURL: "https://example.com/very/long/path/with/many/segments/1/2/3/4/5",
			userID:       &userID,
			wantErr:      false,
		},
		{
			name:         "valid URL with special characters",
			shortCode:    "table4",
			encryptedURL: "https://example.com/path?param=value&other=123",
			userID:       &userID,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := &url_domain.URL{
				ShortCode:    tt.shortCode,
				EncryptedURL: tt.encryptedURL,
				UserID:       tt.userID,
			}

			err := repo.Save(ctx, url)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// BenchmarkURLRepository_Save benchmarks URL save performance
func BenchmarkURLRepository_Save(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	_, _ = testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "bench-url@example.com")

	repo := pg_repo.NewURLRepository(testDB)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := &url_domain.URL{
			ShortCode:    fmt.Sprintf("bench%d", i),
			EncryptedURL: fmt.Sprintf("https://example.com/%d", i),
			UserID:       &userID,
		}
		_ = repo.Save(ctx, url)
	}
}

// BenchmarkURLRepository_FindByShortCode benchmarks URL lookup performance
func BenchmarkURLRepository_FindByShortCode(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	_, _ = testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "bench-find-url@example.com")

	repo := pg_repo.NewURLRepository(testDB)

	// Create test URL
	url := &url_domain.URL{
		ShortCode:    "benchfind",
		EncryptedURL: "https://example.com/bench",
		UserID:       &userID,
	}
	_ = repo.Save(ctx, url)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByShortCode(ctx, "benchfind")
	}
}
