package session_test

import (
	"testing"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomState(t *testing.T) {
	t.Run("should generate random state successfully", func(t *testing.T) {
		state, err := domain.GenerateRandomState()

		assert.NoError(t, err)
		assert.NotEmpty(t, state)
		assert.Greater(t, len(state), 40) // base64 encoded 32 bytes should be > 40 chars
	})

	t.Run("should generate unique states", func(t *testing.T) {
		state1, err1 := domain.GenerateRandomState()
		state2, err2 := domain.GenerateRandomState()

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, state1, state2)
	})

	t.Run("should generate valid base64 URL encoded string", func(t *testing.T) {
		state, err := domain.GenerateRandomState()

		assert.NoError(t, err)
		// Base64 URL encoding should not contain '+' or '/'
		assert.NotContains(t, state, "+")
		assert.NotContains(t, state, "/")
	})
}

func TestStateService_Errors(t *testing.T) {
	t.Run("should have ErrInvalidState constant", func(t *testing.T) {
		assert.Equal(t, "invalid or expired state", domain.ErrInvalidState.Error())
	})

	t.Run("should have ErrStateGeneration constant", func(t *testing.T) {
		assert.Equal(t, "failed to generate state", domain.ErrStateGeneration.Error())
	})
}

func TestGenerateRandomState_EdgeCases(t *testing.T) {
	t.Run("should generate states with consistent length", func(t *testing.T) {
		states := make([]string, 10)
		for i := 0; i < 10; i++ {
			state, err := domain.GenerateRandomState()
			assert.NoError(t, err)
			states[i] = state
		}

		// All base64 encoded 32-byte values should have similar lengths (with padding)
		firstLen := len(states[0])
		for _, state := range states {
			assert.InDelta(t, firstLen, len(state), 2) // allow small variance due to padding
		}
	})
}
