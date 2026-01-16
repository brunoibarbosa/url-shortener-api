package query_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/url/query"
	"github.com/brunoibarbosa/url-shortener/internal/domain"
	url_domain "github.com/brunoibarbosa/url-shortener/internal/domain/url"
	"github.com/brunoibarbosa/url-shortener/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestListUserURLsHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockURLQueryRepository(ctrl)

	userID := uuid.New()
	params := url_domain.ListURLsParams{
		SortBy:   url_domain.ListURLsSortByCreatedAt,
		SortKind: domain.SortDesc,
		Pagination: domain.Pagination{
			Number: 1,
			Size:   10,
		},
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	expectedURLs := []url_domain.ListURLsDTO{
		{
			ID:        uuid.New(),
			ShortCode: "abc123",
			ExpiresAt: &expiresAt,
			CreatedAt: now,
			DeletedAt: nil,
		},
		{
			ID:        uuid.New(),
			ShortCode: "xyz789",
			ExpiresAt: nil,
			CreatedAt: now.Add(-1 * time.Hour),
			DeletedAt: nil,
		},
	}
	var expectedTotal uint64 = 2

	mockRepo.EXPECT().ListByUserID(ctx, userID, params).Return(expectedURLs, expectedTotal, nil)

	handler := query.NewListUserURLsHandler(mockRepo)

	urls, total, err := handler.Handle(ctx, userID, params)

	assert.NoError(t, err)
	assert.Equal(t, expectedURLs, urls)
	assert.Equal(t, expectedTotal, total)
	assert.Len(t, urls, 2)
}

func TestListUserURLsHandler_Handle_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockURLQueryRepository(ctrl)

	userID := uuid.New()
	params := url_domain.ListURLsParams{
		SortBy:   url_domain.ListURLsSortByNone,
		SortKind: domain.SortAsc,
		Pagination: domain.Pagination{
			Number: 1,
			Size:   10,
		},
	}

	var expectedTotal uint64 = 0
	emptyURLs := []url_domain.ListURLsDTO{}

	mockRepo.EXPECT().ListByUserID(ctx, userID, params).Return(emptyURLs, expectedTotal, nil)

	handler := query.NewListUserURLsHandler(mockRepo)

	urls, total, err := handler.Handle(ctx, userID, params)

	assert.NoError(t, err)
	assert.Empty(t, urls)
	assert.Equal(t, expectedTotal, total)
}

func TestListUserURLsHandler_Handle_WithPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockURLQueryRepository(ctrl)

	userID := uuid.New()
	params := url_domain.ListURLsParams{
		SortBy:   url_domain.ListURLsSortByExpiresAt,
		SortKind: domain.SortAsc,
		Pagination: domain.Pagination{
			Number: 2,
			Size:   5,
		},
	}

	now := time.Now()
	expectedURLs := []url_domain.ListURLsDTO{
		{
			ID:        uuid.New(),
			ShortCode: "page2_1",
			ExpiresAt: nil,
			CreatedAt: now,
			DeletedAt: nil,
		},
	}
	var expectedTotal uint64 = 6 // Total across all pages

	mockRepo.EXPECT().ListByUserID(ctx, userID, params).Return(expectedURLs, expectedTotal, nil)

	handler := query.NewListUserURLsHandler(mockRepo)

	urls, total, err := handler.Handle(ctx, userID, params)

	assert.NoError(t, err)
	assert.Len(t, urls, 1)
	assert.Equal(t, expectedTotal, total)
}

func TestListUserURLsHandler_Handle_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockURLQueryRepository(ctrl)

	userID := uuid.New()
	params := url_domain.ListURLsParams{
		SortBy:   url_domain.ListURLsSortByCreatedAt,
		SortKind: domain.SortDesc,
		Pagination: domain.Pagination{
			Number: 1,
			Size:   10,
		},
	}

	mockRepo.EXPECT().ListByUserID(ctx, userID, params).Return(nil, uint64(0), errors.New("database error"))

	handler := query.NewListUserURLsHandler(mockRepo)

	urls, total, err := handler.Handle(ctx, userID, params)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Nil(t, urls)
	assert.Equal(t, uint64(0), total)
}

func TestListUserURLsHandler_Handle_WithDeletedURLs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockURLQueryRepository(ctrl)

	userID := uuid.New()
	params := url_domain.ListURLsParams{
		SortBy:   url_domain.ListURLsSortByCreatedAt,
		SortKind: domain.SortDesc,
		Pagination: domain.Pagination{
			Number: 1,
			Size:   10,
		},
	}

	now := time.Now()
	deletedAt := now.Add(-1 * time.Hour)
	expectedURLs := []url_domain.ListURLsDTO{
		{
			ID:        uuid.New(),
			ShortCode: "active",
			ExpiresAt: nil,
			CreatedAt: now,
			DeletedAt: nil,
		},
		{
			ID:        uuid.New(),
			ShortCode: "deleted",
			ExpiresAt: nil,
			CreatedAt: now.Add(-2 * time.Hour),
			DeletedAt: &deletedAt,
		},
	}
	var expectedTotal uint64 = 2

	mockRepo.EXPECT().ListByUserID(ctx, userID, params).Return(expectedURLs, expectedTotal, nil)

	handler := query.NewListUserURLsHandler(mockRepo)

	urls, total, err := handler.Handle(ctx, userID, params)

	assert.NoError(t, err)
	assert.Len(t, urls, 2)
	assert.Nil(t, urls[0].DeletedAt)
	assert.NotNil(t, urls[1].DeletedAt)
	assert.Equal(t, expectedTotal, total)
}

func TestListUserURLsHandler_Handle_DifferentSortOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockURLQueryRepository(ctrl)

	userID := uuid.New()

	t.Run("sort by created_at ascending", func(t *testing.T) {
		params := url_domain.ListURLsParams{
			SortBy:   url_domain.ListURLsSortByCreatedAt,
			SortKind: domain.SortAsc,
			Pagination: domain.Pagination{
				Number: 1,
				Size:   10,
			},
		}

		mockRepo.EXPECT().ListByUserID(ctx, userID, params).Return([]url_domain.ListURLsDTO{}, uint64(0), nil)

		handler := query.NewListUserURLsHandler(mockRepo)
		urls, total, err := handler.Handle(ctx, userID, params)

		assert.NoError(t, err)
		assert.Empty(t, urls)
		assert.Equal(t, uint64(0), total)
	})

	t.Run("sort by expires_at descending", func(t *testing.T) {
		params := url_domain.ListURLsParams{
			SortBy:   url_domain.ListURLsSortByExpiresAt,
			SortKind: domain.SortDesc,
			Pagination: domain.Pagination{
				Number: 1,
				Size:   10,
			},
		}

		mockRepo.EXPECT().ListByUserID(ctx, userID, params).Return([]url_domain.ListURLsDTO{}, uint64(0), nil)

		handler := query.NewListUserURLsHandler(mockRepo)
		urls, total, err := handler.Handle(ctx, userID, params)

		assert.NoError(t, err)
		assert.Empty(t, urls)
		assert.Equal(t, uint64(0), total)
	})
}
