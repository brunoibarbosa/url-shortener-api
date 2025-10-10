package command

import (
	"context"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserCommand struct {
	Email    string
	Password string
}

type RegisterUserHandler struct {
	repo             domain.UserRepository
	encryptSecretKey string
}

func NewRegisteUserHandler(repo domain.UserRepository, secretKey string) *RegisterUserHandler {
	return &RegisterUserHandler{
		repo:             repo,
		encryptSecretKey: secretKey,
	}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	exists, err := h.repo.Exists(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, domain.ErrEmailAlreadyExists
	}

	return h.repo.CreateUser(ctx, cmd.Email, string(hash))
}
