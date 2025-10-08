package command

import (
	"context"
	"errors"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginUserCommand struct {
	Email    string
	Password string
}

type LoginUserHandler struct {
	repo             user.UserRepository
	encryptSecretKey string
}

func NewLoginUserHandler(repo user.UserRepository, secretKey string) *LoginUserHandler {
	return &LoginUserHandler{
		repo:             repo,
		encryptSecretKey: secretKey,
	}
}

func (h *LoginUserHandler) Handle(ctx context.Context, cmd LoginUserCommand) (string, error) {
	u, err := h.repo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return "", err
	}

	if u.PasswordHash == nil {
		return "", errors.New("user has no password (social login only)")
	}

	if bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(cmd.Password)) != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": u.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(h.encryptSecretKey)
}
