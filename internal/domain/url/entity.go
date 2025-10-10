package url

import (
	"errors"
	"time"
)

var (
	ErrInvalidURLFormat     = errors.New("invalid url")
	ErrMissingURLSchema     = errors.New("url is missing scheme (e.g., http or https)")
	ErrUnsupportedURLSchema = errors.New("unsupported scheme")
	ErrMissingURLHost       = errors.New("url is missing host")
	ErrExpiredURL           = errors.New("expired URL")
	ErrURLNotFound          = errors.New("URL not found")
)

type URL struct {
	ShortCode    string
	EncryptedURL string
	ExpiresAt    *time.Time
}
