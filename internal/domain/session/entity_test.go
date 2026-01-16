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

func TestSession_IsActive(t *testing.T) {
	t.Run("should return true for active non-expired session", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        nil,
		}

		isActive := !session.IsExpired()

		assert.True(t, isActive)
	})

	t.Run("should return false for expired session", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour)
		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        nil,
		}

		isActive := !session.IsExpired()

		assert.False(t, isActive)
	})

	t.Run("should return false for revoked session", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		revokedAt := time.Now()
		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        &revokedAt,
		}

		isActive := !session.IsExpired()

		assert.False(t, isActive)
	})
}

func TestSession_RemainingTime(t *testing.T) {
	t.Run("should calculate remaining time correctly", func(t *testing.T) {
		expiresAt := time.Now().Add(2 * time.Hour)
		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        nil,
		}

		remaining := time.Until(*session.ExpiresAt)

		assert.True(t, remaining > 0)
		assert.True(t, remaining <= 2*time.Hour)
	})

	t.Run("should return negative time for expired session", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour)
		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        nil,
		}

		remaining := time.Until(*session.ExpiresAt)

		assert.True(t, remaining < 0)
	})
}

func TestSession_Lifecycle(t *testing.T) {
	t.Run("should transition from active to revoked", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		session := &domain.Session{
			ID:               uuid.New(),
			UserID:           uuid.New(),
			RefreshTokenHash: "hash123",
			ExpiresAt:        &expiresAt,
			RevokedAt:        nil,
		}

		assert.False(t, session.IsExpired())
		assert.Nil(t, session.RevokedAt)

		revokedAt := time.Now()
		session.RevokedAt = &revokedAt

		assert.True(t, session.IsExpired())
		assert.NotNil(t, session.RevokedAt)
	})
}

func TestSession_WithDifferentDevices(t *testing.T) {
	t.Run("should store different user agents and IPs", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		userID := uuid.New()

		session1 := &domain.Session{
			ID:               uuid.New(),
			UserID:           userID,
			RefreshTokenHash: "hash123",
			UserAgent:        "Mozilla/5.0",
			IPAddress:        "192.168.1.1",
			ExpiresAt:        &expiresAt,
		}

		session2 := &domain.Session{
			ID:               uuid.New(),
			UserID:           userID,
			RefreshTokenHash: "hash456",
			UserAgent:        "Chrome Mobile",
			IPAddress:        "192.168.1.2",
			ExpiresAt:        &expiresAt,
		}

		assert.NotEqual(t, session1.UserAgent, session2.UserAgent)
		assert.NotEqual(t, session1.IPAddress, session2.IPAddress)
		assert.Equal(t, session1.UserID, session2.UserID)
	})
}

func TestSession_Errors(t *testing.T) {
	t.Run("should have all error constants defined", func(t *testing.T) {
		assert.Equal(t, "session not found", domain.ErrNotFound.Error())
		assert.Equal(t, "invalid or expired refresh token", domain.ErrInvalidRefreshToken.Error())
		assert.Equal(t, "failed to revoke token", domain.ErrRevokeFailed.Error())
		assert.Equal(t, "invalid OAuth code", domain.ErrInvalidOAuthCode.Error())
	})
}
