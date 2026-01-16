package jwt_test

import (
	"testing"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/internal/infra/service/jwt"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenService(t *testing.T) {
	secret := "test-secret-key"
	service := jwt.NewTokenService(secret)

	assert.NotNil(t, service)
}

func TestTokenService_GenerateAccessToken_Success(t *testing.T) {
	secret := "test-secret-key-with-32-bytes!!"
	service := jwt.NewTokenService(secret)

	userID := uuid.New()
	sessionID := uuid.New()
	duration := 15 * time.Minute

	params := &session.TokenParams{
		UserID:    userID,
		SessionID: sessionID,
		Duration:  duration,
	}

	token, err := service.GenerateAccessToken(params)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token structure
	parsedToken, err := jwtlib.Parse(token, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Validate claims
	claims, ok := parsedToken.Claims.(jwtlib.MapClaims)
	require.True(t, ok)

	sub, ok := claims["sub"].(string)
	require.True(t, ok)
	assert.Equal(t, userID.String(), sub)

	sid, ok := claims["sid"].(string)
	require.True(t, ok)
	assert.Equal(t, sessionID.String(), sid)

	// Validate expiration
	exp, ok := claims["exp"].(float64)
	require.True(t, ok)
	assert.Greater(t, exp, float64(time.Now().Unix()))

	// Validate issued at
	iat, ok := claims["iat"].(float64)
	require.True(t, ok)
	assert.LessOrEqual(t, iat, float64(time.Now().Unix()))
}

func TestTokenService_GenerateAccessToken_WithDifferentDurations(t *testing.T) {
	secret := "test-secret-key-with-32-bytes!!"
	service := jwt.NewTokenService(secret)

	userID := uuid.New()
	sessionID := uuid.New()

	testCases := []struct {
		name     string
		duration time.Duration
	}{
		{"1 minute", 1 * time.Minute},
		{"15 minutes", 15 * time.Minute},
		{"1 hour", 1 * time.Hour},
		{"24 hours", 24 * time.Hour},
		{"7 days", 7 * 24 * time.Hour},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := &session.TokenParams{
				UserID:    userID,
				SessionID: sessionID,
				Duration:  tc.duration,
			}

			token, err := service.GenerateAccessToken(params)

			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// Parse and validate expiration
			parsedToken, err := jwtlib.Parse(token, func(token *jwtlib.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			require.NoError(t, err)
			claims := parsedToken.Claims.(jwtlib.MapClaims)
			exp := int64(claims["exp"].(float64))
			expectedExp := time.Now().Add(tc.duration).Unix()

			// Allow 2 seconds tolerance for test execution time
			assert.InDelta(t, expectedExp, exp, 2)
		})
	}
}

func TestTokenService_GenerateAccessToken_TokenCanBeVerified(t *testing.T) {
	secret := "test-secret-key"
	service := jwt.NewTokenService(secret)

	userID := uuid.New()
	sessionID := uuid.New()

	params := &session.TokenParams{
		UserID:    userID,
		SessionID: sessionID,
		Duration:  15 * time.Minute,
	}

	token, err := service.GenerateAccessToken(params)
	require.NoError(t, err)

	// Verify with correct secret
	parsedToken, err := jwtlib.Parse(token, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
}

func TestTokenService_GenerateAccessToken_TokenCannotBeVerifiedWithWrongSecret(t *testing.T) {
	secret := "correct-secret"
	wrongSecret := "wrong-secret"
	service := jwt.NewTokenService(secret)

	userID := uuid.New()
	sessionID := uuid.New()

	params := &session.TokenParams{
		UserID:    userID,
		SessionID: sessionID,
		Duration:  15 * time.Minute,
	}

	token, err := service.GenerateAccessToken(params)
	require.NoError(t, err)

	// Try to verify with wrong secret
	parsedToken, err := jwtlib.Parse(token, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(wrongSecret), nil
	})

	assert.Error(t, err)
	assert.False(t, parsedToken.Valid)
}

func TestTokenService_GenerateAccessToken_ExpiredToken(t *testing.T) {
	secret := "test-secret-key"
	service := jwt.NewTokenService(secret)

	userID := uuid.New()
	sessionID := uuid.New()

	// Generate token with negative duration (already expired)
	params := &session.TokenParams{
		UserID:    userID,
		SessionID: sessionID,
		Duration:  -1 * time.Hour,
	}

	token, err := service.GenerateAccessToken(params)
	require.NoError(t, err)

	// Try to parse expired token
	parsedToken, err := jwtlib.Parse(token, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	// Token should be invalid due to expiration
	assert.Error(t, err)
	assert.False(t, parsedToken.Valid)
}

func TestTokenService_GenerateAccessToken_MultipleTokensAreDifferent(t *testing.T) {
	secret := "test-secret-key"
	service := jwt.NewTokenService(secret)

	userID1 := uuid.New()
	userID2 := uuid.New()
	sessionID := uuid.New()

	params1 := &session.TokenParams{
		UserID:    userID1,
		SessionID: sessionID,
		Duration:  15 * time.Minute,
	}

	params2 := &session.TokenParams{
		UserID:    userID2,
		SessionID: sessionID,
		Duration:  15 * time.Minute,
	}

	token1, err := service.GenerateAccessToken(params1)
	require.NoError(t, err)

	token2, err := service.GenerateAccessToken(params2)
	require.NoError(t, err)

	assert.NotEqual(t, token1, token2)
}

func TestTokenService_GenerateRefreshToken_Success(t *testing.T) {
	secret := "test-secret-key"
	service := jwt.NewTokenService(secret)

	token := service.GenerateRefreshToken()

	assert.NotEqual(t, uuid.Nil, token)
	assert.NotEmpty(t, token.String())
}

func TestTokenService_GenerateRefreshToken_EachTokenIsUnique(t *testing.T) {
	secret := "test-secret-key"
	service := jwt.NewTokenService(secret)

	tokens := make(map[uuid.UUID]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		token := service.GenerateRefreshToken()
		assert.NotEqual(t, uuid.Nil, token)
		assert.False(t, tokens[token], "Generated duplicate refresh token")
		tokens[token] = true
	}

	assert.Len(t, tokens, iterations)
}

func TestTokenService_GenerateRefreshToken_ValidUUIDFormat(t *testing.T) {
	secret := "test-secret-key"
	service := jwt.NewTokenService(secret)

	token := service.GenerateRefreshToken()

	// UUID should be parseable
	_, err := uuid.Parse(token.String())
	assert.NoError(t, err)

	// UUID should have correct format (36 characters with dashes)
	assert.Len(t, token.String(), 36)
}

func TestTokenService_GenerateAccessToken_WithZeroDuration(t *testing.T) {
	secret := "test-secret-key"
	service := jwt.NewTokenService(secret)

	userID := uuid.New()
	sessionID := uuid.New()

	params := &session.TokenParams{
		UserID:    userID,
		SessionID: sessionID,
		Duration:  0,
	}

	token, err := service.GenerateAccessToken(params)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token should be immediately expired
	parsedToken, err := jwtlib.Parse(token, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	// Token is technically valid but expired
	assert.Error(t, err)
	assert.False(t, parsedToken.Valid)
}

func TestTokenService_GenerateAccessToken_ContainsAllRequiredClaims(t *testing.T) {
	secret := "test-secret-key"
	service := jwt.NewTokenService(secret)

	userID := uuid.New()
	sessionID := uuid.New()

	params := &session.TokenParams{
		UserID:    userID,
		SessionID: sessionID,
		Duration:  15 * time.Minute,
	}

	token, err := service.GenerateAccessToken(params)
	require.NoError(t, err)

	parsedToken, err := jwtlib.Parse(token, func(token *jwtlib.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)

	claims := parsedToken.Claims.(jwtlib.MapClaims)

	// Verify all required claims exist
	assert.Contains(t, claims, "sub")
	assert.Contains(t, claims, "sid")
	assert.Contains(t, claims, "exp")
	assert.Contains(t, claims, "iat")
}
