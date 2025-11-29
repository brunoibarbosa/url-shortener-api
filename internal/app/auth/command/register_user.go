package command

import (
	"context"
	"time"

	bd_domain "github.com/brunoibarbosa/url-shortener/internal/domain/bd"
	user_domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
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
	Profile   user_domain.UserProfile
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type RegisterUserHandler struct {
	tx                bd_domain.TransactionManager
	userRepo          user_domain.UserRepository
	providerRepo      user_domain.UserProviderRepository
	profileRepo       user_domain.UserProfileRepository
	passwordEncrypter user_domain.UserPasswordEncrypter
}

func NewRegisterUserHandler(
	tx bd_domain.TransactionManager,
	userRepo user_domain.UserRepository,
	providerRepo user_domain.UserProviderRepository,
	profileRepo user_domain.UserProfileRepository,
	passwordEncrypter user_domain.UserPasswordEncrypter,
) *RegisterUserHandler {
	return &RegisterUserHandler{
		tx,
		userRepo,
		providerRepo,
		profileRepo,
		passwordEncrypter,
	}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd RegisterUserCommand) (*RegisterUserResponse, error) {
	hash, err := h.passwordEncrypter.HashPassword(cmd.Password)
	if err != nil {
		return nil, err
	}

	exists, err := h.userRepo.Exists(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, user_domain.ErrEmailAlreadyExists
	}

	var u *user_domain.User
	var pv *user_domain.UserProvider
	var pf *user_domain.UserProfile

	err = h.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		u = &user_domain.User{
			Email: cmd.Email,
		}
		if err := h.userRepo.Create(txCtx, u); err != nil {
			return err
		}

		pv = &user_domain.UserProvider{
			Provider:     user_domain.ProviderPassword,
			ProviderID:   cmd.Email,
			PasswordHash: &hash,
		}
		if err := h.providerRepo.Create(txCtx, u.ID, pv); err != nil {
			return err
		}

		pf = &user_domain.UserProfile{
			Name: cmd.Name,
		}
		if err := h.profileRepo.Create(txCtx, u.ID, pf); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &RegisterUserResponse{
		ID:    u.ID,
		Email: u.Email,
		Profile: user_domain.UserProfile{
			Name:      pf.Name,
			AvatarURL: pf.AvatarURL,
		},
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}
