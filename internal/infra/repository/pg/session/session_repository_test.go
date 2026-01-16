package pg_repo_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	pg_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/session"
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

func cleanDB(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.Exec(ctx, "TRUNCATE sessions, users CASCADE")
	require.NoError(t, err)
}

func createTestUser(t *testing.T, ctx context.Context) uuid.UUID {
	var userID uuid.UUID
	err := testDB.QueryRow(ctx, "INSERT INTO users (email) VALUES ($1) RETURNING id", "test@example.com").Scan(&userID)
	require.NoError(t, err)
	return userID
}

func TestSessionRepository_Create_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "hash123",
		UserAgent:        "Mozilla/5.0",
		IPAddress:        "192.168.1.1",
		ExpiresAt:        &expiresAt,
	}

	err := repo.Create(ctx, session)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, session.ID)
}

func TestSessionRepository_FindByRefreshToken_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)
	hash := "test-hash-123"

	// Create session
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: hash,
		UserAgent:        "Mozilla/5.0",
		IPAddress:        "192.168.1.1",
		ExpiresAt:        &expiresAt,
	}
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Find by refresh token
	found, err := repo.FindByRefreshToken(ctx, hash)

	require.NoError(t, err)
	assert.Equal(t, session.ID, found.ID)
	assert.Equal(t, userID, found.UserID)
	assert.Equal(t, hash, found.RefreshTokenHash)
	assert.Equal(t, "Mozilla/5.0", found.UserAgent)
	assert.Equal(t, "192.168.1.1", found.IPAddress)
	assert.Nil(t, found.RevokedAt)
}

func TestSessionRepository_FindByRefreshToken_NotFound(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewSessionRepository(testDB)

	found, err := repo.FindByRefreshToken(ctx, "nonexistent-hash")

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrNotFound, err)
	assert.Nil(t, found)
}

func TestSessionRepository_Revoke_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)

	// Create session
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "hash-to-revoke",
		UserAgent:        "Mozilla/5.0",
		IPAddress:        "192.168.1.1",
		ExpiresAt:        &expiresAt,
	}
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Revoke session
	err = repo.Revoke(ctx, session.ID)
	require.NoError(t, err)

	// Verify revoked
	found, err := repo.FindByRefreshToken(ctx, "hash-to-revoke")
	require.NoError(t, err)
	assert.NotNil(t, found.RevokedAt)
}

func TestSessionRepository_Revoke_NonexistentSession(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewSessionRepository(testDB)

	// Revoke nonexistent session (should not error)
	err := repo.Revoke(ctx, uuid.New())

	assert.NoError(t, err)
}

func TestSessionRepository_MultipleSessionsPerUser(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		expiresAt := time.Now().Add(24 * time.Hour)
		session := &session_domain.Session{
			UserID:           userID,
			RefreshTokenHash: fmt.Sprintf("hash-%d", i),
			UserAgent:        "Mozilla/5.0",
			IPAddress:        "192.168.1.1",
			ExpiresAt:        &expiresAt,
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
	}

	// Verify all sessions can be found
	for i := 0; i < 3; i++ {
		found, err := repo.FindByRefreshToken(ctx, fmt.Sprintf("hash-%d", i))
		require.NoError(t, err)
		assert.Equal(t, userID, found.UserID)
	}
}

func TestSessionRepository_ExpiredSession(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)

	// Create expired session
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "expired-hash",
		UserAgent:        "Mozilla/5.0",
		IPAddress:        "192.168.1.1",
		ExpiresAt:        &expiresAt,
	}
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Can still find expired session (expiration logic handled in application layer)
	found, err := repo.FindByRefreshToken(ctx, "expired-hash")
	require.NoError(t, err)
	assert.True(t, found.ExpiresAt.Before(time.Now()))
}

func TestSessionRepository_RevokedSession(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)

	// Create and revoke session
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "revoked-hash",
		UserAgent:        "Mozilla/5.0",
		IPAddress:        "192.168.1.1",
		ExpiresAt:        &expiresAt,
	}
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	err = repo.Revoke(ctx, session.ID)
	require.NoError(t, err)

	// Find revoked session
	found, err := repo.FindByRefreshToken(ctx, "revoked-hash")
	require.NoError(t, err)
	assert.NotNil(t, found.RevokedAt)
	assert.True(t, found.RevokedAt.Before(time.Now().Add(1*time.Second)))
}

func TestSessionRepository_DifferentUserAgents(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)

	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
	}

	for i, ua := range userAgents {
		expiresAt := time.Now().Add(24 * time.Hour)
		session := &session_domain.Session{
			UserID:           userID,
			RefreshTokenHash: fmt.Sprintf("hash-ua-%d", i),
			UserAgent:        ua,
			IPAddress:        "192.168.1.1",
			ExpiresAt:        &expiresAt,
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
	}

	// Verify user agents are stored correctly
	for i, ua := range userAgents {
		found, err := repo.FindByRefreshToken(ctx, fmt.Sprintf("hash-ua-%d", i))
		require.NoError(t, err)
		assert.Equal(t, ua, found.UserAgent)
	}
}

func TestSessionRepository_Create_WithNilUserAgent(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "hash-nil-ua",
		UserAgent:        "",
		IPAddress:        "",
		ExpiresAt:        &expiresAt,
	}

	err := repo.Create(ctx, session)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, session.ID)
}

func TestSessionRepository_Create_MultipleWithSameUser(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create multiple sessions for same user (should be allowed)
	for i := 0; i < 3; i++ {
		session := &session_domain.Session{
			UserID:           userID,
			RefreshTokenHash: fmt.Sprintf("hash-%d", i),
			UserAgent:        "Mozilla/5.0",
			IPAddress:        "192.168.1.1",
			ExpiresAt:        &expiresAt,
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, session.ID)
	}
}

func TestSessionRepository_Revoke_AlreadyRevoked(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "hash-revoke-twice",
		UserAgent:        "Mozilla/5.0",
		IPAddress:        "192.168.1.1",
		ExpiresAt:        &expiresAt,
	}
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Revoke once
	err = repo.Revoke(ctx, session.ID)
	require.NoError(t, err)

	// Revoke again (should not error)
	err = repo.Revoke(ctx, session.ID)
	assert.NoError(t, err)
}

func TestSessionRepository_FindByRefreshToken_Revoked(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "hash-check-revoked",
		UserAgent:        "Mozilla/5.0",
		IPAddress:        "192.168.1.1",
		ExpiresAt:        &expiresAt,
	}
	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Revoke session
	err = repo.Revoke(ctx, session.ID)
	require.NoError(t, err)

	// Find should still work but have RevokedAt set
	found, err := repo.FindByRefreshToken(ctx, "hash-check-revoked")
	require.NoError(t, err)
	assert.NotNil(t, found.RevokedAt)
}

func TestSessionRepository_Create_DatabaseError(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewSessionRepository(testDB)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Try to create session with invalid user_id (foreign key constraint)
	session := &session_domain.Session{
		UserID:           uuid.New(), // Non-existent user
		RefreshTokenHash: "hash-invalid-user",
		UserAgent:        "Mozilla/5.0",
		IPAddress:        "192.168.1.1",
		ExpiresAt:        &expiresAt,
	}

	err := repo.Create(ctx, session)
	assert.Error(t, err) // Should fail due to foreign key constraint
}

func TestSessionRepository_FindByRefreshToken_EmptyHash(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewSessionRepository(testDB)

	found, err := repo.FindByRefreshToken(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestSessionRepository_Create_LongUserAgent(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Very long user agent string
	longUserAgent := string(make([]byte, 1000))
	for i := range longUserAgent {
		longUserAgent = longUserAgent[:i] + "A"
	}

	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "hash-long-ua",
		UserAgent:        longUserAgent,
		IPAddress:        "192.168.1.1",
		ExpiresAt:        &expiresAt,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Verify it was saved
	found, err := repo.FindByRefreshToken(ctx, "hash-long-ua")
	require.NoError(t, err)
	assert.NotNil(t, found)
}

func TestSessionRepository_Create_SpecialCharacters(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Hash with special characters
	specialHash := "hash-with-!@#$%^&*()_+-=[]{}|;':\",./<>?"

	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: specialHash,
		UserAgent:        "Mozilla/5.0",
		IPAddress:        "192.168.1.1",
		ExpiresAt:        &expiresAt,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Verify it can be found
	found, err := repo.FindByRefreshToken(ctx, specialHash)
	require.NoError(t, err)
	assert.Equal(t, specialHash, found.RefreshTokenHash)
}

func TestSessionRepository_Revoke_MultipleRevokes(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)

	// Create multiple sessions
	sessionIDs := make([]uuid.UUID, 3)
	for i := 0; i < 3; i++ {
		expiresAt := time.Now().Add(24 * time.Hour)
		session := &session_domain.Session{
			UserID:           userID,
			RefreshTokenHash: fmt.Sprintf("hash-multi-revoke-%d", i),
			UserAgent:        "Mozilla/5.0",
			IPAddress:        "192.168.1.1",
			ExpiresAt:        &expiresAt,
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
		sessionIDs[i] = session.ID
	}

	// Revoke all sessions
	for _, id := range sessionIDs {
		err := repo.Revoke(ctx, id)
		require.NoError(t, err)
	}

	// Verify all are revoked
	for i := 0; i < 3; i++ {
		found, err := repo.FindByRefreshToken(ctx, fmt.Sprintf("hash-multi-revoke-%d", i))
		require.NoError(t, err)
		assert.NotNil(t, found.RevokedAt)
	}
}

// Transaction Tests - Cover BaseRepository.Q()

func TestSessionRepository_WithTransaction_Commit(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewSessionRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	// Create user first
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "txtest@example.com")
	require.NoError(t, err)

	var sessionID uuid.UUID

	// Execute operations within transaction
	err = txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		expiresAt := time.Now().Add(24 * time.Hour)
		session := &session_domain.Session{
			UserID:           userID,
			RefreshTokenHash: "tx_commit_token",
			UserAgent:        "TxTest",
			IPAddress:        "10.0.0.1",
			ExpiresAt:        &expiresAt,
		}

		if err := repo.Create(txCtx, session); err != nil {
			return err
		}
		sessionID = session.ID
		return nil
	})

	require.NoError(t, err)

	// Verify session was committed and persisted
	found, err := repo.FindByRefreshToken(ctx, "tx_commit_token")
	require.NoError(t, err)
	assert.Equal(t, sessionID, found.ID)
	assert.Equal(t, userID, found.UserID)
}

func TestSessionRepository_WithTransaction_Rollback(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewSessionRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	// Create user first
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "txrollback@example.com")
	require.NoError(t, err)

	// Execute operations within transaction that fails
	err = txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		expiresAt := time.Now().Add(24 * time.Hour)
		session := &session_domain.Session{
			UserID:           userID,
			RefreshTokenHash: "tx_rollback_token",
			UserAgent:        "TxTest",
			IPAddress:        "10.0.0.1",
			ExpiresAt:        &expiresAt,
		}

		if err := repo.Create(txCtx, session); err != nil {
			return err
		}

		// Force rollback by returning error
		return errors.New("intentional rollback")
	})

	require.Error(t, err)
	assert.Equal(t, "intentional rollback", err.Error())

	// Verify session was NOT persisted due to rollback
	_, err = repo.FindByRefreshToken(ctx, "tx_rollback_token")
	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrNotFound, err)
}

func TestSessionRepository_WithTransaction_MultipleOperations(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewSessionRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	// Create user first
	userID := uuid.New()
	_, err := testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "txmulti@example.com")
	require.NoError(t, err)

	// Create initial session outside transaction
	expiresAt1 := time.Now().Add(24 * time.Hour)
	initialSession := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "initial_token",
		UserAgent:        "Initial",
		IPAddress:        "10.0.0.1",
		ExpiresAt:        &expiresAt1,
	}
	err = repo.Create(ctx, initialSession)
	require.NoError(t, err)

	// Execute multiple operations within transaction
	err = txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Revoke old session
		if err := repo.Revoke(txCtx, initialSession.ID); err != nil {
			return err
		}

		// Create new session
		expiresAt2 := time.Now().Add(24 * time.Hour)
		newSession := &session_domain.Session{
			UserID:           userID,
			RefreshTokenHash: "new_token",
			UserAgent:        "NewAgent",
			IPAddress:        "10.0.0.2",
			ExpiresAt:        &expiresAt2,
		}
		return repo.Create(txCtx, newSession)
	})

	require.NoError(t, err)

	// Verify old session was revoked (FindByRefreshToken returns revoked sessions)
	oldSession, err := repo.FindByRefreshToken(ctx, "initial_token")
	require.NoError(t, err)
	assert.NotNil(t, oldSession.RevokedAt, "Session should be revoked")

	// Verify new session exists and is not revoked
	newFound, err := repo.FindByRefreshToken(ctx, "new_token")
	require.NoError(t, err)
	assert.Equal(t, userID, newFound.UserID)
	assert.Equal(t, "NewAgent", newFound.UserAgent)
	assert.Nil(t, newFound.RevokedAt, "New session should not be revoked")
}

// High Priority Tests - Database Constraints

func TestSessionRepository_Create_ForeignKeyConstraint(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewSessionRepository(testDB)

	// Try to create session with non-existent user
	nonExistentUserID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           nonExistentUserID,
		RefreshTokenHash: "hash123",
		UserAgent:        "TestAgent",
		IPAddress:        "127.0.0.1",
		ExpiresAt:        &expiresAt,
	}

	err := repo.Create(ctx, session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "violates foreign key constraint")
}

func TestSessionRepository_Create_ValidatesTimestamps(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)

	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "hash123",
		UserAgent:        "TestAgent",
		IPAddress:        "127.0.0.1",
		ExpiresAt:        &expiresAt,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Verify timestamps were set correctly
	found, err := repo.FindByRefreshToken(ctx, "hash123")
	require.NoError(t, err)

	assert.NotZero(t, found.CreatedAt, "CreatedAt should be set")
	assert.True(t, found.CreatedAt.Before(time.Now().Add(1*time.Second)), "CreatedAt should be in the past")
	assert.True(t, found.ExpiresAt.After(time.Now()), "ExpiresAt should be in the future")
	assert.Nil(t, found.RevokedAt, "RevokedAt should be nil for new session")
}

func TestSessionRepository_Create_ContextCanceled(t *testing.T) {
	cleanDB(t)
	userID := createTestUser(t, context.Background())

	repo := pg_repo.NewSessionRepository(testDB)

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "hash123",
		UserAgent:        "TestAgent",
		IPAddress:        "127.0.0.1",
		ExpiresAt:        &expiresAt,
	}

	err := repo.Create(ctx, session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestSessionRepository_Transaction_PersistenceValidation(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	var sessionID uuid.UUID

	// Execute within transaction
	err := txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		expiresAt := time.Now().Add(24 * time.Hour)
		session := &session_domain.Session{
			UserID:           userID,
			RefreshTokenHash: "tx_persist_token",
			UserAgent:        "TxTest",
			IPAddress:        "10.0.0.1",
			ExpiresAt:        &expiresAt,
		}

		if err := repo.Create(txCtx, session); err != nil {
			return err
		}
		sessionID = session.ID
		return nil
	})

	require.NoError(t, err)

	// Verify persistence OUTSIDE transaction with fresh context
	freshCtx := context.Background()
	found, err := repo.FindByRefreshToken(freshCtx, "tx_persist_token")
	require.NoError(t, err, "Session should persist after transaction commit")
	assert.Equal(t, sessionID, found.ID)
	assert.Equal(t, userID, found.UserID)
	assert.NotZero(t, found.CreatedAt)
}

// TestSessionRepository_Concurrent_CreateSessions tests concurrent session creation
// to verify thread safety and detect potential race conditions
func TestSessionRepository_Concurrent_CreateSessions(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)

	const numGoroutines = 10
	errors := make(chan error, numGoroutines)
	sessionIDs := make(chan uuid.UUID, numGoroutines)

	// Create sessions concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			expiresAt := time.Now().Add(24 * time.Hour)
			session := &session_domain.Session{
				UserID:           userID,
				RefreshTokenHash: fmt.Sprintf("concurrent_hash_%d", index),
				UserAgent:        "ConcurrentAgent",
				IPAddress:        "127.0.0.1",
				ExpiresAt:        &expiresAt,
			}

			err := repo.Create(ctx, session)
			if err != nil {
				errors <- err
				return
			}

			sessionIDs <- session.ID
			errors <- nil
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		require.NoError(t, err, "Concurrent create should not fail")
	}

	close(sessionIDs)
	uniqueIDs := make(map[uuid.UUID]bool)
	for id := range sessionIDs {
		uniqueIDs[id] = true
	}

	require.Equal(t, numGoroutines, len(uniqueIDs), "All sessions should have unique IDs")
}

// TestSessionRepository_FindByRefreshToken_ContextTimeout tests that
// FindByRefreshToken properly respects context timeout
func TestSessionRepository_FindByRefreshToken_ContextTimeout(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)

	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "timeout_test_hash",
		UserAgent:        "TimeoutAgent",
		IPAddress:        "127.0.0.1",
		ExpiresAt:        &expiresAt,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// Create context with very short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout has passed

	_, err = repo.FindByRefreshToken(timeoutCtx, "timeout_test_hash")
	require.Error(t, err)
	require.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled),
		"Error should be context deadline exceeded or canceled")
}

// TestSessionRepository_Revoke_VerifiesRevokedAtTimestamp verifies that
// the revoked_at timestamp is properly set when revoking a session
func TestSessionRepository_Revoke_VerifiesRevokedAtTimestamp(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)

	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "revoke_timestamp_hash",
		UserAgent:        "RevokeAgent",
		IPAddress:        "127.0.0.1",
		ExpiresAt:        &expiresAt,
	}

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	beforeRevoke := time.Now().Add(-1 * time.Second)
	err = repo.Revoke(ctx, session.ID)
	require.NoError(t, err)
	afterRevoke := time.Now().Add(1 * time.Second)

	// Verify revoked_at is set
	found, err := repo.FindByRefreshToken(ctx, "revoke_timestamp_hash")
	require.NoError(t, err)
	require.NotNil(t, found.RevokedAt, "RevokedAt should be set")
	assert.True(t, found.RevokedAt.After(beforeRevoke) && found.RevokedAt.Before(afterRevoke),
		"RevokedAt should be between revoke operation time")
}

// TestSessionRepository_Create_TableDriven tests various create scenarios using table-driven approach
func TestSessionRepository_Create_TableDriven(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	userID := createTestUser(t, ctx)

	repo := pg_repo.NewSessionRepository(testDB)

	tests := []struct {
		name      string
		hash      string
		userAgent string
		ipAddress string
		wantErr   bool
	}{
		{
			name:      "valid session with all fields",
			hash:      "table_hash_1",
			userAgent: "Mozilla/5.0",
			ipAddress: "192.168.1.1",
			wantErr:   false,
		},
		{
			name:      "valid session with empty user agent",
			hash:      "table_hash_2",
			userAgent: "",
			ipAddress: "192.168.1.2",
			wantErr:   false,
		},
		{
			name:      "valid session with IPv6",
			hash:      "table_hash_3",
			userAgent: "Chrome/90.0",
			ipAddress: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			wantErr:   false,
		},
		{
			name:      "valid session with special chars in hash",
			hash:      "table_hash_4!@#$%",
			userAgent: "Safari/14.0",
			ipAddress: "10.0.0.1",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiresAt := time.Now().Add(24 * time.Hour)
			session := &session_domain.Session{
				UserID:           userID,
				RefreshTokenHash: tt.hash,
				UserAgent:        tt.userAgent,
				IPAddress:        tt.ipAddress,
				ExpiresAt:        &expiresAt,
			}

			err := repo.Create(ctx, session)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, session.ID)
				assert.NotZero(t, session.CreatedAt)
			}
		})
	}
}

// BenchmarkSessionRepository_Create benchmarks session creation performance
func BenchmarkSessionRepository_Create(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	_, _ = testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "bench@example.com")

	repo := pg_repo.NewSessionRepository(testDB)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expiresAt := time.Now().Add(24 * time.Hour)
		session := &session_domain.Session{
			UserID:           userID,
			RefreshTokenHash: fmt.Sprintf("bench_hash_%d", i),
			UserAgent:        "BenchAgent",
			IPAddress:        "127.0.0.1",
			ExpiresAt:        &expiresAt,
		}
		_ = repo.Create(ctx, session)
	}
}

// BenchmarkSessionRepository_FindByRefreshToken benchmarks session lookup performance
func BenchmarkSessionRepository_FindByRefreshToken(b *testing.B) {
	ctx := context.Background()
	userID := uuid.New()
	_, _ = testDB.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", userID, "bench-find@example.com")

	repo := pg_repo.NewSessionRepository(testDB)

	// Create test session
	expiresAt := time.Now().Add(24 * time.Hour)
	session := &session_domain.Session{
		UserID:           userID,
		RefreshTokenHash: "bench_find_hash",
		UserAgent:        "BenchAgent",
		IPAddress:        "127.0.0.1",
		ExpiresAt:        &expiresAt,
	}
	_ = repo.Create(ctx, session)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.FindByRefreshToken(ctx, "bench_find_hash")
	}
}
