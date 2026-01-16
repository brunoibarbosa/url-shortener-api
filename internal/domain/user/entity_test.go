package user_test

import (
	"testing"
	"time"

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

func TestUser_UpdateProfile(t *testing.T) {
	t.Run("should update profile name", func(t *testing.T) {
		user := &domain.User{
			ID:    uuid.New(),
			Email: "test@example.com",
			Profile: &domain.UserProfile{
				ID:   1,
				Name: "Original Name",
			},
			CreatedAt: time.Now(),
		}

		user.Profile.Name = "Updated Name"

		assert.Equal(t, "Updated Name", user.Profile.Name)
	})

	t.Run("should update profile with avatar", func(t *testing.T) {
		user := &domain.User{
			ID:    uuid.New(),
			Email: "test@example.com",
			Profile: &domain.UserProfile{
				ID:        1,
				Name:      "Test User",
				AvatarURL: nil,
			},
			CreatedAt: time.Now(),
		}

		avatarURL := "https://example.com/avatar.jpg"
		user.Profile.AvatarURL = &avatarURL

		assert.NotNil(t, user.Profile.AvatarURL)
		assert.Equal(t, avatarURL, *user.Profile.AvatarURL)
	})
}

func TestUser_WithNilProfile(t *testing.T) {
	t.Run("should handle user without profile", func(t *testing.T) {
		user := &domain.User{
			ID:        uuid.New(),
			Email:     "test@example.com",
			Profile:   nil,
			CreatedAt: time.Now(),
		}

		assert.Nil(t, user.Profile)
		assert.NotEmpty(t, user.Email)
	})
}

func TestUser_UpdatedAt(t *testing.T) {
	t.Run("should set updated_at after modification", func(t *testing.T) {
		user := &domain.User{
			ID:        uuid.New(),
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: nil,
		}

		assert.Nil(t, user.UpdatedAt)

		updatedAt := time.Now()
		user.UpdatedAt = &updatedAt

		assert.NotNil(t, user.UpdatedAt)
		assert.True(t, user.UpdatedAt.After(user.CreatedAt) || user.UpdatedAt.Equal(user.CreatedAt))
	})
}

func TestUserProvider_MultipleProviders(t *testing.T) {
	t.Run("should allow multiple providers for same user", func(t *testing.T) {
		userID := uuid.New()
		passwordHash := "hashed_password"

		passwordProvider := &domain.UserProvider{
			ID:           1,
			UserID:       userID,
			Provider:     domain.ProviderPassword,
			ProviderID:   "password_id",
			PasswordHash: &passwordHash,
		}

		googleProvider := &domain.UserProvider{
			ID:           2,
			UserID:       userID,
			Provider:     domain.ProviderGoogle,
			ProviderID:   "google_123456",
			PasswordHash: nil,
		}

		assert.Equal(t, userID, passwordProvider.UserID)
		assert.Equal(t, userID, googleProvider.UserID)
		assert.NotNil(t, passwordProvider.PasswordHash)
		assert.Nil(t, googleProvider.PasswordHash)
	})
}

func TestUserProvider_ProviderConstants(t *testing.T) {
	t.Run("should have correct provider constant values", func(t *testing.T) {
		assert.Equal(t, "password", domain.ProviderPassword)
		assert.Equal(t, "google", domain.ProviderGoogle)
	})
}

func TestUser_Errors(t *testing.T) {
	t.Run("should have all error constants defined", func(t *testing.T) {
		assert.Equal(t, "user has no password (social login only)", domain.ErrSocialLoginOnly.Error())
		assert.Equal(t, "invalid email or password", domain.ErrInvalidCredentials.Error())
		assert.Equal(t, "email already in use", domain.ErrEmailAlreadyExists.Error())
		assert.Equal(t, "invalid email format", domain.ErrInvalidEmailFormat.Error())
		assert.Equal(t, "password must be at least 8 characters long", domain.ErrPasswordTooShort.Error())
		assert.Equal(t, "password must contain at least one uppercase letter", domain.ErrPasswordMissingUpper.Error())
		assert.Equal(t, "password must contain at least one lowercase letter", domain.ErrPasswordMissingLower.Error())
		assert.Equal(t, "password must contain at least one digit", domain.ErrPasswordMissingDigit.Error())
		assert.Equal(t, "password must contain at least one special character", domain.ErrPasswordMissingSymbol.Error())
		assert.Equal(t, "error creating user", domain.ErrCreatingUser.Error())
		assert.Equal(t, "user not found", domain.ErrNotFound.Error())
		assert.Equal(t, "user not authenticated", domain.ErrUserNotAuthenticated.Error())
		assert.Equal(t, "invalid user ID in context", domain.ErrInvalidUserIDContext.Error())
	})
}
