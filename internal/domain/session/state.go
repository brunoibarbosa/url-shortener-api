package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

var (
	ErrInvalidState    = errors.New("invalid or expired state")
	ErrStateGeneration = errors.New("failed to generate state")
)

type StateService interface {
	GenerateState(ctx context.Context) (string, error)
	ValidateState(ctx context.Context, state string) error
	DeleteState(ctx context.Context, state string) error
}

func GenerateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", ErrStateGeneration
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
