package cache_test

import (
	"context"
	"os"
	"testing"

	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	cache "github.com/brunoibarbosa/url-shortener/internal/infra/repository/redis/session"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	sharedRedisClient *redis.Client
	sharedContainer   testcontainers.Container
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

	sharedContainer = container

	host, err := container.Host(ctx)
	if err != nil {
		panic(err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		panic(err)
	}

	sharedRedisClient = redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})

	if err := sharedRedisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	code := m.Run()

	sharedRedisClient.Close()
	container.Terminate(ctx)

	os.Exit(code)
}

func cleanRedis(t *testing.T) {
	ctx := context.Background()
	err := sharedRedisClient.FlushDB(ctx).Err()
	require.NoError(t, err)
}

func TestStateRepository_GenerateState_Success(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	state, err := repo.GenerateState(ctx)

	require.NoError(t, err)
	assert.NotEmpty(t, state)
	assert.Len(t, state, 44) // State should be 44 characters (base64 of 32 bytes)
}

func TestStateRepository_GenerateState_Unique(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	states := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		state, err := repo.GenerateState(ctx)
		require.NoError(t, err)
		assert.False(t, states[state], "Generated duplicate state")
		states[state] = true
	}

	assert.Len(t, states, iterations)
}

func TestStateRepository_ValidateState_Valid(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	state, err := repo.GenerateState(ctx)
	require.NoError(t, err)

	err = repo.ValidateState(ctx, state)

	assert.NoError(t, err)
}

func TestStateRepository_ValidateState_Invalid(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	err := repo.ValidateState(ctx, "invalid-state-12345")

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrInvalidState, err)
}

func TestStateRepository_ValidateState_Empty(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	err := repo.ValidateState(ctx, "")

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrInvalidState, err)
}

func TestStateRepository_ValidateState_Consumed(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	state, err := repo.GenerateState(ctx)
	require.NoError(t, err)

	// First validation should succeed
	err = repo.ValidateState(ctx, state)
	require.NoError(t, err)

	// Delete the state (simulate consumption)
	err = repo.DeleteState(ctx, state)
	require.NoError(t, err)

	// Second validation should fail (state consumed)
	err = repo.ValidateState(ctx, state)
	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrInvalidState, err)
}

func TestStateRepository_ValidateState_Expiration(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	// Generate state
	state, err := repo.GenerateState(ctx)
	require.NoError(t, err)

	// State should be valid immediately
	err = repo.ValidateState(ctx, state)
	require.NoError(t, err)

	// Note: We can't easily test expiration without waiting 2 minutes
	// or mocking time, so we'll skip the actual expiration test
	// In real scenario, state expires after 2 minutes
}

func TestStateRepository_KeyFormat(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	state, err := repo.GenerateState(ctx)
	require.NoError(t, err)

	// Verify key format in Redis
	expectedKey := "oauth:state:" + state
	exists, err := sharedRedisClient.Exists(ctx, expectedKey).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)
}

func TestStateRepository_ConcurrentGenerate(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	iterations := 50
	results := make(chan string, iterations)
	errors := make(chan error, iterations)

	for i := 0; i < iterations; i++ {
		go func() {
			state, err := repo.GenerateState(ctx)
			if err != nil {
				errors <- err
			} else {
				results <- state
			}
		}()
	}

	// Collect results
	states := make(map[string]bool)
	for i := 0; i < iterations; i++ {
		select {
		case state := <-results:
			assert.NotEmpty(t, state)
			states[state] = true
		case err := <-errors:
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	// All states should be unique
	assert.Len(t, states, iterations)
}

func TestStateRepository_MultipleValidations(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	// Generate multiple states
	states := make([]string, 5)
	for i := 0; i < 5; i++ {
		state, err := repo.GenerateState(ctx)
		require.NoError(t, err)
		states[i] = state
	}

	// Validate all states
	for _, state := range states {
		err := repo.ValidateState(ctx, state)
		assert.NoError(t, err)
	}

	// Delete all states (simulate consumption)
	for _, state := range states {
		err := repo.DeleteState(ctx, state)
		require.NoError(t, err)
	}

	// All states should now be consumed
	for _, state := range states {
		err := repo.ValidateState(ctx, state)
		assert.Error(t, err)
	}
}

func TestStateRepository_InvalidateAfterUse(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	state, err := repo.GenerateState(ctx)
	require.NoError(t, err)

	// Use the state (validate and delete)
	err = repo.ValidateState(ctx, state)
	require.NoError(t, err)

	err = repo.DeleteState(ctx, state)
	require.NoError(t, err)

	// Try to validate again - should fail
	err = repo.ValidateState(ctx, state)
	assert.Error(t, err)
}

func TestStateRepository_DifferentStatesIndependent(t *testing.T) {
	cleanRedis(t)

	repo := cache.NewStateRepository(sharedRedisClient)
	ctx := context.Background()

	state1, err := repo.GenerateState(ctx)
	require.NoError(t, err)

	state2, err := repo.GenerateState(ctx)
	require.NoError(t, err)

	// Validate and consume state1
	err = repo.ValidateState(ctx, state1)
	require.NoError(t, err)

	// state2 should still be valid
	err = repo.ValidateState(ctx, state2)
	assert.NoError(t, err)
}
