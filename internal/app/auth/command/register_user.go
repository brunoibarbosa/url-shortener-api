package command

import (
	"context"
	"time"

	user_domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/infra/database/pg"
	"github.com/brunoibarbosa/url-shortener/pkg/crypto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	db           *pg.Postgres
	userRepo     user_domain.UserRepository
	providerRepo user_domain.UserProviderRepository
	profileRepo  user_domain.UserProfileRepository
}

func NewRegisterUserHandler(db *pg.Postgres, userRepo user_domain.UserRepository, providerRepo user_domain.UserProviderRepository, profileRepo user_domain.UserProfileRepository) *RegisterUserHandler {
	return &RegisterUserHandler{
		db,
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
		return nil, user_domain.ErrEmailAlreadyExists
	}

	var u *user_domain.User
	var pv *user_domain.UserProvider
	var pf *user_domain.UserProfile

	err = h.db.WithTransaction(ctx, func(tx pgx.Tx) error {
		userRepo := h.userRepo.WithTx(tx)
		providerRepo := h.providerRepo.WithTx(tx)
		profileRepo := h.profileRepo.WithTx(tx)

		u = &user_domain.User{
			Email: cmd.Email,
		}
		if err := userRepo.Create(ctx, u); err != nil {
			return err
		}

		pv = &user_domain.UserProvider{
			Provider:     "password",
			ProviderID:   cmd.Email,
			PasswordHash: &hash,
		}
		if err := providerRepo.Create(ctx, u.ID, pv); err != nil {
			return err
		}

		pf = &user_domain.UserProfile{
			Name: cmd.Name,
		}
		if err := profileRepo.Create(ctx, u.ID, pf); err != nil {
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
