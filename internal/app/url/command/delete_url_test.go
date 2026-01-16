package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/command"
	"github.com/brunoibarbosa/url-shortener/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDeleteURLHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	urlID := uuid.New()
	userID := uuid.New()
	shortCode := "abc123"

	mockRepo := mocks.NewMockURLRepository(ctrl)
	mockCache := mocks.NewMockURLCacheRepository(ctrl)

	mockRepo.EXPECT().SoftDelete(ctx, urlID, userID).Return(shortCode, nil)
	mockCache.EXPECT().Delete(ctx, shortCode).Return(nil)

	handler := command.NewDeleteURLHandler(mockRepo, mockCache)

	cmd := command.DeleteURLCommand{
		ID:     urlID,
		UserID: userID,
	}

	err := handler.Handle(ctx, cmd)

	assert.NoError(t, err)
}

func TestDeleteURLHandler_Handle_SoftDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	urlID := uuid.New()
	userID := uuid.New()
	expectedError := errors.New("soft delete error")

	mockRepo := mocks.NewMockURLRepository(ctrl)
	mockCache := mocks.NewMockURLCacheRepository(ctrl)

	mockRepo.EXPECT().SoftDelete(ctx, urlID, userID).Return("", expectedError)

	handler := command.NewDeleteURLHandler(mockRepo, mockCache)

	cmd := command.DeleteURLCommand{
		ID:     urlID,
		UserID: userID,
	}

	err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestDeleteURLHandler_Handle_CacheDeleteFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	urlID := uuid.New()
	userID := uuid.New()
	shortCode := "abc123"
	cacheError := errors.New("cache delete error")

	mockRepo := mocks.NewMockURLRepository(ctrl)
	mockCache := mocks.NewMockURLCacheRepository(ctrl)

	mockRepo.EXPECT().SoftDelete(ctx, urlID, userID).Return(shortCode, nil)
	mockCache.EXPECT().Delete(ctx, shortCode).Return(cacheError)

	handler := command.NewDeleteURLHandler(mockRepo, mockCache)

	cmd := command.DeleteURLCommand{
		ID:     urlID,
		UserID: userID,
	}

	err := handler.Handle(ctx, cmd)

	// Cache delete error is ignored (error silenciado com _)
	assert.NoError(t, err)
}

func TestDeleteURLHandler_Handle_WrongOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	urlID := uuid.New()
	userID := uuid.New()
	expectedError := errors.New("no rows affected")

	mockRepo := mocks.NewMockURLRepository(ctrl)
	mockCache := mocks.NewMockURLCacheRepository(ctrl)

	mockRepo.EXPECT().SoftDelete(ctx, urlID, userID).Return("", expectedError)

	handler := command.NewDeleteURLHandler(mockRepo, mockCache)

	cmd := command.DeleteURLCommand{
		ID:     urlID,
		UserID: userID,
	}

	err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}
