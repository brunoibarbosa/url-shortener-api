package url_test

import (
	"testing"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestListURLsDTO_Creation(t *testing.T) {
	t.Run("should create ListURLsDTO with all fields", func(t *testing.T) {
		id := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour)
		createdAt := time.Now()
		deletedAt := time.Now()

		dto := domain.ListURLsDTO{
			ID:        id,
			ShortCode: "abc123",
			ExpiresAt: &expiresAt,
			CreatedAt: createdAt,
			DeletedAt: &deletedAt,
		}

		assert.Equal(t, id, dto.ID)
		assert.Equal(t, "abc123", dto.ShortCode)
		assert.NotNil(t, dto.ExpiresAt)
		assert.Equal(t, expiresAt, *dto.ExpiresAt)
		assert.Equal(t, createdAt, dto.CreatedAt)
		assert.NotNil(t, dto.DeletedAt)
	})

	t.Run("should create ListURLsDTO without optional fields", func(t *testing.T) {
		dto := domain.ListURLsDTO{
			ID:        uuid.New(),
			ShortCode: "xyz789",
			ExpiresAt: nil,
			CreatedAt: time.Now(),
			DeletedAt: nil,
		}

		assert.Nil(t, dto.ExpiresAt)
		assert.Nil(t, dto.DeletedAt)
		assert.NotEmpty(t, dto.ShortCode)
	})
}

func TestListURLsSortBy_Constants(t *testing.T) {
	t.Run("should have all sort constants", func(t *testing.T) {
		assert.Equal(t, domain.ListURLsSortBy(0), domain.ListURLsSortByNone)
		assert.Equal(t, domain.ListURLsSortBy(1), domain.ListURLsSortByCreatedAt)
		assert.Equal(t, domain.ListURLsSortBy(2), domain.ListURLsSortByExpiresAt)
	})

	t.Run("should have unique values for each constant", func(t *testing.T) {
		assert.NotEqual(t, domain.ListURLsSortByNone, domain.ListURLsSortByCreatedAt)
		assert.NotEqual(t, domain.ListURLsSortByNone, domain.ListURLsSortByExpiresAt)
		assert.NotEqual(t, domain.ListURLsSortByCreatedAt, domain.ListURLsSortByExpiresAt)
	})
}
