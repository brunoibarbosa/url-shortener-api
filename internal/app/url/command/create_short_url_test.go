package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	"github.com/brunoibarbosa/url-shortener/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateShortURLHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	originalURL := "https://example.com"
	shortCode := "abc123"
	encryptedURL := "encrypted_url"

	mockRepo := mocks.NewMockURLRepository(ctrl)
	mockCache := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)
	mockGenerator := mocks.NewMockShortCodeGenerator(ctrl)

	mockGenerator.EXPECT().Generate(6).Return(shortCode, nil)
	mockCache.EXPECT().Exists(ctx, shortCode).Return(false, nil)
	mockRepo.EXPECT().Exists(ctx, shortCode).Return(false, nil)
	mockEncrypter.EXPECT().Encrypt(originalURL).Return(encryptedURL, nil)
	mockRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)
	mockCache.EXPECT().Save(ctx, gomock.Any(), gomock.Any()).Return(nil)

	handler := command.NewCreateShortURLHandler(
		mockRepo,
		mockCache,
		mockEncrypter,
		mockGenerator,
		24*time.Hour,
		1*time.Hour,
	)

	cmd := command.CreateShortURLCommand{
		OriginalURL: originalURL,
		UserID:      nil,
		Length:      6,
		MaxRetries:  10,
	}

	result, err := handler.Handle(ctx, cmd)

	assert.NoError(t, err)
	assert.Equal(t, shortCode, result.ShortCode)
}

func TestCreateShortURLHandler_Handle_CollisionRetry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	originalURL := "https://example.com"
	firstCode := "abc123"
	secondCode := "xyz789"
	encryptedURL := "encrypted_url"

	mockRepo := mocks.NewMockURLRepository(ctrl)
	mockCache := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)
	mockGenerator := mocks.NewMockShortCodeGenerator(ctrl)

	gomock.InOrder(
		mockGenerator.EXPECT().Generate(6).Return(firstCode, nil),
		mockGenerator.EXPECT().Generate(6).Return(secondCode, nil),
	)
	mockCache.EXPECT().Exists(ctx, firstCode).Return(true, nil)
	mockCache.EXPECT().Exists(ctx, secondCode).Return(false, nil)
	mockRepo.EXPECT().Exists(ctx, secondCode).Return(false, nil)
	mockEncrypter.EXPECT().Encrypt(originalURL).Return(encryptedURL, nil)
	mockRepo.EXPECT().Save(ctx, gomock.Any()).Return(nil)
	mockCache.EXPECT().Save(ctx, gomock.Any(), gomock.Any()).Return(nil)

	handler := command.NewCreateShortURLHandler(
		mockRepo,
		mockCache,
		mockEncrypter,
		mockGenerator,
		24*time.Hour,
		1*time.Hour,
	)

	cmd := command.CreateShortURLCommand{
		OriginalURL: originalURL,
		UserID:      nil,
		Length:      6,
		MaxRetries:  10,
	}

	result, err := handler.Handle(ctx, cmd)

	assert.NoError(t, err)
	assert.Equal(t, secondCode, result.ShortCode)
}

func TestCreateShortURLHandler_Handle_GeneratorError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	originalURL := "https://example.com"
	expectedError := errors.New("generator error")

	mockRepo := mocks.NewMockURLRepository(ctrl)
	mockCache := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)
	mockGenerator := mocks.NewMockShortCodeGenerator(ctrl)

	mockGenerator.EXPECT().Generate(6).Return("", expectedError)

	handler := command.NewCreateShortURLHandler(
		mockRepo,
		mockCache,
		mockEncrypter,
		mockGenerator,
		24*time.Hour,
		1*time.Hour,
	)

	cmd := command.CreateShortURLCommand{
		OriginalURL: originalURL,
		UserID:      nil,
		Length:      6,
		MaxRetries:  10,
	}

	result, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result.ShortCode)
}

func TestCreateShortURLHandler_Handle_EncryptionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	originalURL := "https://example.com"
	shortCode := "abc123"
	expectedError := errors.New("encryption error")

	mockRepo := mocks.NewMockURLRepository(ctrl)
	mockCache := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)
	mockGenerator := mocks.NewMockShortCodeGenerator(ctrl)

	mockGenerator.EXPECT().Generate(6).Return(shortCode, nil)
	mockCache.EXPECT().Exists(ctx, shortCode).Return(false, nil)
	mockRepo.EXPECT().Exists(ctx, shortCode).Return(false, nil)
	mockEncrypter.EXPECT().Encrypt(originalURL).Return("", expectedError)

	handler := command.NewCreateShortURLHandler(
		mockRepo,
		mockCache,
		mockEncrypter,
		mockGenerator,
		24*time.Hour,
		1*time.Hour,
	)

	cmd := command.CreateShortURLCommand{
		OriginalURL: originalURL,
		UserID:      nil,
		Length:      6,
		MaxRetries:  10,
	}

	result, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result.ShortCode)
}

func TestCreateShortURLHandler_Handle_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	originalURL := "https://example.com"
	shortCode := "abc123"
	encryptedURL := "encrypted_url"
	expectedError := errors.New("save error")

	mockRepo := mocks.NewMockURLRepository(ctrl)
	mockCache := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)
	mockGenerator := mocks.NewMockShortCodeGenerator(ctrl)

	mockGenerator.EXPECT().Generate(6).Return(shortCode, nil)
	mockCache.EXPECT().Exists(ctx, shortCode).Return(false, nil)
	mockRepo.EXPECT().Exists(ctx, shortCode).Return(false, nil)
	mockEncrypter.EXPECT().Encrypt(originalURL).Return(encryptedURL, nil)
	mockCache.EXPECT().Save(ctx, gomock.Any(), gomock.Any()).Return(nil)
	mockRepo.EXPECT().Save(ctx, gomock.Any()).Return(expectedError)
	mockCache.EXPECT().Delete(ctx, shortCode).Return(nil) // Cleanup on error

	handler := command.NewCreateShortURLHandler(
		mockRepo,
		mockCache,
		mockEncrypter,
		mockGenerator,
		24*time.Hour,
		1*time.Hour,
	)

	cmd := command.CreateShortURLCommand{
		OriginalURL: originalURL,
		UserID:      nil,
		Length:      6,
		MaxRetries:  10,
	}

	result, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result.ShortCode)
}
