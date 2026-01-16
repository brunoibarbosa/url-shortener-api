//go:build test

package pg_repo_test

import (
	"context"
	"testing"
	"time"

	"url-shortener-api/internal/infra/repository/pg/testhelpers"

	"github.com/stretchr/testify/require"
)

// TestUserRepository_PropertyBased_CreateAndRetrieve validates that any valid user can be created and retrieved
func TestUserRepository_PropertyBased_CreateAndRetrieve(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	repo := setupRepository(t)
	ctx := context.Background()

	// Property: For any valid user data, Create followed by GetByEmail should return the same user
	for i := 0; i < 100; i++ {
		email := testhelpers.Email()
		password := testhelpers.AlphaNumericString(8, 32)

		// Create user
		user, err := repo.Create(ctx, email, password)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.Equal(t, email, user.Email)

		// Retrieve user
		retrieved, err := repo.GetByEmail(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, user.ID, retrieved.ID)
		require.Equal(t, user.Email, retrieved.Email)
		require.Equal(t, user.PasswordHash, retrieved.PasswordHash)
	}
}

// TestUserRepository_PropertyBased_EmailUniqueness validates email uniqueness invariant
func TestUserRepository_PropertyBased_EmailUniqueness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	repo := setupRepository(t)
	ctx := context.Background()

	// Property: Creating two users with the same email should fail
	for i := 0; i < 50; i++ {
		email := testhelpers.Email()
		password1 := testhelpers.AlphaNumericString(8, 32)
		password2 := testhelpers.AlphaNumericString(8, 32)

		// First creation should succeed
		user1, err := repo.Create(ctx, email, password1)
		require.NoError(t, err)
		require.NotNil(t, user1)

		// Second creation with same email should fail
		user2, err := repo.Create(ctx, email, password2)
		require.Error(t, err)
		require.Nil(t, user2)
	}
}

// TestUserRepository_PropertyBased_TimestampOrdering validates timestamp invariants
func TestUserRepository_PropertyBased_TimestampOrdering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	repo := setupRepository(t)
	ctx := context.Background()

	// Property: CreatedAt <= UpdatedAt for all users
	for i := 0; i < 100; i++ {
		email := testhelpers.Email()
		password := testhelpers.AlphaNumericString(8, 32)

		user, err := repo.Create(ctx, email, password)
		require.NoError(t, err)
		require.NotNil(t, user)

		// CreatedAt should be before or equal to UpdatedAt
		require.True(t,
			user.CreatedAt.Before(user.UpdatedAt) || user.CreatedAt.Equal(user.UpdatedAt),
			"CreatedAt (%v) should be <= UpdatedAt (%v)", user.CreatedAt, user.UpdatedAt,
		)

		// Both timestamps should be recent (within last 5 seconds)
		now := time.Now()
		require.WithinDuration(t, now, user.CreatedAt, 5*time.Second)
		require.WithinDuration(t, now, user.UpdatedAt, 5*time.Second)
	}
}

// TestUserRepository_PropertyBased_SoftDelete validates soft delete properties
func TestUserRepository_PropertyBased_SoftDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	repo := setupRepository(t)
	ctx := context.Background()

	// Property: Deleted users should not be retrievable by email
	for i := 0; i < 50; i++ {
		email := testhelpers.Email()
		password := testhelpers.AlphaNumericString(8, 32)

		// Create user
		user, err := repo.Create(ctx, email, password)
		require.NoError(t, err)

		// Delete user
		err = repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		// Should not be retrievable by email
		retrieved, err := repo.GetByEmail(ctx, email)
		require.NoError(t, err)
		require.Nil(t, retrieved, "Deleted user should not be retrievable")

		// But should be retrievable by ID (soft delete)
		retrievedByID, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedByID)
		require.NotNil(t, retrievedByID.DeletedAt, "DeletedAt should be set")
	}
}
