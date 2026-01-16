package session_test

import (
	"testing"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/stretchr/testify/assert"
)

func TestListSessionsDTO_Creation(t *testing.T) {
	t.Run("should create ListSessionsDTO with all fields", func(t *testing.T) {
		now := time.Now()
		expires := now.Add(24 * time.Hour)

		dto := domain.ListSessionsDTO{
			UserAgent: "Mozilla/5.0",
			IPAddress: "192.168.1.1",
			CreatedAt: now,
			ExpiresAt: expires,
		}

		assert.Equal(t, "Mozilla/5.0", dto.UserAgent)
		assert.Equal(t, "192.168.1.1", dto.IPAddress)
		assert.Equal(t, now, dto.CreatedAt)
		assert.Equal(t, expires, dto.ExpiresAt)
	})

	t.Run("should create multiple DTOs with different data", func(t *testing.T) {
		dto1 := domain.ListSessionsDTO{
			UserAgent: "Chrome",
			IPAddress: "10.0.0.1",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}

		dto2 := domain.ListSessionsDTO{
			UserAgent: "Firefox",
			IPAddress: "10.0.0.2",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(2 * time.Hour),
		}

		assert.NotEqual(t, dto1.UserAgent, dto2.UserAgent)
		assert.NotEqual(t, dto1.IPAddress, dto2.IPAddress)
	})
}

func TestListSessionsSortBy_Constants(t *testing.T) {
	t.Run("should have all sort constants", func(t *testing.T) {
		assert.Equal(t, domain.ListSessionsSortBy(0), domain.ListSessionsSortByNone)
		assert.Equal(t, domain.ListSessionsSortBy(1), domain.ListSessionsSortByUserAgent)
		assert.Equal(t, domain.ListSessionsSortBy(2), domain.ListSessionsSortByIPAddress)
		assert.Equal(t, domain.ListSessionsSortBy(3), domain.ListSessionsSortByCreatedAt)
		assert.Equal(t, domain.ListSessionsSortBy(4), domain.ListSessionsSortByExpiresAt)
	})

	t.Run("should have unique values for each constant", func(t *testing.T) {
		constants := []domain.ListSessionsSortBy{
			domain.ListSessionsSortByNone,
			domain.ListSessionsSortByUserAgent,
			domain.ListSessionsSortByIPAddress,
			domain.ListSessionsSortByCreatedAt,
			domain.ListSessionsSortByExpiresAt,
		}

		for i := 0; i < len(constants); i++ {
			for j := i + 1; j < len(constants); j++ {
				assert.NotEqual(t, constants[i], constants[j])
			}
		}
	})
}
