package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	user_domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestLoginGoogleHandler_Handle_Success_NewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTxManager := mocks.NewMockTransactionManager(ctrl)
	mockProvider := mocks.NewMockOAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)
	mockStateService := mocks.NewMockStateService(ctrl)

	cmd := command.LoginGoogleCommand{
		Code:      "valid_code",
		State:     "valid_state",
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}

	avatarURL := "https://example.com/avatar.jpg"
	oauthUser := &session_domain.OAuthUser{
		ID:        "google_123",
		Email:     "user@example.com",
		Name:      "Test User",
		AvatarURL: &avatarURL,
	}

	mockStateService.EXPECT().ValidateState(ctx, "valid_state").Return(nil)
	mockStateService.EXPECT().DeleteState(ctx, "valid_state").Return(nil)
	mockProvider.EXPECT().ExchangeCode(ctx, "valid_code").Return(oauthUser, nil)

	mockTxManager.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			mockProviderRepo.EXPECT().Find(ctx, user_domain.ProviderGoogle, "google_123").Return(nil, user_domain.ErrNotFound)
			mockUserRepo.EXPECT().GetByEmail(ctx, "user@example.com").Return(nil, user_domain.ErrNotFound)
			mockUserRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
			mockProviderRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(nil)
			mockProfileRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(nil)
			mockSessionRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
			return fn(ctx)
		},
	)

	refreshTokenUUID := uuid.New()
	mockTokenService.EXPECT().GenerateRefreshToken().Return(refreshTokenUUID)
	mockSessionEncrypter.EXPECT().HashRefreshToken(refreshTokenUUID.String()).Return("hashed_refresh")
	mockTokenService.EXPECT().GenerateAccessToken(gomock.Any()).Return("access_token", nil)

	handler := command.NewLoginGoogleHandler(
		mockTxManager,
		mockProvider,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockSessionRepo,
		mockTokenService,
		mockSessionEncrypter,
		mockStateService,
		24*time.Hour,
		15*time.Minute,
	)

	accessToken, refreshToken, err := handler.Handle(ctx, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "access_token", accessToken)
	assert.Equal(t, refreshTokenUUID.String(), refreshToken)
}

func TestLoginGoogleHandler_Handle_Success_ExistingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTxManager := mocks.NewMockTransactionManager(ctrl)
	mockProvider := mocks.NewMockOAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)
	mockStateService := mocks.NewMockStateService(ctrl)

	cmd := command.LoginGoogleCommand{
		Code:      "valid_code",
		State:     "valid_state",
		UserAgent: "Chrome",
		IPAddress: "10.0.0.1",
	}

	userID := uuid.New()
	existingProvider := &user_domain.UserProvider{
		ID:         1,
		UserID:     userID,
		Provider:   user_domain.ProviderGoogle,
		ProviderID: "google_456",
	}

	existingUser := &user_domain.User{
		ID:    userID,
		Email: "existing@example.com",
	}

	oauthUser := &session_domain.OAuthUser{
		ID:    "google_456",
		Email: "existing@example.com",
		Name:  "Existing User",
	}

	mockStateService.EXPECT().ValidateState(ctx, "valid_state").Return(nil)
	mockStateService.EXPECT().DeleteState(ctx, "valid_state").Return(nil)
	mockProvider.EXPECT().ExchangeCode(ctx, "valid_code").Return(oauthUser, nil)

	mockTxManager.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			mockProviderRepo.EXPECT().Find(ctx, user_domain.ProviderGoogle, "google_456").Return(existingProvider, nil)
			mockUserRepo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil)
			mockSessionRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
			return fn(ctx)
		},
	)

	refreshTokenUUID := uuid.New()
	mockTokenService.EXPECT().GenerateRefreshToken().Return(refreshTokenUUID)
	mockSessionEncrypter.EXPECT().HashRefreshToken(refreshTokenUUID.String()).Return("hashed_refresh")
	mockTokenService.EXPECT().GenerateAccessToken(gomock.Any()).Return("access_token_2", nil)

	handler := command.NewLoginGoogleHandler(
		mockTxManager,
		mockProvider,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockSessionRepo,
		mockTokenService,
		mockSessionEncrypter,
		mockStateService,
		24*time.Hour,
		15*time.Minute,
	)

	accessToken, refreshToken, err := handler.Handle(ctx, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "access_token_2", accessToken)
	assert.Equal(t, refreshTokenUUID.String(), refreshToken)
}

func TestLoginGoogleHandler_Handle_EmptyCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTxManager := mocks.NewMockTransactionManager(ctrl)
	mockProvider := mocks.NewMockOAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)
	mockStateService := mocks.NewMockStateService(ctrl)

	cmd := command.LoginGoogleCommand{
		Code:      "",
		State:     "valid_state",
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}

	handler := command.NewLoginGoogleHandler(
		mockTxManager,
		mockProvider,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockSessionRepo,
		mockTokenService,
		mockSessionEncrypter,
		mockStateService,
		24*time.Hour,
		15*time.Minute,
	)

	accessToken, refreshToken, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrInvalidOAuthCode, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLoginGoogleHandler_Handle_InvalidState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTxManager := mocks.NewMockTransactionManager(ctrl)
	mockProvider := mocks.NewMockOAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)
	mockStateService := mocks.NewMockStateService(ctrl)

	cmd := command.LoginGoogleCommand{
		Code:      "valid_code",
		State:     "invalid_state",
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}

	mockStateService.EXPECT().ValidateState(ctx, "invalid_state").Return(session_domain.ErrInvalidState)

	handler := command.NewLoginGoogleHandler(
		mockTxManager,
		mockProvider,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockSessionRepo,
		mockTokenService,
		mockSessionEncrypter,
		mockStateService,
		24*time.Hour,
		15*time.Minute,
	)

	accessToken, refreshToken, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrInvalidState, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLoginGoogleHandler_Handle_ExchangeCodeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTxManager := mocks.NewMockTransactionManager(ctrl)
	mockProvider := mocks.NewMockOAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)
	mockStateService := mocks.NewMockStateService(ctrl)

	cmd := command.LoginGoogleCommand{
		Code:      "invalid_code",
		State:     "valid_state",
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}

	mockStateService.EXPECT().ValidateState(ctx, "valid_state").Return(nil)
	mockStateService.EXPECT().DeleteState(ctx, "valid_state").Return(nil)
	mockProvider.EXPECT().ExchangeCode(ctx, "invalid_code").Return(nil, errors.New("oauth error"))

	handler := command.NewLoginGoogleHandler(
		mockTxManager,
		mockProvider,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockSessionRepo,
		mockTokenService,
		mockSessionEncrypter,
		mockStateService,
		24*time.Hour,
		15*time.Minute,
	)

	accessToken, refreshToken, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, "oauth error", err.Error())
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLoginGoogleHandler_Handle_TransactionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTxManager := mocks.NewMockTransactionManager(ctrl)
	mockProvider := mocks.NewMockOAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)
	mockStateService := mocks.NewMockStateService(ctrl)

	cmd := command.LoginGoogleCommand{
		Code:      "valid_code",
		State:     "valid_state",
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}

	oauthUser := &session_domain.OAuthUser{
		ID:    "google_789",
		Email: "error@example.com",
		Name:  "Error User",
	}

	mockStateService.EXPECT().ValidateState(ctx, "valid_state").Return(nil)
	mockStateService.EXPECT().DeleteState(ctx, "valid_state").Return(nil)
	mockProvider.EXPECT().ExchangeCode(ctx, "valid_code").Return(oauthUser, nil)
	mockTxManager.EXPECT().WithinTransaction(ctx, gomock.Any()).Return(errors.New("transaction error"))

	handler := command.NewLoginGoogleHandler(
		mockTxManager,
		mockProvider,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockSessionRepo,
		mockTokenService,
		mockSessionEncrypter,
		mockStateService,
		24*time.Hour,
		15*time.Minute,
	)

	accessToken, refreshToken, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, "transaction error", err.Error())
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLoginGoogleHandler_Handle_GenerateAccessTokenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTxManager := mocks.NewMockTransactionManager(ctrl)
	mockProvider := mocks.NewMockOAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)
	mockStateService := mocks.NewMockStateService(ctrl)

	cmd := command.LoginGoogleCommand{
		Code:      "valid_code",
		State:     "valid_state",
		UserAgent: "Mozilla/5.0",
		IPAddress: "192.168.1.1",
	}

	oauthUser := &session_domain.OAuthUser{
		ID:    "google_999",
		Email: "tokenError@example.com",
		Name:  "Token Error User",
	}

	mockStateService.EXPECT().ValidateState(ctx, "valid_state").Return(nil)
	mockStateService.EXPECT().DeleteState(ctx, "valid_state").Return(nil)
	mockProvider.EXPECT().ExchangeCode(ctx, "valid_code").Return(oauthUser, nil)

	mockTxManager.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			mockProviderRepo.EXPECT().Find(ctx, user_domain.ProviderGoogle, "google_999").Return(nil, user_domain.ErrNotFound)
			mockUserRepo.EXPECT().GetByEmail(ctx, "tokenError@example.com").Return(nil, user_domain.ErrNotFound)
			mockUserRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
			mockProviderRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(nil)
			mockProfileRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(nil)
			mockSessionRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
			return fn(ctx)
		},
	)

	refreshTokenUUID := uuid.New()
	mockTokenService.EXPECT().GenerateRefreshToken().Return(refreshTokenUUID)
	mockSessionEncrypter.EXPECT().HashRefreshToken(refreshTokenUUID.String()).Return("hashed_refresh")
	mockTokenService.EXPECT().GenerateAccessToken(gomock.Any()).Return("", session_domain.ErrTokenGenerate)

	handler := command.NewLoginGoogleHandler(
		mockTxManager,
		mockProvider,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockSessionRepo,
		mockTokenService,
		mockSessionEncrypter,
		mockStateService,
		24*time.Hour,
		15*time.Minute,
	)

	accessToken, refreshToken, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrTokenGenerate, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLoginGoogleHandler_Handle_NewUser_WithoutProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTxManager := mocks.NewMockTransactionManager(ctrl)
	mockProvider := mocks.NewMockOAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)
	mockStateService := mocks.NewMockStateService(ctrl)

	cmd := command.LoginGoogleCommand{
		Code:      "valid_code",
		State:     "valid_state",
		UserAgent: "Safari",
		IPAddress: "172.16.0.1",
	}

	oauthUser := &session_domain.OAuthUser{
		ID:    "google_noname",
		Email: "noname@example.com",
		Name:  "", // Empty name should skip profile creation
	}

	mockStateService.EXPECT().ValidateState(ctx, "valid_state").Return(nil)
	mockStateService.EXPECT().DeleteState(ctx, "valid_state").Return(nil)
	mockProvider.EXPECT().ExchangeCode(ctx, "valid_code").Return(oauthUser, nil)

	mockTxManager.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			mockProviderRepo.EXPECT().Find(ctx, user_domain.ProviderGoogle, "google_noname").Return(nil, user_domain.ErrNotFound)
			mockUserRepo.EXPECT().GetByEmail(ctx, "noname@example.com").Return(nil, user_domain.ErrNotFound)
			mockUserRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
			mockProviderRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(nil)
			// Profile creation should NOT be called
			mockSessionRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
			return fn(ctx)
		},
	)

	refreshTokenUUID := uuid.New()
	mockTokenService.EXPECT().GenerateRefreshToken().Return(refreshTokenUUID)
	mockSessionEncrypter.EXPECT().HashRefreshToken(refreshTokenUUID.String()).Return("hashed_refresh")
	mockTokenService.EXPECT().GenerateAccessToken(gomock.Any()).Return("access_token_noname", nil)

	handler := command.NewLoginGoogleHandler(
		mockTxManager,
		mockProvider,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockSessionRepo,
		mockTokenService,
		mockSessionEncrypter,
		mockStateService,
		24*time.Hour,
		15*time.Minute,
	)

	accessToken, refreshToken, err := handler.Handle(ctx, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "access_token_noname", accessToken)
	assert.Equal(t, refreshTokenUUID.String(), refreshToken)
}

func TestLoginGoogleHandler_Handle_ExistingUserByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockTxManager := mocks.NewMockTransactionManager(ctrl)
	mockProvider := mocks.NewMockOAuthProvider(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)
	mockStateService := mocks.NewMockStateService(ctrl)

	cmd := command.LoginGoogleCommand{
		Code:      "valid_code",
		State:     "valid_state",
		UserAgent: "Edge",
		IPAddress: "10.0.0.5",
	}

	userID := uuid.New()
	existingUser := &user_domain.User{
		ID:    userID,
		Email: "existing_email@example.com",
	}

	oauthUser := &session_domain.OAuthUser{
		ID:    "google_new_provider",
		Email: "existing_email@example.com",
		Name:  "Link Account",
	}

	mockStateService.EXPECT().ValidateState(ctx, "valid_state").Return(nil)
	mockStateService.EXPECT().DeleteState(ctx, "valid_state").Return(nil)
	mockProvider.EXPECT().ExchangeCode(ctx, "valid_code").Return(oauthUser, nil)

	mockTxManager.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			mockProviderRepo.EXPECT().Find(ctx, user_domain.ProviderGoogle, "google_new_provider").Return(nil, user_domain.ErrNotFound)
			mockUserRepo.EXPECT().GetByEmail(ctx, "existing_email@example.com").Return(existingUser, nil)
			mockProviderRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(nil)
			mockProfileRepo.EXPECT().Create(ctx, gomock.Any(), gomock.Any()).Return(nil)
			mockSessionRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
			return fn(ctx)
		},
	)

	refreshTokenUUID := uuid.New()
	mockTokenService.EXPECT().GenerateRefreshToken().Return(refreshTokenUUID)
	mockSessionEncrypter.EXPECT().HashRefreshToken(refreshTokenUUID.String()).Return("hashed_refresh")
	mockTokenService.EXPECT().GenerateAccessToken(gomock.Any()).Return("access_token_link", nil)

	handler := command.NewLoginGoogleHandler(
		mockTxManager,
		mockProvider,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockSessionRepo,
		mockTokenService,
		mockSessionEncrypter,
		mockStateService,
		24*time.Hour,
		15*time.Minute,
	)

	accessToken, refreshToken, err := handler.Handle(ctx, cmd)

	assert.NoError(t, err)
	assert.Equal(t, "access_token_link", accessToken)
	assert.Equal(t, refreshTokenUUID.String(), refreshToken)
}
