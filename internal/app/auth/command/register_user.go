package command

import (
	"context"
	"time"

	domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
	"github.com/google/uuid"
)

type RegisterUserCommand struct {
	Email    string
	Password string
	Name     string
}

type RegisterUserResponse struct {
	ID        uuid.UUID
	Email     string
	Profile   domain.UserProfile
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type RegisterUserHandler struct {
	userRepo     domain.UserRepository
	providerRepo domain.UserProviderRepository
	profileRepo  domain.UserProfileRepository
}

func NewRegisterUserHandler(userRepo domain.UserRepository, providerRepo domain.UserProviderRepository, profileRepo domain.UserProfileRepository) *RegisterUserHandler {
	return &RegisterUserHandler{
		userRepo,
		providerRepo,
		profileRepo,
	}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (*RegisterUserResponse, error) {
	hash, err := crypto.HashPassword(cmd.Password)
	if err != nil {
		return nil, err
	}

	exists, err := h.userRepo.Exists(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, domain.ErrEmailAlreadyExists
	}

	// Cria o usuário
	u := &domain.User{
		Email: cmd.Email,
	}
	err = h.userRepo.Create(ctx, u)
	if err != nil {
		return nil, err
	}

	// Cria o provider de senha
	pv := &domain.UserProvider{
		Provider:     "password",
		ProviderID:   cmd.Email,
		PasswordHash: &hash,
	}
	err = h.providerRepo.Create(ctx, u.ID, pv)
	if err != nil {
		return nil, err
	}

	// Cria o perfil do usuário
	pf := &domain.UserProfile{
		Name: cmd.Name,
	}
	err = h.profileRepo.Create(ctx, u.ID, pf)
	if err != nil {
		return nil, err
	}

	return &RegisterUserResponse{
		ID:    u.ID,
		Email: u.Email,
		Profile: domain.UserProfile{
			Name:      pf.Name,
			AvatarURL: pf.AvatarURL,
		},
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}
