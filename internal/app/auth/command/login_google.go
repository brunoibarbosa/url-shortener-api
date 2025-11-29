package command

import (
	"context"
	"errors"
	"time"

	bd_domain "github.com/brunoibarbosa/url-shortener/internal/domain/bd"
	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	user_domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
)

type LoginGoogleCommand struct {
	Code      string
	State     string
	UserAgent string
	IPAddress string
}

type LoginGoogleHandler struct {
	txManager            bd_domain.TransactionManager
	provider             session_domain.OAuthProvider
	userRepo             user_domain.UserRepository
	providerRepo         user_domain.UserProviderRepository
	profileRepo          user_domain.UserProfileRepository
	sessionRepo          session_domain.SessionRepository
	tokenService         session_domain.TokenService
	sessionEncrypter     session_domain.SessionEncrypter
	stateService         session_domain.StateService
	refreshTokenDuration time.Duration
	accessTokenDuration  time.Duration
}

func NewLoginGoogleHandler(
	txManager bd_domain.TransactionManager,
	provider session_domain.OAuthProvider,
	userRepo user_domain.UserRepository,
	providerRepo user_domain.UserProviderRepository,
	profileRepo user_domain.UserProfileRepository,
	sessionRepo session_domain.SessionRepository,
	tokenService session_domain.TokenService,
	sessionEncrypter session_domain.SessionEncrypter,
	stateService session_domain.StateService,
	refreshTokenDuration time.Duration,
	accessTokenDuration time.Duration,
) *LoginGoogleHandler {
	return &LoginGoogleHandler{
		txManager,
		provider,
		userRepo,
		providerRepo,
		profileRepo,
		sessionRepo,
		tokenService,
		sessionEncrypter,
		stateService,
		refreshTokenDuration,
		accessTokenDuration,
	}
}

func (h *LoginGoogleHandler) Handle(ctx context.Context, cmd LoginGoogleCommand) (string, string, error) {
	if cmd.Code == "" {
		return "", "", session_domain.ErrInvalidOAuthCode
	}

	if err := h.stateService.ValidateState(ctx, cmd.State); err != nil {
		return "", "", err
	}

	defer h.stateService.DeleteState(ctx, cmd.State)

	oauthUser, err := h.provider.ExchangeCode(ctx, cmd.Code)
	if err != nil {
		return "", "", err
	}

	var session *session_domain.Session
	var refreshToken string

	err = h.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		existingProvider, err := h.providerRepo.Find(txCtx, user_domain.ProviderGoogle, oauthUser.ID)
		if err != nil {
			switch {
			case errors.Is(err, user_domain.ErrNotFound):
				existingProvider = nil
			default:
				return err
			}
		}

		var user *user_domain.User

		if existingProvider != nil {
			user, err = h.userRepo.GetByID(txCtx, existingProvider.UserID)
			if err != nil {
				return err
			}
		} else {
			user, err = h.userRepo.GetByEmail(txCtx, oauthUser.Email)
			if err != nil && err != user_domain.ErrNotFound {
				return err
			}
			if user == nil {
				user = &user_domain.User{
					Email: oauthUser.Email,
				}
				if err := h.userRepo.Create(txCtx, user); err != nil {
					return user_domain.ErrCreatingUser
				}
			}

			providerEntry := &user_domain.UserProvider{
				UserID:     user.ID,
				Provider:   user_domain.ProviderGoogle,
				ProviderID: oauthUser.ID,
			}
			if err := h.providerRepo.Create(txCtx, user.ID, providerEntry); err != nil {
				return err
			}

			if oauthUser.Name != "" {
				profile := &user_domain.UserProfile{
					Name:      oauthUser.Name,
					AvatarURL: oauthUser.AvatarURL,
				}
				if err := h.profileRepo.Create(txCtx, user.ID, profile); err != nil {
					return err
				}
			}
		}

		refreshTokenObj := h.tokenService.GenerateRefreshToken()
		refreshToken = refreshTokenObj.String()
		refreshHash := h.sessionEncrypter.HashRefreshToken(refreshToken)

		expiresAt := time.Now().Add(h.refreshTokenDuration)

		session = &session_domain.Session{
			UserID:           user.ID,
			UserAgent:        cmd.UserAgent,
			IPAddress:        cmd.IPAddress,
			RefreshTokenHash: refreshHash,
			ExpiresAt:        &expiresAt,
		}

		return h.sessionRepo.Create(txCtx, session)
	})

	if err != nil {
		return "", "", err
	}

	accessToken, err := h.tokenService.GenerateAccessToken(&session_domain.TokenParams{
		UserID:    session.UserID,
		SessionID: session.ID,
		Duration:  h.accessTokenDuration,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
