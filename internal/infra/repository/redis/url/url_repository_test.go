package redis_repo_test

import (
	"context"
	"os"
	"testing"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	redis_repo "github.com/brunoibarbosa/url-shortener/internal/infra/repository/redis/url"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	urlRedisClient *redis.Client
	urlContainer   testcontainers.Container
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	urlContainer = container

	host, err := container.Host(ctx)
	if err != nil {
		panic(err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		panic(err)
	}

	urlRedisClient = redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})

	if err := urlRedisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	code := m.Run()

	urlRedisClient.Close()
	container.Terminate(ctx)

	os.Exit(code)
}

func cleanURLRedis(t *testing.T) {
	ctx := context.Background()
	err := urlRedisClient.FlushDB(ctx).Err()
	require.NoError(t, err)
}

func TestURLCacheRepository_Save_Success(t *testing.T) {
	cleanURLRedis(t)

	repo := redis_repo.NewURLCacheRepository(urlRedisClient)
	ctx := context.Background()

	url := &domain.URL{
		ShortCode:    "abc123",
		EncryptedURL: "encrypted-url-data",
		ExpiresAt:    nil,
	}

	err := repo.Save(ctx, url, 5*time.Minute)

	require.NoError(t, err)
}

func TestURLCacheRepository_FindByShortCode_Success(t *testing.T) {
	cleanURLRedis(t)

	repo := redis_repo.NewURLCacheRepository(urlRedisClient)
	ctx := context.Background()

	shortCode := "abc123"
	encryptedURL := "encrypted-url-data"

	url := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: encryptedURL,
		ExpiresAt:    nil,
	}

	err := repo.Save(ctx, url, 5*time.Minute)
	require.NoError(t, err)

	foundURL, err := repo.FindByShortCode(ctx, shortCode)

	require.NoError(t, err)
	assert.NotNil(t, foundURL)
	assert.Equal(t, shortCode, foundURL.ShortCode)
	assert.Equal(t, encryptedURL, foundURL.EncryptedURL)
	assert.Nil(t, foundURL.ExpiresAt)
}

func TestURLCacheRepository_FindByShortCode_NotFound(t *testing.T) {
	cleanURLRedis(t)

	repo := redis_repo.NewURLCacheRepository(urlRedisClient)
	ctx := context.Background()

	foundURL, err := repo.FindByShortCode(ctx, "nonexistent")

	require.NoError(t, err)
	assert.Nil(t, foundURL)
}

func TestURLCacheRepository_Exists_True(t *testing.T) {
	cleanURLRedis(t)

	repo := redis_repo.NewURLCacheRepository(urlRedisClient)
	ctx := context.Background()

	url := &domain.URL{
		ShortCode:    "exists123",
		EncryptedURL: "encrypted-data",
		ExpiresAt:    nil,
	}

	err := repo.Save(ctx, url, 5*time.Minute)
	require.NoError(t, err)

	exists, err := repo.Exists(ctx, "exists123")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestURLCacheRepository_Exists_False(t *testing.T) {
	cleanURLRedis(t)

	repo := redis_repo.NewURLCacheRepository(urlRedisClient)
	ctx := context.Background()

	exists, err := repo.Exists(ctx, "nonexistent")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestURLCacheRepository_Delete_Success(t *testing.T) {
	cleanURLRedis(t)

	repo := redis_repo.NewURLCacheRepository(urlRedisClient)
	ctx := context.Background()

	url := &domain.URL{
		ShortCode:    "delete123",
		EncryptedURL: "encrypted-data",
		ExpiresAt:    nil,
	}

	err := repo.Save(ctx, url, 5*time.Minute)
	require.NoError(t, err)

	err = repo.Delete(ctx, "delete123")
	require.NoError(t, err)

	exists, err := repo.Exists(ctx, "delete123")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestURLCacheRepository_Expiration(t *testing.T) {
	cleanURLRedis(t)

	repo := redis_repo.NewURLCacheRepository(urlRedisClient)
	ctx := context.Background()

	url := &domain.URL{
		ShortCode:    "expire123",
		EncryptedURL: "encrypted-data",
		ExpiresAt:    nil,
	}

	// Save with 1 second expiration
	err := repo.Save(ctx, url, 1*time.Second)
	require.NoError(t, err)

	// Should exist immediately
	exists, err := repo.Exists(ctx, "expire123")
	require.NoError(t, err)
	assert.True(t, exists)

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Should not exist after expiration
	exists, err = repo.Exists(ctx, "expire123")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestURLCacheRepository_SaveOverwrite(t *testing.T) {
	cleanURLRedis(t)

	repo := redis_repo.NewURLCacheRepository(urlRedisClient)
	ctx := context.Background()

	shortCode := "overwrite123"

	url1 := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: "first-encrypted-data",
		ExpiresAt:    nil,
	}

	url2 := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: "second-encrypted-data",
		ExpiresAt:    nil,
	}

	// Save first URL
	err := repo.Save(ctx, url1, 5*time.Minute)
	require.NoError(t, err)

	// Save second URL with same shortCode (should overwrite)
	err = repo.Save(ctx, url2, 5*time.Minute)
	require.NoError(t, err)

	// Should retrieve the second URL
	foundURL, err := repo.FindByShortCode(ctx, shortCode)
	require.NoError(t, err)
	assert.Equal(t, "second-encrypted-data", foundURL.EncryptedURL)
}

func TestURLCacheRepository_MultipleURLs(t *testing.T) {
	cleanURLRedis(t)

	repo := redis_repo.NewURLCacheRepository(urlRedisClient)
	ctx := context.Background()

	urls := []*domain.URL{
		{ShortCode: "url1", EncryptedURL: "encrypted1"},
		{ShortCode: "url2", EncryptedURL: "encrypted2"},
		{ShortCode: "url3", EncryptedURL: "encrypted3"},
	}

	// Save all URLs
	for _, url := range urls {
		err := repo.Save(ctx, url, 5*time.Minute)
		require.NoError(t, err)
	}

	// Verify all exist
	for _, url := range urls {
		foundURL, err := repo.FindByShortCode(ctx, url.ShortCode)
		require.NoError(t, err)
		assert.Equal(t, url.EncryptedURL, foundURL.EncryptedURL)
	}
}

func TestURLCacheRepository_KeyFormat(t *testing.T) {
	cleanURLRedis(t)

	repo := redis_repo.NewURLCacheRepository(urlRedisClient)
	ctx := context.Background()

	shortCode := "test123"
	url := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: "encrypted-data",
		ExpiresAt:    nil,
	}

	err := repo.Save(ctx, url, 5*time.Minute)
	require.NoError(t, err)

	// Verify key format in Redis
	expectedKey := "url:short_code:test123"
	exists, err := urlRedisClient.Exists(ctx, expectedKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)
}
