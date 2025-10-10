package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrEmailAlreadyExists    = errors.New("email already in use")
	ErrInvalidEmailFormat    = errors.New("invalid email format")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters long")
	ErrPasswordMissingUpper  = errors.New("password must contain at least one uppercase letter")
	ErrPasswordMissingLower  = errors.New("password must contain at least one lowercase letter")
	ErrPasswordMissingDigit  = errors.New("password must contain at least one digit")
	ErrPasswordMissingSymbol = errors.New("password must contain at least one special character")
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash *string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}

type UserProfile struct {
	ID        int64
	UserID    int64
	Name      string
	AvatarURL string
	Phone     string
	BirthDate *time.Time
}

type UserProvider struct {
	ID           int64
	UserID       int64
	Provider     string
	ProviderID   string
	AccessToken  string
	RefreshToken string
	CreatedAt    time.Time
}
