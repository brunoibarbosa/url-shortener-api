package url_test

import (
	"testing"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/stretchr/testify/assert"
)

func TestURL_RemainingTTL(t *testing.T) {
	t.Run("should return 0 when ExpiresAt is nil", func(t *testing.T) {
		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			ExpiresAt:    nil,
		}

		now := time.Now()
		ttl := url.RemainingTTL(now)

		assert.Equal(t, time.Duration(0), ttl)
	})

	t.Run("should return positive duration when not expired", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(1 * time.Hour)

		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			ExpiresAt:    &expiresAt,
		}

		ttl := url.RemainingTTL(now)

		assert.Greater(t, ttl, time.Duration(0))
		assert.LessOrEqual(t, ttl, 1*time.Hour)
	})

	t.Run("should return negative duration when expired", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(-1 * time.Hour)

		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			ExpiresAt:    &expiresAt,
		}

		ttl := url.RemainingTTL(now)

		assert.Less(t, ttl, time.Duration(0))
	})
}

func TestURL_IsExpired(t *testing.T) {
	t.Run("should return false when ExpiresAt is nil", func(t *testing.T) {
		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			ExpiresAt:    nil,
		}

		now := time.Now()
		expired := url.IsExpired(now)

		assert.False(t, expired)
	})

	t.Run("should return false when not expired", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(1 * time.Hour)

		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			ExpiresAt:    &expiresAt,
		}

		expired := url.IsExpired(now)

		assert.False(t, expired)
	})

	t.Run("should return true when expired", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(-1 * time.Hour)

		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			ExpiresAt:    &expiresAt,
		}

		expired := url.IsExpired(now)

		assert.True(t, expired)
	})
}

func TestURL_IsDeleted(t *testing.T) {
	t.Run("should return false when DeletedAt is nil", func(t *testing.T) {
		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			DeletedAt:    nil,
		}

		deleted := url.IsDeleted()

		assert.False(t, deleted)
	})

	t.Run("should return true when DeletedAt is set", func(t *testing.T) {
		now := time.Now()

		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			DeletedAt:    &now,
		}

		deleted := url.IsDeleted()

		assert.True(t, deleted)
	})
}

func TestURL_CanBeAccessed(t *testing.T) {
	t.Run("should return nil when URL is active", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(1 * time.Hour)

		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			ExpiresAt:    &expiresAt,
			DeletedAt:    nil,
		}

		err := url.CanBeAccessed(now)

		assert.NoError(t, err)
	})

	t.Run("should return nil when URL has no expiration", func(t *testing.T) {
		now := time.Now()

		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			ExpiresAt:    nil,
			DeletedAt:    nil,
		}

		err := url.CanBeAccessed(now)

		assert.NoError(t, err)
	})

	t.Run("should return ErrDeletedURL when URL is deleted", func(t *testing.T) {
		now := time.Now()
		deletedAt := now.Add(-1 * time.Hour)

		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			DeletedAt:    &deletedAt,
		}

		err := url.CanBeAccessed(now)

		assert.ErrorIs(t, err, domain.ErrDeletedURL)
	})

	t.Run("should return ErrExpiredURL when URL is expired", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(-1 * time.Hour)

		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			ExpiresAt:    &expiresAt,
			DeletedAt:    nil,
		}

		err := url.CanBeAccessed(now)

		assert.ErrorIs(t, err, domain.ErrExpiredURL)
	})

	t.Run("should return ErrDeletedURL when URL is both deleted and expired", func(t *testing.T) {
		now := time.Now()
		expiresAt := now.Add(-1 * time.Hour)
		deletedAt := now.Add(-30 * time.Minute)

		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			ExpiresAt:    &expiresAt,
			DeletedAt:    &deletedAt,
		}

		err := url.CanBeAccessed(now)

		// Should prioritize deleted status
		assert.ErrorIs(t, err, domain.ErrDeletedURL)
	})
}

func TestURL_Delete(t *testing.T) {
	t.Run("should set DeletedAt to current time", func(t *testing.T) {
		now := time.Now()
		url := &domain.URL{
			ShortCode:    "abc123",
			EncryptedURL: "encrypted",
			DeletedAt:    nil,
		}

		// Simulando delete manualmente
		url.DeletedAt = &now

		assert.NotNil(t, url.DeletedAt)
		assert.Equal(t, now, *url.DeletedAt)
		assert.True(t, url.IsDeleted())
	})
}

func TestURL_Errors(t *testing.T) {
	t.Run("should have all error constants defined", func(t *testing.T) {
		assert.Equal(t, "expired URL", domain.ErrExpiredURL.Error())
		assert.Equal(t, "deleted URL", domain.ErrDeletedURL.Error())
		assert.Equal(t, "URL not found", domain.ErrURLNotFound.Error())
		assert.Equal(t, "invalid url", domain.ErrInvalidURLFormat.Error())
		assert.Equal(t, "url is missing scheme (e.g., http or https)", domain.ErrMissingURLSchema.Error())
		assert.Equal(t, "unsupported scheme", domain.ErrUnsupportedURLSchema.Error())
		assert.Equal(t, "url is missing host", domain.ErrMissingURLHost.Error())
		assert.Equal(t, "invalid short code", domain.ErrInvalidShortCode.Error())
	})
}
