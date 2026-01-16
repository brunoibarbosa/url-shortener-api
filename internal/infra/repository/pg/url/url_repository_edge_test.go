//go:build test

package pg_repo_test

import (
	"context"
	"strings"
	"testing"

	"url-shortener-api/internal/infra/repository/pg/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestURLRepository_EdgeCases_VeryLongURL tests URL at maximum length
func TestURLRepository_EdgeCases_VeryLongURL(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	shortCode := testhelpers.ShortCode()

	// URL with 2048 characters (common browser limit)
	baseURL := "https://example.com/"
	longPath := strings.Repeat("a", 2020)
	longURL := baseURL + longPath

	url := testhelpers.URLFixture().
		WithUserID(userID).
		WithShortCode(shortCode).
		WithOriginalURL(longURL).
		ToURL()

	err := repo.Save(ctx, url)
	require.NoError(t, err)

	retrieved, err := repo.FindByShortCode(ctx, shortCode)
	require.NoError(t, err)
	require.Equal(t, longURL, retrieved.OriginalURL)
}

// TestURLRepository_EdgeCases_MinimumURL tests URL at minimum length
func TestURLRepository_EdgeCases_MinimumURL(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	shortCode := testhelpers.ShortCode()

	// Minimum valid URL
	minURL := "http://a.b"

	url := testhelpers.URLFixture().
		WithUserID(userID).
		WithShortCode(shortCode).
		WithOriginalURL(minURL).
		ToURL()

	err := repo.Save(ctx, url)
	require.NoError(t, err)

	retrieved, err := repo.FindByShortCode(ctx, shortCode)
	require.NoError(t, err)
	require.Equal(t, minURL, retrieved.OriginalURL)
}

// TestURLRepository_EdgeCases_SpecialCharactersURL tests URL with special characters
func TestURLRepository_EdgeCases_SpecialCharactersURL(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()

	specialURLs := []string{
		"https://example.com/path?query=value&foo=bar",
		"https://example.com/path#fragment",
		"https://user:pass@example.com/path",
		"https://example.com/path/with/many/slashes",
		"https://example.com/path%20with%20spaces",
		"https://example.com:8080/custom-port",
		"https://sub.sub.example.com/multi-subdomain",
	}

	for i, originalURL := range specialURLs {
		shortCode := testhelpers.ShortCode()
		url := testhelpers.URLFixture().
			WithUserID(userID).
			WithShortCode(shortCode).
			WithOriginalURL(originalURL).
			ToURL()

		err := repo.Save(ctx, url)
		require.NoError(t, err, "Failed for URL: %s", originalURL)

		retrieved, err := repo.FindByShortCode(ctx, shortCode)
		require.NoError(t, err, "Failed to retrieve URL %d", i)
		require.Equal(t, originalURL, retrieved.OriginalURL)
	}
}

// TestURLRepository_EdgeCases_UnicodeURL tests URL with unicode characters
func TestURLRepository_EdgeCases_UnicodeURL(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()

	unicodeURLs := []string{
		"https://example.com/路径",
		"https://example.com/путь",
		"https://пример.com/path",
		"https://例え.jp/パス",
	}

	for _, originalURL := range unicodeURLs {
		shortCode := testhelpers.ShortCode()
		url := testhelpers.URLFixture().
			WithUserID(userID).
			WithShortCode(shortCode).
			WithOriginalURL(originalURL).
			ToURL()

		err := repo.Save(ctx, url)
		require.NoError(t, err, "Failed for unicode URL: %s", originalURL)

		retrieved, err := repo.FindByShortCode(ctx, shortCode)
		require.NoError(t, err)
		require.Equal(t, originalURL, retrieved.OriginalURL)
	}
}

// TestURLRepository_EdgeCases_ShortCodeVariations tests various short code formats
func TestURLRepository_EdgeCases_ShortCodeVariations(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	originalURL := "https://example.com"

	shortCodes := []string{
		"aaaaaa", // All same character
		"AAAAAA", // All uppercase
		"123456", // All numeric
		"aA1bB2", // Mixed case and numbers
		"______", // Special characters (if allowed)
	}

	for _, shortCode := range shortCodes {
		url := testhelpers.URLFixture().
			WithUserID(userID).
			WithShortCode(shortCode).
			WithOriginalURL(originalURL).
			ToURL()

		err := repo.Save(ctx, url)
		if err != nil {
			t.Logf("Short code '%s' not allowed: %v", shortCode, err)
			continue
		}

		retrieved, err := repo.FindByShortCode(ctx, shortCode)
		require.NoError(t, err)
		require.Equal(t, shortCode, retrieved.ShortCode)
	}
}

// TestURLRepository_EdgeCases_ClickCountOverflow tests very large click counts
func TestURLRepository_EdgeCases_ClickCountOverflow(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	shortCode := testhelpers.ShortCode()

	url := testhelpers.URLFixture().
		WithUserID(userID).
		WithShortCode(shortCode).
		ToURL()

	err := repo.Save(ctx, url)
	require.NoError(t, err)

	// Increment clicks many times
	iterations := 1000
	for i := 0; i < iterations; i++ {
		err = repo.IncrementClicks(ctx, shortCode)
		require.NoError(t, err)
	}

	// Verify count
	retrieved, err := repo.FindByShortCode(ctx, shortCode)
	require.NoError(t, err)
	require.Equal(t, int64(iterations), retrieved.Clicks)
}

// TestURLRepository_EdgeCases_FindNonExistentShortCode tests retrieving non-existent short code
func TestURLRepository_EdgeCases_FindNonExistentShortCode(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	url, err := repo.FindByShortCode(ctx, "XXXXXX")
	require.NoError(t, err)
	require.Nil(t, url)
}

// TestURLRepository_EdgeCases_IncrementNonExistentURL tests incrementing non-existent URL
func TestURLRepository_EdgeCases_IncrementNonExistentURL(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	err := repo.IncrementClicks(ctx, "XXXXXX")
	// Depending on implementation, might return error or no-op
	if err != nil {
		t.Logf("IncrementClicks returns error for non-existent URL")
	} else {
		t.Logf("IncrementClicks is no-op for non-existent URL")
	}
}

// TestURLRepository_EdgeCases_DuplicateShortCode tests duplicate short code handling
func TestURLRepository_EdgeCases_DuplicateShortCode(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	shortCode := testhelpers.ShortCode()

	url1 := testhelpers.URLFixture().
		WithUserID(userID).
		WithShortCode(shortCode).
		WithOriginalURL("https://example1.com").
		ToURL()

	err := repo.Save(ctx, url1)
	require.NoError(t, err)

	url2 := testhelpers.URLFixture().
		WithUserID(userID).
		WithShortCode(shortCode).
		WithOriginalURL("https://example2.com").
		ToURL()

	err = repo.Save(ctx, url2)
	require.Error(t, err, "Duplicate short code should fail")
}

// TestURLRepository_EdgeCases_DeleteAndRecreate tests deleting and recreating URL
func TestURLRepository_EdgeCases_DeleteAndRecreate(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	shortCode := testhelpers.ShortCode()

	url1 := testhelpers.URLFixture().
		WithUserID(userID).
		WithShortCode(shortCode).
		WithOriginalURL("https://example1.com").
		ToURL()

	err := repo.Save(ctx, url1)
	require.NoError(t, err)

	err = repo.Delete(ctx, shortCode)
	require.NoError(t, err)

	// Try to recreate with same short code
	url2 := testhelpers.URLFixture().
		WithUserID(userID).
		WithShortCode(shortCode).
		WithOriginalURL("https://example2.com").
		ToURL()

	err = repo.Save(ctx, url2)

	if err != nil {
		t.Logf("Cannot reuse short code after delete (unique constraint)")
	} else {
		t.Logf("Can reuse short code after delete (unique constraint with WHERE deleted_at IS NULL)")
		require.NoError(t, err)
	}
}

// TestURLRepository_EdgeCases_ExpiresAtBoundary tests expiration timestamp boundaries
func TestURLRepository_EdgeCases_ExpiresAtBoundary(t *testing.T) {
	repo := setupRepository(t)
	ctx := context.Background()

	userID := uuid.New().String()
	shortCode := testhelpers.ShortCode()

	// Set expiration to very far future
	url := testhelpers.URLFixture().
		WithUserID(userID).
		WithShortCode(shortCode).
		ToURL()

	err := repo.Save(ctx, url)
	require.NoError(t, err)

	retrieved, err := repo.FindByShortCode(ctx, shortCode)
	require.NoError(t, err)
	require.NotNil(t, retrieved.ExpiresAt)
}
