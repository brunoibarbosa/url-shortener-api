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

func TestLogoutHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	refreshToken := "valid_refresh_token"
	hashedToken := "hashed_token"
	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)

	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockBlacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
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
	mockSessionRepo.EXPECT().Revoke(ctx, sessionID).Return(nil)
	mockBlacklistRepo.EXPECT().Revoke(ctx, hashedToken, gomock.Any()).Return(nil)

	handler := command.NewLogoutHandler(
		mockSessionRepo,
		mockBlacklistRepo,
		mockSessionEncrypter,
	)

	cmd := command.LogoutCommand{
		RefreshToken: refreshToken,
	}

	err := handler.Handle(ctx, cmd)

	assert.NoError(t, err)
}

func TestLogoutHandler_Handle_InvalidRefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	refreshToken := "invalid_token"
	hashedToken := "hashed_invalid_token"

	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockBlacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
	mockSessionEncrypter := mocks.NewMockSessionEncrypter(ctrl)

	mockSessionEncrypter.EXPECT().HashRefreshToken(refreshToken).Return(hashedToken)
	mockSessionRepo.EXPECT().FindByRefreshToken(ctx, hashedToken).Return(nil, session_domain.ErrNotFound)

	handler := command.NewLogoutHandler(
		mockSessionRepo,
		mockBlacklistRepo,
		mockSessionEncrypter,
	)

	cmd := command.LogoutCommand{
		RefreshToken: refreshToken,
	}

	err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrInvalidRefreshToken, err)
}

func TestLogoutHandler_Handle_ExpiredSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	refreshToken := "expired_token"
	hashedToken := "hashed_expired_token"
	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired

	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockBlacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
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

	handler := command.NewLogoutHandler(
		mockSessionRepo,
		mockBlacklistRepo,
		mockSessionEncrypter,
	)

	cmd := command.LogoutCommand{
		RefreshToken: refreshToken,
	}

	err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrInvalidRefreshToken, err)
}

func TestLogoutHandler_Handle_RevokeSessionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	refreshToken := "valid_token"
	hashedToken := "hashed_token"
	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)
	expectedError := errors.New("revoke error")

	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockBlacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
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
	mockSessionRepo.EXPECT().Revoke(ctx, sessionID).Return(expectedError)

	handler := command.NewLogoutHandler(
		mockSessionRepo,
		mockBlacklistRepo,
		mockSessionEncrypter,
	)

	cmd := command.LogoutCommand{
		RefreshToken: refreshToken,
	}

	err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrRevokeFailed, err)
}

func TestLogoutHandler_Handle_BlacklistRetry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	refreshToken := "valid_token"
	hashedToken := "hashed_token"
	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour)
	blacklistError := errors.New("blacklist error")

	mockSessionRepo := mocks.NewMockSessionRepository(ctrl)
	mockBlacklistRepo := mocks.NewMockBlacklistRepository(ctrl)
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
	mockSessionRepo.EXPECT().Revoke(ctx, sessionID).Return(nil)
	
	// First call fails, retries also fail
	mockBlacklistRepo.EXPECT().Revoke(ctx, hashedToken, gomock.Any()).Return(blacklistError)
	mockBlacklistRepo.EXPECT().Revoke(ctx, hashedToken, gomock.Any()).Return(blacklistError).Times(3)

	handler := command.NewLogoutHandler(
		mockSessionRepo,
		mockBlacklistRepo,
		mockSessionEncrypter,
	)

	cmd := command.LogoutCommand{
		RefreshToken: refreshToken,
	}

	err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, session_domain.ErrRevokeFailed, err)
}
