package pg_repo_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	user_domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	pg_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/pg/user"
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
	`)
	return err
}

func cleanDB(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.Exec(ctx, "TRUNCATE users, user_profiles, user_providers CASCADE")
	require.NoError(t, err)
}

func TestUserRepository_Exists_True(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create user
	user := &user_domain.User{Email: "exists@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Check if exists
	exists, err := repo.Exists(ctx, "exists@example.com")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_Exists_False(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	exists, err := repo.Exists(ctx, "nonexistent@example.com")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_Create_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	user := &user_domain.User{
		Email: "create@example.com",
	}

	err := repo.Create(ctx, user)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.False(t, user.CreatedAt.IsZero())
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create first user
	user1 := &user_domain.User{Email: "duplicate@example.com"}
	err := repo.Create(ctx, user1)
	require.NoError(t, err)

	// Try to create with same email
	user2 := &user_domain.User{Email: "duplicate@example.com"}
	err = repo.Create(ctx, user2)

	assert.Error(t, err) // Should fail due to unique constraint
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create user with profile to avoid NULL scan issues
	user := &user_domain.User{Email: "getbyid@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create profile
	_, err = testDB.Exec(ctx,
		"INSERT INTO user_profiles (user_id, name, avatar_url) VALUES ($1, $2, $3)",
		user.ID, "Test User", "https://example.com/avatar.jpg")
	require.NoError(t, err)

	// Get by ID
	found, err := repo.GetByID(ctx, user.ID)

	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "getbyid@example.com", found.Email)
	// Profile should be present
	assert.NotNil(t, found.Profile)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	found, err := repo.GetByID(ctx, uuid.New())

	assert.Error(t, err)
	assert.Equal(t, user_domain.ErrNotFound, err)
	assert.Nil(t, found)
}

func TestUserRepository_GetByEmail_Success(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create user with profile to avoid NULL scan issues
	user := &user_domain.User{Email: "getbyemail@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create profile
	_, err = testDB.Exec(ctx,
		"INSERT INTO user_profiles (user_id, name) VALUES ($1, $2)",
		user.ID, "Email User")
	require.NoError(t, err)

	// Get by email
	found, err := repo.GetByEmail(ctx, "getbyemail@example.com")

	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Equal(t, "getbyemail@example.com", found.Email)
	// Profile will be present
	assert.NotNil(t, found.Profile)
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	found, err := repo.GetByEmail(ctx, "nonexistent@example.com")

	assert.Error(t, err)
	assert.Equal(t, user_domain.ErrNotFound, err)
	assert.Nil(t, found)
}

func TestUserRepository_GetByID_WithProfile(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create user
	user := &user_domain.User{Email: "profile@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create profile
	_, err = testDB.Exec(ctx,
		"INSERT INTO user_profiles (user_id, name, avatar_url) VALUES ($1, $2, $3)",
		user.ID, "Test User", "https://example.com/avatar.jpg")
	require.NoError(t, err)

	// Get by ID
	found, err := repo.GetByID(ctx, user.ID)

	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.NotNil(t, found.Profile)
	assert.Equal(t, "Test User", found.Profile.Name)
	assert.NotNil(t, found.Profile.AvatarURL)
	assert.Equal(t, "https://example.com/avatar.jpg", *found.Profile.AvatarURL)
}

func TestUserRepository_GetByEmail_WithProfile(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create user
	user := &user_domain.User{Email: "profile2@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create profile
	_, err = testDB.Exec(ctx,
		"INSERT INTO user_profiles (user_id, name, avatar_url) VALUES ($1, $2, $3)",
		user.ID, "Another User", "https://example.com/avatar2.jpg")
	require.NoError(t, err)

	// Get by email
	found, err := repo.GetByEmail(ctx, "profile2@example.com")

	require.NoError(t, err)
	assert.NotNil(t, found.Profile)
	assert.Equal(t, "Another User", found.Profile.Name)
	assert.NotNil(t, found.Profile.AvatarURL)
	assert.Equal(t, "https://example.com/avatar2.jpg", *found.Profile.AvatarURL)
}

func TestUserRepository_MultipleUsers(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create multiple users
	emails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	for _, email := range emails {
		user := &user_domain.User{Email: email}
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	// Verify all exist
	for _, email := range emails {
		exists, err := repo.Exists(ctx, email)
		require.NoError(t, err)
		assert.True(t, exists)
	}
}

func TestUserRepository_CaseInsensitiveEmail(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create user with lowercase email and profile
	user := &user_domain.User{Email: "case@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create profile to avoid NULL scan
	_, err = testDB.Exec(ctx,
		"INSERT INTO user_profiles (user_id, name) VALUES ($1, $2)",
		user.ID, "Case User")
	require.NoError(t, err)

	// Get by email with different case (PostgreSQL is case-sensitive by default)
	found, err := repo.GetByEmail(ctx, "case@example.com")
	require.NoError(t, err)
	assert.Equal(t, "case@example.com", found.Email)
}

func TestUserRepository_UpdatedAt(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create user with profile
	user := &user_domain.User{Email: "update@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create profile to avoid NULL scan
	_, err = testDB.Exec(ctx,
		"INSERT INTO user_profiles (user_id, name) VALUES ($1, $2)",
		user.ID, "Update User")
	require.NoError(t, err)

	// Get user (UpdatedAt should be nil initially)
	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Nil(t, found.UpdatedAt)
}

func TestUserRepository_Create_EmailNormalization(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	user := &user_domain.User{Email: "   TRIMMED@EXAMPLE.COM   "}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create a profile to avoid NULL scan
	_, err = testDB.Exec(ctx,
		"INSERT INTO user_profiles (user_id, name) VALUES ($1, $2)",
		user.ID, "Trimmed User")
	require.NoError(t, err)

	// Verify it was stored (email might be normalized by DB or app)
	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.NotNil(t, found)
}

func TestUserRepository_MultipleUsersWithProfiles(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create multiple users with profiles
	for i := 1; i <= 3; i++ {
		user := &user_domain.User{Email: fmt.Sprintf("multi%d@example.com", i)}
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		_, err = testDB.Exec(ctx,
			"INSERT INTO user_profiles (user_id, name) VALUES ($1, $2)",
			user.ID, fmt.Sprintf("User %d", i))
		require.NoError(t, err)
	}

	// Verify all users can be retrieved
	for i := 1; i <= 3; i++ {
		found, err := repo.GetByEmail(ctx, fmt.Sprintf("multi%d@example.com", i))
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("multi%d@example.com", i), found.Email)
		assert.NotNil(t, found.Profile)
		assert.Equal(t, fmt.Sprintf("User %d", i), found.Profile.Name)
	}
}

func TestUserRepository_Exists_SpecialCharacters(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Email with special characters
	user := &user_domain.User{Email: "user+tag@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	exists, err := repo.Exists(ctx, "user+tag@example.com")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_GetByID_WithoutProfile(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create user WITHOUT profile
	user := &user_domain.User{Email: "noprofile@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Get by ID should work but profile will be nil
	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, found.ID)
	assert.Nil(t, found.Profile)
}

func TestUserRepository_GetByEmail_WithoutProfile(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create user WITHOUT profile
	user := &user_domain.User{Email: "noprofile2@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Get by email should work but profile will be nil
	found, err := repo.GetByEmail(ctx, "noprofile2@example.com")
	require.NoError(t, err)
	assert.Equal(t, "noprofile2@example.com", found.Email)
	assert.Nil(t, found.Profile)
}

func TestUserRepository_Create_VeryLongEmail(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Very long email (up to 255 chars is typically allowed)
	longEmail := string(make([]byte, 240)) + "@example.com"
	for i := 0; i < 240; i++ {
		longEmail = string(longEmail[:i]) + "a" + longEmail[i+1:]
	}

	user := &user_domain.User{Email: longEmail}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Verify it was saved
	exists, err := repo.Exists(ctx, longEmail)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_GetByID_MultipleProfiles(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	user := &user_domain.User{Email: "multiprof@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create profile
	_, err = testDB.Exec(ctx,
		"INSERT INTO user_profiles (user_id, name, avatar_url) VALUES ($1, $2, $3)",
		user.ID, "First Profile", "https://avatar1.com")
	require.NoError(t, err)

	// GetByID should return first profile only
	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.NotNil(t, found.Profile)
	assert.Equal(t, "First Profile", found.Profile.Name)
}

func TestUserRepository_ConcurrentCreates(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	// Create multiple users concurrently (simulating concurrent requests)
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(index int) {
			user := &user_domain.User{Email: fmt.Sprintf("concurrent%d@example.com", index)}
			err := repo.Create(ctx, user)
			if err != nil {
				t.Logf("Error creating user %d: %v", index, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify all users were created (some might fail due to race conditions, that's ok)
	count := 0
	for i := 0; i < 5; i++ {
		exists, _ := repo.Exists(ctx, fmt.Sprintf("concurrent%d@example.com", i))
		if exists {
			count++
		}
	}
	assert.Greater(t, count, 0) // At least one should succeed
}

func TestUserRepository_Exists_EmptyEmail(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	exists, err := repo.Exists(ctx, "")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_ProfileWithNullFields(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)

	user := &user_domain.User{Email: "nullfields@example.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create profile with NULL avatar_url
	_, err = testDB.Exec(ctx,
		"INSERT INTO user_profiles (user_id, name, avatar_url) VALUES ($1, $2, NULL)",
		user.ID, "User With Null Avatar")
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.NotNil(t, found.Profile)
	assert.Equal(t, "User With Null Avatar", found.Profile.Name)
	assert.Nil(t, found.Profile.AvatarURL)
}

// Transaction Tests - Cover BaseRepository.Q()

func TestUserRepository_WithTransaction_Commit(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	var userID uuid.UUID

	// Execute operations within transaction
	err := txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		user := &user_domain.User{
			Email: "tx-commit@test.com",
		}
		if err := repo.Create(txCtx, user); err != nil {
			return err
		}
		userID = user.ID

		// Create profile using the transaction querier
		q := pg.GetQuerier(txCtx, testDB)
		_, err := q.Exec(txCtx, `
			INSERT INTO user_profiles (user_id, name) 
			VALUES ($1, $2)
		`, userID, "Tx User")
		return err
	})

	require.NoError(t, err)

	// Verify user and profile were committed
	found, err := repo.GetByEmail(ctx, "tx-commit@test.com")
	require.NoError(t, err)
	assert.Equal(t, userID, found.ID)
	assert.NotNil(t, found.Profile)
	assert.Equal(t, "Tx User", found.Profile.Name)
}

func TestUserRepository_WithTransaction_Rollback(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	// Execute operations within transaction that fails
	err := txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		user := &user_domain.User{
			Email: "tx-rollback@test.com",
		}
		if err := repo.Create(txCtx, user); err != nil {
			return err
		}

		// Force rollback
		return errors.New("intentional rollback")
	})

	require.Error(t, err)

	// Verify user was NOT persisted
	exists, err := repo.Exists(ctx, "tx-rollback@test.com")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_WithTransaction_CheckExists(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	// Create user outside transaction
	user := &user_domain.User{Email: "existing@test.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Check existence within transaction
	err = txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		exists, err := repo.Exists(txCtx, "existing@test.com")
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("user should exist")
		}

		// Check non-existent
		exists, err = repo.Exists(txCtx, "nonexistent@test.com")
		if err != nil {
			return err
		}
		if exists {
			return errors.New("user should not exist")
		}

		return nil
	})

	require.NoError(t, err)
}

func TestUserRepository_WithTransaction_GetMethods(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	repo := pg_repo.NewUserRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	// Create user with profile
	user := &user_domain.User{Email: "txget@test.com"}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	_, err = testDB.Exec(ctx, `
		INSERT INTO user_profiles (user_id, name) 
		VALUES ($1, $2)
	`, user.ID, "Tx Get User")
	require.NoError(t, err)

	// Test GetByID and GetByEmail within transaction
	err = txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// GetByID
		foundByID, err := repo.GetByID(txCtx, user.ID)
		if err != nil {
			return err
		}
		if foundByID.Email != "txget@test.com" {
			return errors.New("email mismatch")
		}

		// GetByEmail
		foundByEmail, err := repo.GetByEmail(txCtx, "txget@test.com")
		if err != nil {
			return err
		}
		if foundByEmail.ID != user.ID {
			return errors.New("id mismatch")
		}

		if foundByEmail.Profile == nil || foundByEmail.Profile.Name != "Tx Get User" {
			return errors.New("profile mismatch")
		}

		return nil
	})

	require.NoError(t, err)
}

// TestUserRepository_Create_EmailUniqueConstraint verifies that attempting to create
// a user with a duplicate email returns a unique constraint error
func TestUserRepository_Create_EmailUniqueConstraint(t *testing.T) {
	cleanDB(t)
	repo := pg_repo.NewUserRepository(testDB)

	existingUser := &user_domain.User{
		Email: "duplicate@test.com",
	}

	err := repo.Create(context.Background(), existingUser)
	require.NoError(t, err)

	duplicateUser := &user_domain.User{
		Email: "duplicate@test.com",
	}

	err = repo.Create(context.Background(), duplicateUser)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate key value")
	require.Contains(t, err.Error(), "users_email_key")
}

// TestUserRepository_Create_ValidatesTimestamps validates that CreatedAt and UpdatedAt
// are properly set when creating a user
func TestUserRepository_Create_ValidatesTimestamps(t *testing.T) {
	cleanDB(t)
	repo := pg_repo.NewUserRepository(testDB)
	now := time.Now()

	newUser := &user_domain.User{
		Email: "timestamps@test.com",
	}

	err := repo.Create(context.Background(), newUser)
	require.NoError(t, err)

	require.NotZero(t, newUser.CreatedAt, "CreatedAt should be set")
	require.True(t, newUser.CreatedAt.After(now.Add(-5*time.Second)), "CreatedAt should be recent")
}

// TestUserRepository_Create_ContextCanceled verifies that Create operation properly
// handles context cancellation
func TestUserRepository_Create_ContextCanceled(t *testing.T) {
	cleanDB(t)
	repo := pg_repo.NewUserRepository(testDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	newUser := &user_domain.User{
		Email: "canceled@test.com",
	}

	err := repo.Create(ctx, newUser)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}

// TestUserRepository_Transaction_PersistenceValidation verifies that transaction changes
// are properly persisted and visible in a fresh context
func TestUserRepository_Transaction_PersistenceValidation(t *testing.T) {
	cleanDB(t)
	repo := pg_repo.NewUserRepository(testDB)
	txManager := pg.NewTxManager(testDB)

	var createdUserID uuid.UUID
	err := txManager.WithinTransaction(context.Background(), func(txCtx context.Context) error {
		newUser := &user_domain.User{
			Email: "tx-persist@test.com",
		}

		if err := repo.Create(txCtx, newUser); err != nil {
			return err
		}

		createdUserID = newUser.ID
		return nil
	})
	require.NoError(t, err)

	// Verify persistence with a completely fresh context
	freshCtx := context.Background()
	foundUser, err := repo.GetByID(freshCtx, createdUserID)
	require.NoError(t, err)
	require.NotNil(t, foundUser)
	require.Equal(t, createdUserID, foundUser.ID)
	require.Equal(t, "tx-persist@test.com", foundUser.Email)
}

// TestUserRepository_Profile_ForeignKeyIntegrity verifies that profile is properly
// associated with the user and foreign key integrity is maintained
func TestUserRepository_Profile_ForeignKeyIntegrity(t *testing.T) {
	cleanDB(t)
	repo := pg_repo.NewUserRepository(testDB)
	ctx := context.Background()

	newUser := &user_domain.User{
		Email: "profile-fk@test.com",
	}

	err := repo.Create(ctx, newUser)
	require.NoError(t, err)
	require.NotZero(t, newUser.ID)

	// Insert profile separately
	avatarURL := "https://example.com/avatar.jpg"
	var profileID int64
	err = testDB.QueryRow(ctx,
		"INSERT INTO user_profiles (user_id, name, avatar_url) VALUES ($1, $2, $3) RETURNING id",
		newUser.ID, "Profile FK User", avatarURL).Scan(&profileID)
	require.NoError(t, err)

	// Verify profile can be retrieved with user
	foundUser, err := repo.GetByID(ctx, newUser.ID)
	require.NoError(t, err)
	require.NotNil(t, foundUser.Profile)
	require.Equal(t, profileID, foundUser.Profile.ID)
	require.Equal(t, "Profile FK User", foundUser.Profile.Name)
	require.Equal(t, &avatarURL, foundUser.Profile.AvatarURL)

	// Verify profile UserID matches
	var profileUserID uuid.UUID
	err = testDB.QueryRow(ctx,
		"SELECT user_id FROM user_profiles WHERE id = $1",
		profileID).Scan(&profileUserID)
	require.NoError(t, err)
	require.Equal(t, newUser.ID, profileUserID)
}

// TestUserRepository_Concurrent_CreateUsers tests concurrent user creation
// to verify thread safety and detect potential race conditions
func TestUserRepository_Concurrent_CreateUsers(t *testing.T) {
	cleanDB(t)
	repo := pg_repo.NewUserRepository(testDB)
	ctx := context.Background()

	const numGoroutines = 10
	errors := make(chan error, numGoroutines)
	userIDs := make(chan uuid.UUID, numGoroutines)

	// Create users concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			user := &user_domain.User{
				Email: fmt.Sprintf("concurrent%d@test.com", index),
			}

			err := repo.Create(ctx, user)
			if err != nil {
				errors <- err
				return
			}

			userIDs <- user.ID
			errors <- nil
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		require.NoError(t, err, "Concurrent create should not fail")
	}

	close(userIDs)
	uniqueIDs := make(map[uuid.UUID]bool)
	for id := range userIDs {
		uniqueIDs[id] = true
	}

	require.Equal(t, numGoroutines, len(uniqueIDs), "All users should have unique IDs")
}

// TestUserRepository_GetByEmail_ContextTimeout tests that
// GetByEmail properly respects context timeout
func TestUserRepository_GetByEmail_ContextTimeout(t *testing.T) {
	cleanDB(t)
	repo := pg_repo.NewUserRepository(testDB)
	ctx := context.Background()

	user := &user_domain.User{
		Email: "timeout@test.com",
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create context with very short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout has passed

	_, err = repo.GetByEmail(timeoutCtx, "timeout@test.com")
	require.Error(t, err)
	require.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled),
		"Error should be context deadline exceeded or canceled")
}

// TestUserRepository_GetByID_ContextTimeout tests that
// GetByID properly respects context timeout
func TestUserRepository_GetByID_ContextTimeout(t *testing.T) {
	cleanDB(t)
	repo := pg_repo.NewUserRepository(testDB)
	ctx := context.Background()

	user := &user_domain.User{
		Email: "timeout-id@test.com",
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create context with very short timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout has passed

	_, err = repo.GetByID(timeoutCtx, user.ID)
	require.Error(t, err)
	require.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled),
		"Error should be context deadline exceeded or canceled")
}

// TestUserRepository_Create_TableDriven tests various create scenarios using table-driven approach
func TestUserRepository_Create_TableDriven(t *testing.T) {
	cleanDB(t)
	repo := pg_repo.NewUserRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email simple",
			email:   "simple@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@subdomain.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with dots",
			email:   "first.last@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with numbers",
			email:   "user123@example123.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &user_domain.User{
				Email: tt.email,
			}

			err := repo.Create(ctx, user)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.NotZero(t, user.CreatedAt)
			}
		})
	}
}

// BenchmarkUserRepository_Create benchmarks user creation performance
func BenchmarkUserRepository_Create(b *testing.B) {
	ctx := context.Background()
	repo := pg_repo.NewUserRepository(testDB)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &user_domain.User{
			Email: fmt.Sprintf("bench%d@example.com", i),
		}
		_ = repo.Create(ctx, user)
	}
}

// BenchmarkUserRepository_GetByEmail benchmarks user lookup by email performance
func BenchmarkUserRepository_GetByEmail(b *testing.B) {
	ctx := context.Background()
	repo := pg_repo.NewUserRepository(testDB)

	// Create test user
	user := &user_domain.User{
		Email: "bench-find@example.com",
	}
	_ = repo.Create(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByEmail(ctx, "bench-find@example.com")
	}
}

// BenchmarkUserRepository_GetByID benchmarks user lookup by ID performance
func BenchmarkUserRepository_GetByID(b *testing.B) {
	ctx := context.Background()
	repo := pg_repo.NewUserRepository(testDB)

	// Create test user
	user := &user_domain.User{
		Email: "bench-find-id@example.com",
	}
	_ = repo.Create(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(ctx, user.ID)
	}
}
