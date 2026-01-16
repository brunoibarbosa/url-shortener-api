package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRefreshTokenHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	oldRefreshToken := "old_refresh_token"
	hashedOldToken := "hashed_old_token"
	hashedNewToken := "hashed_new_token"
	accessToken := "new_access_token"
	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockBlacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

	session := &session_domain.Session{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: hashedOldToken,
		ExpiresAt:        &expiresAt,
		RevokedAt:        nil,
	}

	mockSessionEncrypter.EXPECT().HashRefreshToken(oldRefreshToken).Return(hashedOldToken)
	mockSessionRepo.EXPECT().FindByRefreshToken(ctx, hashedOldToken).Return(session, nil)
	mockBlacklistRepo.EXPECT().IsRevoked(ctx, hashedOldToken).Return(false, nil)

	mockTx.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)

	mockSessionRepo.EXPECT().Revoke(gomock.Any(), sessionID).Return(nil)
	mockBlacklistRepo.EXPECT().Revoke(gomock.Any(), hashedOldToken, gomock.Any()).Return(nil)
	mockTokenService.EXPECT().GenerateRefreshToken().Return(uuid.New())
	mockSessionEncrypter.EXPECT().HashRefreshToken(gomock.Any()).Return(hashedNewToken)
	mockSessionRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	mockTokenService.EXPECT().GenerateAccessToken(gomock.Any()).Return(accessToken, nil)

	handler := command.NewRefreshTokenHandler(
		mockTx,
		mockSessionRepo,
		mockBlacklistRepo,
		mockTokenService,
		mockSessionEncrypter,
		24*time.Hour,
		15*time.Minute,
	)

	cmd := command.RefreshTokenCommand{
		RefreshToken: oldRefreshToken,
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "127.0.0.1",
	}

	result, err := handler.Handle(ctx, cmd)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
}

func TestRefreshTokenHandler_Handle_EmptyRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockBlacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

	handler := command.NewRefreshTokenHandler(
		mockTx,
		mockSessionRepo,
		mockBlacklistRepo,
		mockTokenService,
		mockSessionEncrypter,
		24*time.Hour,
		15*time.Minute,
	)

	cmd := command.RefreshTokenCommand{
		RefreshToken: "",
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "127.0.0.1",
	}

	result, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrInvalidRefreshToken, err)
	assert.Empty(t, result.AccessToken)
	assert.Empty(t, result.RefreshToken)
}

func TestRefreshTokenHandler_Handle_SessionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	refreshToken := "invalid_token"
	hashedToken := "hashed_invalid_token"

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockBlacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

	mockSessionEncrypter.EXPECT().HashRefreshToken(refreshToken).Return(hashedToken)
	mockSessionRepo.EXPECT().FindByRefreshToken(ctx, hashedToken).Return(nil, session_domain.ErrNotFound)

	handler := command.NewRefreshTokenHandler(
		mockTx,
		mockSessionRepo,
		mockBlacklistRepo,
		mockTokenService,
		mockSessionEncrypter,
		24*time.Hour,
		15*time.Minute,
	)

	cmd := command.RefreshTokenCommand{
		RefreshToken: refreshToken,
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "127.0.0.1",
	}

	result, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrInvalidRefreshToken, err)
	assert.Empty(t, result.AccessToken)
	assert.Empty(t, result.RefreshToken)
}

func TestRefreshTokenHandler_Handle_ExpiredSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	refreshToken := "expired_token"
	hashedToken := "hashed_expired_token"
	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockBlacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

	session := &session_domain.Session{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: hashedToken,
		ExpiresAt:        &expiresAt,
		RevokedAt:        nil,
	}

	mockSessionEncrypter.EXPECT().HashRefreshToken(refreshToken).Return(hashedToken)
	mockSessionRepo.EXPECT().FindByRefreshToken(ctx, hashedToken).Return(session, nil)

	handler := command.NewRefreshTokenHandler(
		mockTx,
		mockSessionRepo,
		mockBlacklistRepo,
		mockTokenService,
		mockSessionEncrypter,
		24*time.Hour,
		15*time.Minute,
	)

	cmd := command.RefreshTokenCommand{
		RefreshToken: refreshToken,
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "127.0.0.1",
	}

	result, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrInvalidRefreshToken, err)
	assert.Empty(t, result.AccessToken)
	assert.Empty(t, result.RefreshToken)
}

func TestRefreshTokenHandler_Handle_RevokedToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	refreshToken := "revoked_token"
	hashedToken := "hashed_revoked_token"
	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockBlacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

	session := &session_domain.Session{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: hashedToken,
		ExpiresAt:        &expiresAt,
		RevokedAt:        nil,
	}

	mockSessionEncrypter.EXPECT().HashRefreshToken(refreshToken).Return(hashedToken)
	mockSessionRepo.EXPECT().FindByRefreshToken(ctx, hashedToken).Return(session, nil)
	mockBlacklistRepo.EXPECT().IsRevoked(ctx, hashedToken).Return(true, nil)

	handler := command.NewRefreshTokenHandler(
		mockTx,
		mockSessionRepo,
		mockBlacklistRepo,
		mockTokenService,
		mockSessionEncrypter,
		24*time.Hour,
		15*time.Minute,
	)

	cmd := command.RefreshTokenCommand{
		RefreshToken: refreshToken,
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "127.0.0.1",
	}

	result, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrInvalidRefreshToken, err)
	assert.Empty(t, result.AccessToken)
	assert.Empty(t, result.RefreshToken)
}

func TestRefreshTokenHandler_Handle_TransactionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	refreshToken := "valid_token"
	hashedToken := "hashed_token"
	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)
	expectedError := errors.New("transaction error")

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockBlacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
	mockTokenService := mocks.NewMockTokenService(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

	session := &session_domain.Session{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: hashedToken,
		ExpiresAt:        &expiresAt,
		RevokedAt:        nil,
	}

	mockSessionEncrypter.EXPECT().HashRefreshToken(refreshToken).Return(hashedToken)
	mockSessionRepo.EXPECT().FindByRefreshToken(ctx, hashedToken).Return(session, nil)
	mockBlacklistRepo.EXPECT().IsRevoked(ctx, hashedToken).Return(false, nil)
	mockTx.EXPECT().WithinTransaction(ctx, gomock.Any()).Return(expectedError)

	handler := command.NewRefreshTokenHandler(
		mockTx,
		mockSessionRepo,
		mockBlacklistRepo,
		mockTokenService,
		mockSessionEncrypter,
		24*time.Hour,
		15*time.Minute,
	)

	cmd := command.RefreshTokenCommand{
		RefreshToken: refreshToken,
		UserAgent:    "Mozilla/5.0",
		IPAddress:    "127.0.0.1",
	}

	result, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result.AccessToken)
	assert.Empty(t, result.RefreshToken)
}
