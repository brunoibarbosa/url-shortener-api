package command

import (
	"context"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
)

type LoginUserCommand struct {
	Email    string
	Password string
}

type LoginUserHandler struct {
	providerRepo  domain.UserProviderRepository
	tokenService  domain.TokenService
	passwordCheck func(hash, plain string) bool
}

func NewLoginUserHandler(
	providerRepo domain.UserProviderRepository,
	tokenService domain.TokenService,
	passwordCheck func(hash, plain string) bool,
) *LoginUserHandler {
	return &LoginUserHandler{
		providerRepo,
		tokenService,
		passwordCheck,
	}
}

func (h *LoginUserHandler) Handle(ctx context.Context, cmd LoginUserCommand) (string, error) {
	u, err := h.providerRepo.Find(ctx, "password", cmd.Email)
	if err != nil {
		return "", err
	}

	if u == nil {
		return "", domain.ErrInvalidCredentials
	}

	if !h.passwordCheck(cmd.Password, *u.PasswordHash) {
		return "", domain.ErrInvalidCredentials
	}

	token, err := h.tokenService.GenerateAccessToken(&domain.TokenParams{
		UserID: u.UserID,
	})
	if err != nil {
		return "", err
	}

	return token, nil
}
