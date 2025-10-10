package command

import (
	"context"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginUserCommand struct {
	Email    string
	Password string
}

type LoginUserHandler struct {
	repo             domain.UserRepository
	encryptSecretKey string
}

type LoginUserResult struct {
	Token     string
	ExpiresAt time.Time
	UserID    uuid.UUID
	Email     string
}

func NewLoginUserHandler(repo domain.UserRepository, secretKey string) *LoginUserHandler {
	return &LoginUserHandler{
		repo:             repo,
		encryptSecretKey: secretKey,
	}
}

func (h *LoginUserHandler) Handle(ctx context.Context, cmd LoginUserCommand) (*LoginUserResult, error) {
	u, err := h.repo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}

	if u.PasswordHash == nil {
		return nil, domain.ErrSocialLoginOnly
	}

	if bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(cmd.Password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	sToken, err := token.SignedString([]byte(h.encryptSecretKey))

	if err != nil {
		return nil, err
	}

	expiration := time.Now().Add(24 * time.Hour)

	return &LoginUserResult{
		Token:     sToken,
		ExpiresAt: expiration,
		UserID:    u.ID,
		Email:     u.Email,
	}, nil
}
