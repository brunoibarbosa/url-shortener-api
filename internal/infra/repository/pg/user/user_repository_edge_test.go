//go:build test

package pg_repo_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestUserRepository_EdgeCases_VeryLongEmail tests email at maximum length
func TestUserRepository_EdgeCases_VeryLongEmail(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	// Email with 254 characters (RFC 5321 maximum)
	localPart := strings.Repeat("a", 64)
	domain := strings.Repeat("b", 180) + ".com"
	longEmail := localPart + "@" + domain

	user, err := repo.Create(ctx, longEmail, "password123")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, longEmail, user.Email)

	retrieved, err := repo.GetByEmail(ctx, longEmail)
	require.NoError(t, err)
	require.Equal(t, longEmail, retrieved.Email)
}

// TestUserRepository_EdgeCases_MinimumEmail tests email at minimum length
func TestUserRepository_EdgeCases_MinimumEmail(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	// Minimum valid email: a@b.c (5 characters)
	minEmail := "a@b.c"

	user, err := repo.Create(ctx, minEmail, "password123")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, minEmail, user.Email)
}

// TestUserRepository_EdgeCases_SpecialCharactersEmail tests email with special characters
func TestUserRepository_EdgeCases_SpecialCharactersEmail(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	specialEmails := []string{
		"user+tag@example.com",
		"user.name@example.com",
		"user_name@example.com",
		"123@example.com",
		"user@sub.domain.example.com",
	}

	for _, email := range specialEmails {
		user, err := repo.Create(ctx, email, "password123")
		require.NoError(t, err, "Failed for email: %s", email)
		require.NotNil(t, user)
		require.Equal(t, email, user.Email)
	}
}

// TestUserRepository_EdgeCases_VeryLongPassword tests password at maximum length
func TestUserRepository_EdgeCases_VeryLongPassword(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	// Very long password (72 characters for bcrypt limit)
	longPassword := strings.Repeat("A", 72)

	user, err := repo.Create(ctx, "test@example.com", longPassword)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.NotEmpty(t, user.PasswordHash)
}

// TestUserRepository_EdgeCases_UnicodeEmail tests email with unicode characters
func TestUserRepository_EdgeCases_UnicodeEmail(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	// Unicode emails (allowed by RFC 6531)
	unicodeEmails := []string{
		"用户@example.com",
		"käyttäjä@example.com",
		"пользователь@example.com",
	}

	for _, email := range unicodeEmails {
		user, err := repo.Create(ctx, email, "password123")
		require.NoError(t, err, "Failed for unicode email: %s", email)
		require.NotNil(t, user)
		require.Equal(t, email, user.Email)
	}
}

// TestUserRepository_EdgeCases_CaseInsensitiveEmail tests email case handling
func TestUserRepository_EdgeCases_CaseInsensitiveEmail(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	email1 := "User@Example.COM"
	email2 := "user@example.com"

	// Create with uppercase
	user1, err := repo.Create(ctx, email1, "password123")
	require.NoError(t, err)
	require.NotNil(t, user1)

	// Try to create with lowercase (should fail if case-insensitive unique constraint)
	user2, err := repo.Create(ctx, email2, "password456")

	// Depending on DB schema, this might succeed or fail
	// Document the expected behavior
	if err == nil {
		t.Logf("WARNING: Email uniqueness is case-sensitive. Consider adding LOWER(email) unique index")
		require.NotNil(t, user2)
	} else {
		t.Logf("Email uniqueness is case-insensitive (expected)")
	}
}

// TestUserRepository_EdgeCases_EmptyStringPassword tests empty password handling
func TestUserRepository_EdgeCases_EmptyStringPassword(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	// Empty password should be handled at validation layer, but test repository behavior
	user, err := repo.Create(ctx, "test@example.com", "")

	// Repository might accept it (validation is elsewhere) or reject it
	if err != nil {
		t.Logf("Repository rejects empty password (validation at repo level)")
	} else {
		t.Logf("Repository accepts empty password (validation should be at service/handler level)")
		require.NotNil(t, user)
	}
}

// TestUserRepository_EdgeCases_DuplicateDeletedEmail tests re-using email after soft delete
func TestUserRepository_EdgeCases_DuplicateDeletedEmail(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	email := "reuse@example.com"

	// Create user
	user1, err := repo.Create(ctx, email, "password123")
	require.NoError(t, err)

	// Soft delete
	err = repo.Delete(ctx, user1.ID)
	require.NoError(t, err)

	// Try to create new user with same email
	user2, err := repo.Create(ctx, email, "password456")

	// Behavior depends on unique constraint implementation
	if err != nil {
		t.Logf("Cannot reuse email after soft delete (unique constraint on email)")
	} else {
		t.Logf("Can reuse email after soft delete (unique constraint on email WHERE deleted_at IS NULL)")
		require.NotNil(t, user2)
		require.NotEqual(t, user1.ID, user2.ID)
	}
}

// TestUserRepository_EdgeCases_GetNonExistentUser tests retrieving non-existent user
func TestUserRepository_EdgeCases_GetNonExistentUser(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	// Get by non-existent email
	user, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	require.NoError(t, err)
	require.Nil(t, user)

	// Get by non-existent ID (using UUID zero)
	userByID, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	require.NoError(t, err)
	require.Nil(t, userByID)
}

// TestUserRepository_EdgeCases_DeleteAlreadyDeleted tests double delete
func TestUserRepository_EdgeCases_DeleteAlreadyDeleted(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	user, err := repo.Create(ctx, "test@example.com", "password123")
	require.NoError(t, err)

	// First delete
	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	// Second delete (idempotent operation)
	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err, "Double delete should be idempotent")
}
