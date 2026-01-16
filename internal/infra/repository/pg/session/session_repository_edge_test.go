//go:build test

package pg_repo_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"url-shortener-api/internal/infra/repository/pg/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestSessionRepository_EdgeCases_VeryLongRefreshToken tests token at maximum length
func TestSessionRepository_EdgeCases_VeryLongRefreshToken(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	longToken := strings.Repeat("a", 512)

	session := testhelpers.SessionFixture().
		WithUserID(userID).
		WithRefreshToken(longToken).
		ToSession()

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	retrieved, err := repo.FindByRefreshToken(ctx, longToken)
	require.NoError(t, err)
	require.Equal(t, longToken, retrieved.RefreshToken)
}

// TestSessionRepository_EdgeCases_MinimumRefreshToken tests token at minimum length
func TestSessionRepository_EdgeCases_MinimumRefreshToken(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	minToken := "a"

	session := testhelpers.SessionFixture().
		WithUserID(userID).
		WithRefreshToken(minToken).
		ToSession()

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	retrieved, err := repo.FindByRefreshToken(ctx, minToken)
	require.NoError(t, err)
	require.Equal(t, minToken, retrieved.RefreshToken)
}

// TestSessionRepository_EdgeCases_SpecialCharactersUserAgent tests user agent with special chars
func TestSessionRepository_EdgeCases_SpecialCharactersUserAgent(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()

	specialUserAgents := []string{
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8.1.13) Gecko/20080311 Firefox/2.0.0.13",
		"User-Agent: *",
		strings.Repeat("A", 500),        // Very long
		"<script>alert('xss')</script>", // Potential XSS
		"'; DROP TABLE sessions; --",    // SQL injection attempt
	}

	for _, userAgent := range specialUserAgents {
		token := testhelpers.AlphaNumericString(32, 32)
		session := testhelpers.SessionFixture().
			WithUserID(userID).
			WithRefreshToken(token).
			WithUserAgent(userAgent).
			ToSession()

		err := repo.Create(ctx, session)
		require.NoError(t, err, "Failed for user agent: %s", userAgent)

		retrieved, err := repo.FindByRefreshToken(ctx, token)
		require.NoError(t, err)
		require.Equal(t, userAgent, retrieved.UserAgent)
	}
}

// TestSessionRepository_EdgeCases_IPAddressVariations tests various IP formats
func TestSessionRepository_EdgeCases_IPAddressVariations(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()

	ipAddresses := []string{
		"127.0.0.1",       // Localhost
		"0.0.0.0",         // Zero address
		"255.255.255.255", // Max IPv4
		"192.168.1.1",     // Private network
		"::1",             // IPv6 localhost
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334", // Full IPv6
		"2001:db8::1", // Compressed IPv6
	}

	for _, ipAddress := range ipAddresses {
		token := testhelpers.AlphaNumericString(32, 32)
		session := testhelpers.SessionFixture().
			WithUserID(userID).
			WithRefreshToken(token).
			WithIPAddress(ipAddress).
			ToSession()

		err := repo.Create(ctx, session)
		require.NoError(t, err, "Failed for IP: %s", ipAddress)

		retrieved, err := repo.FindByRefreshToken(ctx, token)
		require.NoError(t, err)
		require.Equal(t, ipAddress, retrieved.IPAddress)
	}
}

// TestSessionRepository_EdgeCases_ExpirationBoundaries tests expiration edge cases
func TestSessionRepository_EdgeCases_ExpirationBoundaries(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()

	testCases := []struct {
		name      string
		expiresAt time.Time
	}{
		{"Past expiration", time.Now().Add(-1 * time.Hour)},
		{"Current time", time.Now()},
		{"Near future", time.Now().Add(1 * time.Minute)},
		{"Far future", time.Now().Add(100 * 365 * 24 * time.Hour)}, // 100 years
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := testhelpers.AlphaNumericString(32, 32)
			session := testhelpers.SessionFixture().
				WithUserID(userID).
				WithRefreshToken(token).
				WithExpiresAt(tc.expiresAt).
				ToSession()

			err := repo.Create(ctx, session)
			require.NoError(t, err)

			retrieved, err := repo.FindByRefreshToken(ctx, token)
			require.NoError(t, err)
			require.WithinDuration(t, tc.expiresAt, retrieved.ExpiresAt, time.Second)
		})
	}
}

// TestSessionRepository_EdgeCases_RevokeAlreadyRevoked tests double revoke
func TestSessionRepository_EdgeCases_RevokeAlreadyRevoked(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	token := testhelpers.AlphaNumericString(32, 32)

	session := testhelpers.SessionFixture().
		WithUserID(userID).
		WithRefreshToken(token).
		ToSession()

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	// First revoke
	err = repo.Revoke(ctx, token)
	require.NoError(t, err)

	// Second revoke (idempotent)
	err = repo.Revoke(ctx, token)
	require.NoError(t, err, "Double revoke should be idempotent")

	retrieved, err := repo.FindByRefreshToken(ctx, token)
	require.NoError(t, err)
	require.NotNil(t, retrieved.RevokedAt)
}

// TestSessionRepository_EdgeCases_FindNonExistentToken tests retrieving non-existent token
func TestSessionRepository_EdgeCases_FindNonExistentToken(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	session, err := repo.FindByRefreshToken(ctx, "nonexistent-token")
	require.NoError(t, err)
	require.Nil(t, session)
}

// TestSessionRepository_EdgeCases_RevokeNonExistentToken tests revoking non-existent token
func TestSessionRepository_EdgeCases_RevokeNonExistentToken(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	err := repo.Revoke(ctx, "nonexistent-token")
	// Depending on implementation, might return error or no-op
	if err != nil {
		t.Logf("Revoke returns error for non-existent token")
	} else {
		t.Logf("Revoke is no-op for non-existent token")
	}
}

// TestSessionRepository_EdgeCases_ListWithZeroLimit tests list with limit 0
func TestSessionRepository_EdgeCases_ListWithZeroLimit(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()

	// Create some sessions
	for i := 0; i < 5; i++ {
		token := testhelpers.AlphaNumericString(32, 32)
		session := testhelpers.SessionFixture().
			WithUserID(userID).
			WithRefreshToken(token).
			ToSession()
		err := repo.Create(ctx, session)
		require.NoError(t, err)
	}

	// List with limit 0
	sessions, err := repo.ListByUserID(ctx, userID, 0, 0)
	require.NoError(t, err)
	// Should return empty or all sessions depending on implementation
	t.Logf("List with limit 0 returned %d sessions", len(sessions))
}

// TestSessionRepository_EdgeCases_ListWithNegativeOffset tests list with negative offset
func TestSessionRepository_EdgeCases_ListWithNegativeOffset(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()

	// Create some sessions
	for i := 0; i < 5; i++ {
		token := testhelpers.AlphaNumericString(32, 32)
		session := testhelpers.SessionFixture().
			WithUserID(userID).
			WithRefreshToken(token).
			ToSession()
		err := repo.Create(ctx, session)
		require.NoError(t, err)
	}

	// List with negative offset
	sessions, err := repo.ListByUserID(ctx, userID, 10, -5)

	if err != nil {
		t.Logf("Negative offset returns error")
	} else {
		t.Logf("Negative offset treated as 0, returned %d sessions", len(sessions))
		require.NotNil(t, sessions)
	}
}

// TestSessionRepository_EdgeCases_DuplicateRefreshToken tests duplicate token handling
func TestSessionRepository_EdgeCases_DuplicateRefreshToken(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	token := testhelpers.AlphaNumericString(32, 32)

	session1 := testhelpers.SessionFixture().
		WithUserID(userID).
		WithRefreshToken(token).
		ToSession()

	err := repo.Create(ctx, session1)
	require.NoError(t, err)

	session2 := testhelpers.SessionFixture().
		WithUserID(userID).
		WithRefreshToken(token).
		ToSession()

	err = repo.Create(ctx, session2)
	require.Error(t, err, "Duplicate refresh token should fail")
}

// TestSessionRepository_EdgeCases_DeleteActiveSession tests deleting active session
func TestSessionRepository_EdgeCases_DeleteActiveSession(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	token := testhelpers.AlphaNumericString(32, 32)

	session := testhelpers.SessionFixture().
		WithUserID(userID).
		WithRefreshToken(token).
		ToSession()

	err := repo.Create(ctx, session)
	require.NoError(t, err)

	err = repo.Delete(ctx, session.ID)
	require.NoError(t, err)

	// Should not be retrievable after delete
	retrieved, err := repo.FindByRefreshToken(ctx, token)
	require.NoError(t, err)
	require.Nil(t, retrieved, "Deleted session should not be retrievable")
}
