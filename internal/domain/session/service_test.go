package session_test

import (
	"testing"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOAuthUser_Creation(t *testing.T) {
	t.Run("should create OAuthUser with all fields", func(t *testing.T) {
		avatarURL := "https://example.com/avatar.jpg"
		user := &domain.OAuthUser{
			ID:           "google_123456",
			Name:         "Test User",
			Email:        "test@example.com",
			AccessToken:  "access_token_123",
			RefreshToken: "refresh_token_456",
			AvatarURL:    &avatarURL,
		}

		assert.Equal(t, "google_123456", user.ID)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "access_token_123", user.AccessToken)
		assert.Equal(t, "refresh_token_456", user.RefreshToken)
		assert.NotNil(t, user.AvatarURL)
		assert.Equal(t, avatarURL, *user.AvatarURL)
	})

	t.Run("should create OAuthUser without avatar", func(t *testing.T) {
		user := &domain.OAuthUser{
			ID:           "google_789",
			Name:         "User Without Avatar",
			Email:        "noavatar@example.com",
			AccessToken:  "token_abc",
			RefreshToken: "refresh_def",
			AvatarURL:    nil,
		}

		assert.Nil(t, user.AvatarURL)
		assert.NotEmpty(t, user.ID)
		assert.NotEmpty(t, user.Email)
	})
}

func TestTokenParams_Creation(t *testing.T) {
	t.Run("should create TokenParams with valid data", func(t *testing.T) {
		userID := uuid.New()
		sessionID := uuid.New()
		duration := 24 * time.Hour

		params := &domain.TokenParams{
			UserID:    userID,
			SessionID: sessionID,
			Duration:  duration,
		}

		assert.Equal(t, userID, params.UserID)
		assert.Equal(t, sessionID, params.SessionID)
		assert.Equal(t, duration, params.Duration)
	})

	t.Run("should create TokenParams with different durations", func(t *testing.T) {
		shortDuration := 15 * time.Minute
		longDuration := 30 * 24 * time.Hour

		shortParams := &domain.TokenParams{
			UserID:    uuid.New(),
			SessionID: uuid.New(),
			Duration:  shortDuration,
		}

		longParams := &domain.TokenParams{
			UserID:    uuid.New(),
			SessionID: uuid.New(),
			Duration:  longDuration,
		}

		assert.Equal(t, shortDuration, shortParams.Duration)
		assert.Equal(t, longDuration, longParams.Duration)
		assert.NotEqual(t, shortParams.Duration, longParams.Duration)
	})
}

func TestTokenClaims_Creation(t *testing.T) {
	t.Run("should create TokenClaims with all fields", func(t *testing.T) {
		now := time.Now().Unix()
		claims := &domain.TokenClaims{
			Sub: "user_123",
			Sid: "session_456",
			Exp: now + 3600,
			Iat: now,
		}

		assert.Equal(t, "user_123", claims.Sub)
		assert.Equal(t, "session_456", claims.Sid)
		assert.Equal(t, now+3600, claims.Exp)
		assert.Equal(t, now, claims.Iat)
	})

	t.Run("should verify token expiration logic", func(t *testing.T) {
		now := time.Now().Unix()

		expiredClaims := &domain.TokenClaims{
			Sub: "user_123",
			Sid: "session_456",
			Exp: now - 3600, // expired 1 hour ago
			Iat: now - 7200, // issued 2 hours ago
		}

		validClaims := &domain.TokenClaims{
			Sub: "user_789",
			Sid: "session_789",
			Exp: now + 3600, // expires in 1 hour
			Iat: now,
		}

		assert.True(t, time.Now().Unix() > expiredClaims.Exp)
		assert.True(t, time.Now().Unix() < validClaims.Exp)
	})
}

func TestTokenService_Errors(t *testing.T) {
	t.Run("should have ErrTokenGenerate constant", func(t *testing.T) {
		assert.Equal(t, "failed to generate token", domain.ErrTokenGenerate.Error())
	})
}
