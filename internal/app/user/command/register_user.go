package command

import (
	"context"

	"github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserCommand struct {
	Email    string
	Password string
}

type RegisterUserHandler struct {
	repo             user.UserRepository
	encryptSecretKey string
}

func NewRegisteUserHandler(repo user.UserRepository, secretKey string) *RegisterUserHandler {
	return &RegisterUserHandler{
		repo:             repo,
		encryptSecretKey: secretKey,
	}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (*user.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return h.repo.CreateUser(ctx, cmd.Email, string(hash))
}
