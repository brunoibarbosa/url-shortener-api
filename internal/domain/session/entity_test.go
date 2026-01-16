package session_test

import (
	"testing"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSession_IsExpired(t *testing.T) {
	t.Run("should return false when session is active and not expired", func(t *testing.T) {
		expiresAt := time.Now().Add(1 * time.Hour)
		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        nil,
		}

		expired := session.IsExpired()

		assert.False(t, expired)
	})

	t.Run("should return true when session is expired", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour)
		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        nil,
		}

		expired := session.IsExpired()

		assert.True(t, expired)
	})

	t.Run("should return true when session is revoked", func(t *testing.T) {
		expiresAt := time.Now().Add(1 * time.Hour)
		revokedAt := time.Now().Add(-30 * time.Minute)
		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        &revokedAt,
		}

		expired := session.IsExpired()

		assert.True(t, expired)
	})

	t.Run("should return true when session is both expired and revoked", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour)
		revokedAt := time.Now().Add(-30 * time.Minute)
		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        &revokedAt,
		}

		expired := session.IsExpired()

		assert.True(t, expired)
	})
}

func TestSession_Creation(t *testing.T) {
	t.Run("should create session with valid data", func(t *testing.T) {
		id := uuid.New()
		userID := uuid.New()
		hash := "refresh_token_hash"
		userAgent := "Mozilla/5.0"
		ipAddress := "192.168.1.1"
		expiresAt := time.Now().Add(7 * 24 * time.Hour)

		session := &domain.Session{
			ID:               id,
			UserID:           userID,
			RefreshTokenHash: hash,
			UserAgent:        userAgent,
			IPAddress:        ipAddress,
			ExpiresAt:        &expiresAt,
			RevokedAt:        nil,
		}

		assert.Equal(t, id, session.ID)
		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, hash, session.RefreshTokenHash)
		assert.Equal(t, userAgent, session.UserAgent)
		assert.Equal(t, ipAddress, session.IPAddress)
		assert.NotNil(t, session.ExpiresAt)
		assert.Nil(t, session.RevokedAt)
	})
}

func TestSession_Revoked(t *testing.T) {
	t.Run("should identify revoked session", func(t *testing.T) {
		expiresAt := time.Now().Add(1 * time.Hour)
		revokedAt := time.Now()

		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        &revokedAt,
		}

		assert.NotNil(t, session.RevokedAt)
		assert.True(t, session.IsExpired())
	})

	t.Run("should identify active session", func(t *testing.T) {
		expiresAt := time.Now().Add(1 * time.Hour)

		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        nil,
		}

		assert.Nil(t, session.RevokedAt)
		assert.False(t, session.IsExpired())
	})
}
