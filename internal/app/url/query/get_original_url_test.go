package query_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/query"
	domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/internal/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetOriginalURLHandler_Handle_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockPersistRepo := mocks.NewMockURLRepository(ctrl)
	mockCacheRepo := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)

	shortCode := "abc123"
	encryptedURL := "encrypted_original_url"
	originalURL := "https://example.com"

	cachedURL := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: encryptedURL,
		ExpiresAt:    nil,
		DeletedAt:    nil,
	}

	mockCacheRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(cachedURL, nil)
	mockEncrypter.EXPECT().Decrypt(encryptedURL).Return(originalURL, nil)

	handler := query.NewGetOriginalURLHandler(
		mockPersistRepo,
		mockCacheRepo,
		mockEncrypter,
		1*time.Hour,
	)

	q := query.GetOriginalURLQuery{ShortCode: shortCode}
	result, err := handler.Handle(ctx, q)

	assert.NoError(t, err)
	assert.Equal(t, originalURL, result)
}

func TestGetOriginalURLHandler_Handle_CacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockPersistRepo := mocks.NewMockURLRepository(ctrl)
	mockCacheRepo := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)

	shortCode := "xyz789"
	encryptedURL := "encrypted_url"
	originalURL := "https://example.org"

	url := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: encryptedURL,
		ExpiresAt:    nil,
		DeletedAt:    nil,
	}

	mockCacheRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(nil, nil)
	mockPersistRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(url, nil)
	mockCacheRepo.EXPECT().Save(ctx, url, 1*time.Hour).Return(nil)
	mockEncrypter.EXPECT().Decrypt(encryptedURL).Return(originalURL, nil)

	handler := query.NewGetOriginalURLHandler(
		mockPersistRepo,
		mockCacheRepo,
		mockEncrypter,
		1*time.Hour,
	)

	q := query.GetOriginalURLQuery{ShortCode: shortCode}
	result, err := handler.Handle(ctx, q)

	assert.NoError(t, err)
	assert.Equal(t, originalURL, result)
}

func TestGetOriginalURLHandler_Handle_URLNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockPersistRepo := mocks.NewMockURLRepository(ctrl)
	mockCacheRepo := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)

	shortCode := "notfound"

	mockCacheRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(nil, nil)
	mockPersistRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(nil, nil)

	handler := query.NewGetOriginalURLHandler(
		mockPersistRepo,
		mockCacheRepo,
		mockEncrypter,
		1*time.Hour,
	)

	q := query.GetOriginalURLQuery{ShortCode: shortCode}
	result, err := handler.Handle(ctx, q)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrURLNotFound, err)
	assert.Empty(t, result)
}

func TestGetOriginalURLHandler_Handle_ExpiredURL_Cache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockPersistRepo := mocks.NewMockURLRepository(ctrl)
	mockCacheRepo := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)

	shortCode := "expired"
	expiresAt := time.Now().Add(-1 * time.Hour)

	cachedURL := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: "encrypted",
		ExpiresAt:    &expiresAt,
		DeletedAt:    nil,
	}

	mockCacheRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(cachedURL, nil)

	handler := query.NewGetOriginalURLHandler(
		mockPersistRepo,
		mockCacheRepo,
		mockEncrypter,
		1*time.Hour,
	)

	q := query.GetOriginalURLQuery{ShortCode: shortCode}
	result, err := handler.Handle(ctx, q)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrExpiredURL, err)
	assert.Empty(t, result)
}

func TestGetOriginalURLHandler_Handle_DeletedURL_Cache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockPersistRepo := mocks.NewMockURLRepository(ctrl)
	mockCacheRepo := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)

	shortCode := "deleted"
	deletedAt := time.Now()

	cachedURL := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: "encrypted",
		ExpiresAt:    nil,
		DeletedAt:    &deletedAt,
	}

	mockCacheRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(cachedURL, nil)

	handler := query.NewGetOriginalURLHandler(
		mockPersistRepo,
		mockCacheRepo,
		mockEncrypter,
		1*time.Hour,
	)

	q := query.GetOriginalURLQuery{ShortCode: shortCode}
	result, err := handler.Handle(ctx, q)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrDeletedURL, err)
	assert.Empty(t, result)
}

func TestGetOriginalURLHandler_Handle_ExpiredURL_Persist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockPersistRepo := mocks.NewMockURLRepository(ctrl)
	mockCacheRepo := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)

	shortCode := "expired"
	expiresAt := time.Now().Add(-1 * time.Hour)

	url := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: "encrypted",
		ExpiresAt:    &expiresAt,
		DeletedAt:    nil,
	}

	mockCacheRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(nil, nil)
	mockPersistRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(url, nil)

	handler := query.NewGetOriginalURLHandler(
		mockPersistRepo,
		mockCacheRepo,
		mockEncrypter,
		1*time.Hour,
	)

	q := query.GetOriginalURLQuery{ShortCode: shortCode}
	result, err := handler.Handle(ctx, q)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrExpiredURL, err)
	assert.Empty(t, result)
}

func TestGetOriginalURLHandler_Handle_CacheWithExpiringURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockPersistRepo := mocks.NewMockURLRepository(ctrl)
	mockCacheRepo := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)

	shortCode := "expiring"
	encryptedURL := "encrypted"
	originalURL := "https://expiring.com"
	expiresAt := time.Now().Add(30 * time.Minute) // Expires in 30 minutes

	url := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: encryptedURL,
		ExpiresAt:    &expiresAt,
		DeletedAt:    nil,
	}

	mockCacheRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(nil, nil)
	mockPersistRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(url, nil)
	// Should cache for 30 minutes (remaining TTL) instead of 1 hour
	mockCacheRepo.EXPECT().Save(ctx, url, gomock.Any()).Return(nil)
	mockEncrypter.EXPECT().Decrypt(encryptedURL).Return(originalURL, nil)

	handler := query.NewGetOriginalURLHandler(
		mockPersistRepo,
		mockCacheRepo,
		mockEncrypter,
		1*time.Hour,
	)

	q := query.GetOriginalURLQuery{ShortCode: shortCode}
	result, err := handler.Handle(ctx, q)

	assert.NoError(t, err)
	assert.Equal(t, originalURL, result)
}

func TestGetOriginalURLHandler_Handle_CacheError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockPersistRepo := mocks.NewMockURLRepository(ctrl)
	mockCacheRepo := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)

	shortCode := "cache_error"

	mockCacheRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(nil, errors.New("cache error"))

	handler := query.NewGetOriginalURLHandler(
		mockPersistRepo,
		mockCacheRepo,
		mockEncrypter,
		1*time.Hour,
	)

	q := query.GetOriginalURLQuery{ShortCode: shortCode}
	result, err := handler.Handle(ctx, q)

	assert.Error(t, err)
	assert.Equal(t, "cache error", err.Error())
	assert.Empty(t, result)
}

func TestGetOriginalURLHandler_Handle_PersistError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockPersistRepo := mocks.NewMockURLRepository(ctrl)
	mockCacheRepo := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)

	shortCode := "persist_error"

	mockCacheRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(nil, nil)
	mockPersistRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(nil, errors.New("database error"))

	handler := query.NewGetOriginalURLHandler(
		mockPersistRepo,
		mockCacheRepo,
		mockEncrypter,
		1*time.Hour,
	)

	q := query.GetOriginalURLQuery{ShortCode: shortCode}
	result, err := handler.Handle(ctx, q)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Empty(t, result)
}

func TestGetOriginalURLHandler_Handle_DecryptError_Cache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockPersistRepo := mocks.NewMockURLRepository(ctrl)
	mockCacheRepo := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)

	shortCode := "decrypt_error"
	encryptedURL := "corrupted_encrypted"

	cachedURL := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: encryptedURL,
		ExpiresAt:    nil,
		DeletedAt:    nil,
	}

	mockCacheRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(cachedURL, nil)
	mockEncrypter.EXPECT().Decrypt(encryptedURL).Return("", errors.New("decryption failed"))

	handler := query.NewGetOriginalURLHandler(
		mockPersistRepo,
		mockCacheRepo,
		mockEncrypter,
		1*time.Hour,
	)

	q := query.GetOriginalURLQuery{ShortCode: shortCode}
	result, err := handler.Handle(ctx, q)

	assert.Error(t, err)
	assert.Equal(t, "decryption failed", err.Error())
	assert.Empty(t, result)
}

func TestGetOriginalURLHandler_Handle_DecryptError_Persist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockPersistRepo := mocks.NewMockURLRepository(ctrl)
	mockCacheRepo := mocks.NewMockURLCacheRepository(ctrl)
	mockEncrypter := mocks.NewMockURLEncrypter(ctrl)

	shortCode := "decrypt_error"
	encryptedURL := "corrupted"

	url := &domain.URL{
		ShortCode:    shortCode,
		EncryptedURL: encryptedURL,
		ExpiresAt:    nil,
		DeletedAt:    nil,
	}

	mockCacheRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(nil, nil)
	mockPersistRepo.EXPECT().FindByShortCode(ctx, shortCode).Return(url, nil)
	mockCacheRepo.EXPECT().Save(ctx, url, 1*time.Hour).Return(nil)
	mockEncrypter.EXPECT().Decrypt(encryptedURL).Return("", errors.New("decryption failed"))

	handler := query.NewGetOriginalURLHandler(
		mockPersistRepo,
		mockCacheRepo,
		mockEncrypter,
		1*time.Hour,
	)

	q := query.GetOriginalURLQuery{ShortCode: shortCode}
	result, err := handler.Handle(ctx, q)

	assert.Error(t, err)
	assert.Equal(t, "decryption failed", err.Error())
	assert.Empty(t, result)
}
