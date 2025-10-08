package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmailFormat = errors.New("invalid email format")
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
