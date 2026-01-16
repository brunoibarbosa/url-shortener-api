package cache_test

import (
	"context"
	"testing"
	"time"

	cache "github.com/brunoibarbosa/url-shortener/internal/infra/repository/redis/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: TestMain is defined in state_repository_test.go for shared container setup
// Both blacklist and state tests use the sharedRedisClient from that file

func cleanBlacklistRedis(t *testing.T) {
	ctx := context.Background()
	err := sharedRedisClient.FlushDB(ctx).Err()
	require.NoError(t, err)
}

func TestBlacklistRepository_Revoke_Success(t *testing.T) {
	cleanBlacklistRedis(t)

	repo := cache.NewBlacklistRepository(sharedRedisClient)
	ctx := context.Background()

	token := "test-token-123"

	err := repo.Revoke(ctx, token, 5*time.Minute)

	require.NoError(t, err)
}

func TestBlacklistRepository_IsRevoked_True(t *testing.T) {
	cleanBlacklistRedis(t)

	repo := cache.NewBlacklistRepository(sharedRedisClient)
	ctx := context.Background()

	token := "revoked-token"

	err := repo.Revoke(ctx, token, 5*time.Minute)
	require.NoError(t, err)

	isRevoked, err := repo.IsRevoked(ctx, token)

	require.NoError(t, err)
	assert.True(t, isRevoked)
}

func TestBlacklistRepository_IsRevoked_False(t *testing.T) {
	cleanBlacklistRedis(t)

	repo := cache.NewBlacklistRepository(sharedRedisClient)
	ctx := context.Background()

	isRevoked, err := repo.IsRevoked(ctx, "nonexistent-token")

	require.NoError(t, err)
	assert.False(t, isRevoked)
}

func TestBlacklistRepository_Expiration(t *testing.T) {
	cleanBlacklistRedis(t)

	repo := cache.NewBlacklistRepository(sharedRedisClient)
	ctx := context.Background()

	token := "expiring-token"

	// Revoke with 1 second expiration
	err := repo.Revoke(ctx, token, 1*time.Second)
	require.NoError(t, err)

	// Should be revoked immediately
	isRevoked, err := repo.IsRevoked(ctx, token)
	require.NoError(t, err)
	assert.True(t, isRevoked)

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Should not be revoked after expiration
	isRevoked, err = repo.IsRevoked(ctx, token)
	require.NoError(t, err)
	assert.False(t, isRevoked)
}

func TestBlacklistRepository_MultipleTokens(t *testing.T) {
	cleanBlacklistRedis(t)

	repo := cache.NewBlacklistRepository(sharedRedisClient)
	ctx := context.Background()

	tokens := []string{"token1", "token2", "token3"}

	// Revoke all tokens
	for _, token := range tokens {
		err := repo.Revoke(ctx, token, 5*time.Minute)
		require.NoError(t, err)
	}

	// Verify all are revoked
	for _, token := range tokens {
		isRevoked, err := repo.IsRevoked(ctx, token)
		require.NoError(t, err)
		assert.True(t, isRevoked, "Token %s should be revoked", token)
	}
}

func TestBlacklistRepository_RevokeOverwrite(t *testing.T) {
	cleanBlacklistRedis(t)

	repo := cache.NewBlacklistRepository(sharedRedisClient)
	ctx := context.Background()

	token := "overwrite-token"

	// Revoke with short expiration
	err := repo.Revoke(ctx, token, 1*time.Second)
	require.NoError(t, err)

	// Revoke again with longer expiration
	err = repo.Revoke(ctx, token, 10*time.Minute)
	require.NoError(t, err)

	// Wait for original expiration time
	time.Sleep(2 * time.Second)

	// Should still be revoked (longer expiration)
	isRevoked, err := repo.IsRevoked(ctx, token)
	require.NoError(t, err)
	assert.True(t, isRevoked)
}

func TestBlacklistRepository_KeyFormat(t *testing.T) {
	cleanBlacklistRedis(t)

	repo := cache.NewBlacklistRepository(sharedRedisClient)
	ctx := context.Background()

	token := "test-token"

	err := repo.Revoke(ctx, token, 5*time.Minute)
	require.NoError(t, err)

	// Verify key format in Redis
	expectedKey := "session:revoked:test-token"
	exists, err := sharedRedisClient.Exists(ctx, expectedKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)
}

func TestBlacklistRepository_LongToken(t *testing.T) {
	cleanBlacklistRedis(t)

	repo := cache.NewBlacklistRepository(sharedRedisClient)
	ctx := context.Background()

	// Very long token
	longToken := ""
	for i := 0; i < 1000; i++ {
		longToken += "a"
	}

	err := repo.Revoke(ctx, longToken, 5*time.Minute)
	require.NoError(t, err)

	isRevoked, err := repo.IsRevoked(ctx, longToken)
	require.NoError(t, err)
	assert.True(t, isRevoked)
}

func TestBlacklistRepository_SpecialCharacters(t *testing.T) {
	cleanBlacklistRedis(t)

	repo := cache.NewBlacklistRepository(sharedRedisClient)
	ctx := context.Background()

	tokens := []string{
		"token-with-dashes",
		"token_with_underscores",
		"token.with.dots",
		"token:with:colons",
		"token/with/slashes",
	}

	for _, token := range tokens {
		t.Run(token, func(t *testing.T) {
			err := repo.Revoke(ctx, token, 5*time.Minute)
			require.NoError(t, err)

			isRevoked, err := repo.IsRevoked(ctx, token)
			require.NoError(t, err)
			assert.True(t, isRevoked)
		})
	}
}

func TestBlacklistRepository_ConcurrentRevoke(t *testing.T) {
	cleanBlacklistRedis(t)

	repo := cache.NewBlacklistRepository(sharedRedisClient)
	ctx := context.Background()

	iterations := 50
	done := make(chan bool, iterations)

	for i := 0; i < iterations; i++ {
		go func(index int) {
			token := "concurrent-token"
			err := repo.Revoke(ctx, token, 5*time.Minute)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < iterations; i++ {
		<-done
	}

	// Verify token is revoked
	isRevoked, err := repo.IsRevoked(ctx, "concurrent-token")
	require.NoError(t, err)
	assert.True(t, isRevoked)
}
