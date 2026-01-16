package testhelpers

import (
	"context"
	"time"

	"testing"

	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	url_domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	user_domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// SessionFixture provides test data for session tests
type SessionFixture struct {
	UserID           uuid.UUID
	RefreshTokenHash string
	UserAgent        string
	IPAddress        string
	ExpiresAt        *time.Time
}

// DefaultSessionFixture returns a session fixture with default values
func DefaultSessionFixture(userID uuid.UUID) *SessionFixture {
	expiresAt := time.Now().Add(24 * time.Hour)
	return &SessionFixture{
		UserID:           userID,
		RefreshTokenHash: "default_hash_" + uuid.New().String(),
		UserAgent:        "Mozilla/5.0 Test Agent",
		IPAddress:        "127.0.0.1",
		ExpiresAt:        &expiresAt,
	}
}

// ToSession converts fixture to domain session entity
func (f *SessionFixture) ToSession() *session_domain.Session {
	return &session_domain.Session{
		UserID:           f.UserID,
		RefreshTokenHash: f.RefreshTokenHash,
		UserAgent:        f.UserAgent,
		IPAddress:        f.IPAddress,
		ExpiresAt:        f.ExpiresAt,
	}
}

// WithRefreshTokenHash sets a custom refresh token hash
func (f *SessionFixture) WithRefreshTokenHash(hash string) *SessionFixture {
	f.RefreshTokenHash = hash
	return f
}

// WithUserAgent sets a custom user agent
func (f *SessionFixture) WithUserAgent(userAgent string) *SessionFixture {
	f.UserAgent = userAgent
	return f
}

// WithIPAddress sets a custom IP address
func (f *SessionFixture) WithIPAddress(ip string) *SessionFixture {
	f.IPAddress = ip
	return f
}

// WithExpiresAt sets a custom expiration time
func (f *SessionFixture) WithExpiresAt(expiresAt time.Time) *SessionFixture {
	f.ExpiresAt = &expiresAt
	return f
}

// URLFixture provides test data for URL tests
type URLFixture struct {
	ShortCode    string
	EncryptedURL string
	UserID       *uuid.UUID
	ExpiresAt    *time.Time
	DeletedAt    *time.Time
}

// DefaultURLFixture returns a URL fixture with default values
func DefaultURLFixture(userID *uuid.UUID) *URLFixture {
	return &URLFixture{
		ShortCode:    "test" + uuid.New().String()[:8],
		EncryptedURL: "https://example.com/test",
		UserID:       userID,
		ExpiresAt:    nil,
		DeletedAt:    nil,
	}
}

// ToURL converts fixture to domain URL entity
func (f *URLFixture) ToURL() *url_domain.URL {
	return &url_domain.URL{
		ShortCode:    f.ShortCode,
		EncryptedURL: f.EncryptedURL,
		UserID:       f.UserID,
		ExpiresAt:    f.ExpiresAt,
		DeletedAt:    f.DeletedAt,
	}
}

// WithShortCode sets a custom short code
func (f *URLFixture) WithShortCode(code string) *URLFixture {
	f.ShortCode = code
	return f
}

// WithEncryptedURL sets a custom encrypted URL
func (f *URLFixture) WithEncryptedURL(url string) *URLFixture {
	f.EncryptedURL = url
	return f
}

// WithExpiresAt sets a custom expiration time
func (f *URLFixture) WithExpiresAt(expiresAt time.Time) *URLFixture {
	f.ExpiresAt = &expiresAt
	return f
}

// UserFixture provides test data for user tests
type UserFixture struct {
	Email string
}

// DefaultUserFixture returns a user fixture with default values
func DefaultUserFixture() *UserFixture {
	return &UserFixture{
		Email: "test-" + uuid.New().String()[:8] + "@example.com",
	}
}

// ToUser converts fixture to domain user entity
func (f *UserFixture) ToUser() *user_domain.User {
	return &user_domain.User{
		Email: f.Email,
	}
}

// WithEmail sets a custom email
func (f *UserFixture) WithEmail(email string) *UserFixture {
	f.Email = email
	return f
}

// CreateTestUser creates a user in the database for testing
func CreateTestUser(t *testing.T, ctx context.Context, db *pgxpool.Pool) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	_, err := db.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)",
		userID, "testuser-"+userID.String()+"@example.com")
	require.NoError(t, err)
	return userID
}
