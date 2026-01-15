package url

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidURLFormat     = errors.New("invalid url")
	ErrMissingURLSchema     = errors.New("url is missing scheme (e.g., http or https)")
	ErrUnsupportedURLSchema = errors.New("unsupported scheme")
	ErrMissingURLHost       = errors.New("url is missing host")
	ErrExpiredURL           = errors.New("expired URL")
	ErrDeletedURL           = errors.New("deleted URL")
	ErrURLNotFound          = errors.New("URL not found")
	ErrInvalidShortCode     = errors.New("invalid short code")
)

type URL struct {
	ShortCode    string
	EncryptedURL string
	UserID       *uuid.UUID
	ExpiresAt    *time.Time
	DeletedAt    *time.Time
}

func (u *URL) RemainingTTL(now time.Time) time.Duration {
	if u.ExpiresAt == nil {
		return 0
	}
	return u.ExpiresAt.Sub(now)
}

func (u *URL) IsExpired(now time.Time) bool {
	if u.ExpiresAt == nil {
		return false
	}
	return now.After(*u.ExpiresAt)
}

func (u *URL) IsDeleted() bool {
	return u.DeletedAt != nil
}

func (u *URL) CanBeAccessed(now time.Time) error {
	if u.IsDeleted() {
		return ErrDeletedURL
	}
	if u.IsExpired(now) {
		return ErrExpiredURL
	}
	return nil
}
