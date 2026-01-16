package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	user_domain "github.com/brunoibarbosa/url-shortener/internal/domain/user"
	"github.com/brunoibarbosa/url-shortener/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestLoginUserHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userID := uuid.New()
	email := "user@example.com"
	password := "password123"
	passwordHash := "hashed_password"
	accessToken := "access_token_string"

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockPasswordEncrypter := mocks.NewMockUserPasswordEncrypter(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

	provider := &user_domain.UserProvider{
		UserID:       userID,
		Provider:     user_domain.ProviderPassword,
		ProviderID:   email,
		PasswordHash: &passwordHash,
	}

	mockProviderRepo.EXPECT().Find(ctx, user_domain.ProviderPassword, email).Return(provider, nil)
	mockPasswordEncrypter.EXPECT().CheckPassword(passwordHash, password).Return(true)

	mockTx.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)

	mockTokenService.EXPECT().GenerateRefreshToken().Return(uuid.New())
	mockSessionEncrypter.EXPECT().HashRefreshToken(gomock.Any()).Return("hashed_refresh_token")
	mockSessionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	mockTokenService.EXPECT().GenerateAccessToken(gomock.Any()).Return(accessToken, nil)

	handler := command.NewLoginUserHandler(
		mockTx,
		mockProviderRepo,
		mockSessionRepo,
		mockTokenService,
		mockPasswordEncrypter,
		mockSessionEncrypter,
		24*time.Hour,
		15*time.Minute,
	)

	cmd := command.LoginUserCommand{
		Email:     email,
		Password:  password,
		UserAgent: "Mozilla/5.0",
		IPAddress: "127.0.0.1",
	}

	access, refresh, err := handler.Handle(ctx, cmd)

	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
}

func TestLoginUserHandler_Handle_EmptyCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockPasswordEncrypter := mocks.NewMockUserPasswordEncrypter(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

	handler := command.NewLoginUserHandler(
		mockTx,
		mockProviderRepo,
		mockSessionRepo,
		mockTokenService,
		mockPasswordEncrypter,
		mockSessionEncrypter,
		24*time.Hour,
		15*time.Minute,
	)

	cmd := command.LoginUserCommand{
		Email:     "",
		Password:  "password",
		UserAgent: "Mozilla/5.0",
		IPAddress: "127.0.0.1",
	}

	access, refresh, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, user_domain.ErrInvalidCredentials, err)
	assert.Empty(t, access)
	assert.Empty(t, refresh)
}

func TestLoginUserHandler_Handle_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	email := "notfound@example.com"
	password := "password123"

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockPasswordEncrypter := mocks.NewMockUserPasswordEncrypter(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

	mockProviderRepo.EXPECT().Find(ctx, user_domain.ProviderPassword, email).Return(nil, user_domain.ErrNotFound)

	handler := command.NewLoginUserHandler(
		mockTx,
		mockProviderRepo,
		mockSessionRepo,
		mockTokenService,
		mockPasswordEncrypter,
		mockSessionEncrypter,
		24*time.Hour,
		15*time.Minute,
	)

	cmd := command.LoginUserCommand{
		Email:     email,
		Password:  password,
		UserAgent: "Mozilla/5.0",
		IPAddress: "127.0.0.1",
	}

	access, refresh, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, user_domain.ErrInvalidCredentials, err)
	assert.Empty(t, access)
	assert.Empty(t, refresh)
}

func TestLoginUserHandler_Handle_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	userID := uuid.New()
	email := "user@example.com"
	password := "wrongpassword"
	passwordHash := "hashed_password"

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockPasswordEncrypter := mocks.NewMockUserPasswordEncrypter(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

	provider := &user_domain.UserProvider{
		UserID:       userID,
		Provider:     user_domain.ProviderPassword,
		ProviderID:   email,
		PasswordHash: &passwordHash,
	}

	mockProviderRepo.EXPECT().Find(ctx, user_domain.ProviderPassword, email).Return(provider, nil)
	mockPasswordEncrypter.EXPECT().CheckPassword(passwordHash, password).Return(false)

	handler := command.NewLoginUserHandler(
		mockTx,
		mockProviderRepo,
		mockSessionRepo,
		mockTokenService,
		mockPasswordEncrypter,
		mockSessionEncrypter,
		24*time.Hour,
		15*time.Minute,
	)

	cmd := command.LoginUserCommand{
		Email:     email,
		Password:  password,
		UserAgent: "Mozilla/5.0",
		IPAddress: "127.0.0.1",
	}

	access, refresh, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, user_domain.ErrInvalidCredentials, err)
	assert.Empty(t, access)
	assert.Empty(t, refresh)
}

func TestLoginUserHandler_Handle_GenerateAccessTokenError(t *testing.T) {
ctrl := gomock.NewController(t)
defer ctrl.Finish()

ctx := context.Background()
userID := uuid.New()
email := "user@example.com"
password := "password123"
passwordHash := "hashed_password"

mockTx := mocks.NewMockTransactionManager(ctrl)
mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
mockTokenService := mocks.NewMockTokenService(ctrl)
mockPasswordEncrypter := mocks.NewMockUserPasswordEncrypter(ctrl)
mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

provider := &user_domain.UserProvider{
UserID:       userID,
Provider:     user_domain.ProviderPassword,
ProviderID:   email,
PasswordHash: &passwordHash,
}

mockProviderRepo.EXPECT().Find(ctx, user_domain.ProviderPassword, email).Return(provider, nil)
mockPasswordEncrypter.EXPECT().CheckPassword(passwordHash, password).Return(true)

mockTx.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
func(ctx context.Context, fn func(context.Context) error) error {
return fn(ctx)
},
)

mockTokenService.EXPECT().GenerateRefreshToken().Return(uuid.New())
mockSessionEncrypter.EXPECT().HashRefreshToken(gomock.Any()).Return("hashed_refresh_token")
mockSessionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
mockTokenService.EXPECT().GenerateAccessToken(gomock.Any()).Return("", errors.New("token generation failed"))

handler := command.NewLoginUserHandler(
mockTx,
mockProviderRepo,
mockSessionRepo,
mockTokenService,
mockPasswordEncrypter,
mockSessionEncrypter,
24*time.Hour,
15*time.Minute,
)

cmd := command.LoginUserCommand{
Email:     email,
Password:  password,
UserAgent: "Mozilla/5.0",
IPAddress: "127.0.0.1",
}

access, refresh, err := handler.Handle(ctx, cmd)

assert.Error(t, err)
assert.Equal(t, "token generation failed", err.Error())
assert.Empty(t, access)
assert.Empty(t, refresh)
}
