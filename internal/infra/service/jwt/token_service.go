package jwt

import (
	"errors"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrTokenGeneration = errors.New("token generation failed")
)

type TokenService struct {
	secret string
}

func NewTokenService(secret string) *TokenService {
	return &TokenService{secret: secret}
}

func (s *TokenService) GenerateAccessToken(params *session.TokenParams) (string, error) {
	claims := jwt.MapClaims{
		"sub": params.UserID.String(),
		"sid": params.SessionID.String(),
		"exp": time.Now().Add(params.Duration).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", session.ErrTokenGenerate
	}

	return signed, nil
}

func (s *TokenService) GenerateRefreshToken() uuid.UUID {
	tokenID := uuid.New()
	return tokenID
}
