package user_test

import (
	"testing"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_Creation(t *testing.T) {
	t.Run("should create user with valid data", func(t *testing.T) {
		id := uuid.New()
		email := "user@example.com"

		user := &domain.User{
			ID:    id,
			Email: email,
		}

		assert.Equal(t, id, user.ID)
		assert.Equal(t, email, user.Email)
		assert.Nil(t, user.Profile)
		assert.Nil(t, user.UpdatedAt)
	})

	t.Run("should create user with profile", func(t *testing.T) {
		id := uuid.New()
		email := "user@example.com"
		name := "John Doe"

		profile := &domain.UserProfile{
			ID:   1,
			Name: name,
		}

		user := &domain.User{
			ID:      id,
			Email:   email,
			Profile: profile,
		}

		assert.Equal(t, id, user.ID)
		assert.Equal(t, email, user.Email)
		assert.NotNil(t, user.Profile)
		assert.Equal(t, name, user.Profile.Name)
	})
}

func TestUserProvider_IsPasswordProvider(t *testing.T) {
	t.Run("should return true for password provider", func(t *testing.T) {
		provider := &domain.UserProvider{
			Provider: domain.ProviderPassword,
		}

		assert.Equal(t, domain.ProviderPassword, provider.Provider)
	})

	t.Run("should return false for non-password provider", func(t *testing.T) {
		provider := &domain.UserProvider{
			Provider: domain.ProviderGoogle,
		}

		assert.NotEqual(t, domain.ProviderPassword, provider.Provider)
		assert.Equal(t, domain.ProviderGoogle, provider.Provider)
	})
}

func TestUserProvider_HasPassword(t *testing.T) {
	t.Run("should return true when PasswordHash is set", func(t *testing.T) {
		hash := "$2a$10$hashedpassword"
		provider := &domain.UserProvider{
			Provider:     domain.ProviderPassword,
			PasswordHash: &hash,
		}

		assert.NotNil(t, provider.PasswordHash)
		assert.Equal(t, hash, *provider.PasswordHash)
	})

	t.Run("should return false when PasswordHash is nil", func(t *testing.T) {
		provider := &domain.UserProvider{
			Provider:     domain.ProviderGoogle,
			PasswordHash: nil,
		}

		assert.Nil(t, provider.PasswordHash)
	})
}

func TestUserProfile_WithAvatarURL(t *testing.T) {
	t.Run("should create profile without avatar", func(t *testing.T) {
		profile := &domain.UserProfile{
			ID:        1,
			Name:      "John Doe",
			AvatarURL: nil,
		}

		assert.Equal(t, "John Doe", profile.Name)
		assert.Nil(t, profile.AvatarURL)
	})

	t.Run("should create profile with avatar", func(t *testing.T) {
		avatarURL := "https://example.com/avatar.jpg"
		profile := &domain.UserProfile{
			ID:        1,
			Name:      "John Doe",
			AvatarURL: &avatarURL,
		}

		assert.Equal(t, "John Doe", profile.Name)
		assert.NotNil(t, profile.AvatarURL)
		assert.Equal(t, avatarURL, *profile.AvatarURL)
	})
}

func TestProviderConstants(t *testing.T) {
	t.Run("should have correct provider constants", func(t *testing.T) {
		assert.Equal(t, "password", domain.ProviderPassword)
		assert.Equal(t, "google", domain.ProviderGoogle)
	})
}
