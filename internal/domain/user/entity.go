package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSocialLoginOnly       = errors.New("user has no password (social login only)")
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrEmailAlreadyExists    = errors.New("email already in use")
	ErrInvalidEmailFormat    = errors.New("invalid email format")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters long")
	ErrPasswordMissingUpper  = errors.New("password must contain at least one uppercase letter")
	ErrPasswordMissingLower  = errors.New("password must contain at least one lowercase letter")
	ErrPasswordMissingDigit  = errors.New("password must contain at least one digit")
	ErrPasswordMissingSymbol = errors.New("password must contain at least one special character")
	ErrCreatingUser          = errors.New("error creating user")
)

type User struct {
	ID        uuid.UUID
	Email     string
	Profile   *UserProfile
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type UserProfile struct {
	ID        int64
	Name      string
	AvatarURL *string
}

type UserProvider struct {
	ID           int64
	UserID       uuid.UUID
	Provider     string
	ProviderID   string
	PasswordHash *string
}

type UserSession struct {
	ID           int64
	UserID       uuid.UUID
	RefreshToken uuid.UUID
	UserAgent    *string
	IPAddress    *string
	ExpiresAt    *time.Time
	RevokedAt    *time.Time
	CreatedAt    time.Time
}

func (s *UserSession) IsExpired() bool {
	return time.Now().After(*s.ExpiresAt)
}
