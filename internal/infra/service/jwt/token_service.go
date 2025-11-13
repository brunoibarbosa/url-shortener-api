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

func (s *TokenService) ParseAndValidate(tokenStr string) (*session.TokenClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return &session.TokenClaims{
		Sub: claims["sub"].(string),
		Sid: claims["sid"].(string),
		Exp: claims["exp"].(int64),
		Iat: claims["iat"].(int64),
	}, nil
}

func (s *TokenService) GenerateRefreshToken() uuid.UUID {
	tokenID := uuid.New()
	return tokenID
}
