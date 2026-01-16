package query_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brunoibarbosa/url-shortener/internal/app/session/query"
	"github.com/brunoibarbosa/url-shortener/internal/domain"
	session_domain "github.com/brunoibarbosa/url-shortener/internal/domain/session"
	"github.com/brunoibarbosa/url-shortener/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestListSessionsHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockSessionQueryRepository(ctrl)

	params := session_domain.ListSessionsParams{
		SortBy:   session_domain.ListSessionsSortByCreatedAt,
		SortKind: domain.SortDesc,
		Pagination: domain.Pagination{
			Number: 1,
			Size:   10,
		},
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	expectedSessions := []session_domain.ListSessionsDTO{
		{
			UserAgent: "Mozilla/5.0",
			IPAddress: "192.168.1.1",
			CreatedAt: now,
			ExpiresAt: expiresAt,
		},
		{
			UserAgent: "Chrome/90.0",
			IPAddress: "192.168.1.2",
			CreatedAt: now.Add(-1 * time.Hour),
			ExpiresAt: expiresAt,
		},
	}
	var expectedTotal uint64 = 2

	mockRepo.EXPECT().List(ctx, params).Return(expectedSessions, expectedTotal, nil)

	handler := query.NewListSessionsHandler(mockRepo)

	sessions, total, err := handler.Handle(ctx, params)

	assert.NoError(t, err)
	assert.Equal(t, expectedSessions, sessions)
	assert.Equal(t, expectedTotal, total)
	assert.Len(t, sessions, 2)
}

func TestListSessionsHandler_Handle_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockSessionQueryRepository(ctrl)

	params := session_domain.ListSessionsParams{
		SortBy:   session_domain.ListSessionsSortByNone,
		SortKind: domain.SortAsc,
		Pagination: domain.Pagination{
			Number: 1,
			Size:   10,
		},
	}

	var expectedTotal uint64 = 0
	emptySessions := []session_domain.ListSessionsDTO{}

	mockRepo.EXPECT().List(ctx, params).Return(emptySessions, expectedTotal, nil)

	handler := query.NewListSessionsHandler(mockRepo)

	sessions, total, err := handler.Handle(ctx, params)

	assert.NoError(t, err)
	assert.Empty(t, sessions)
	assert.Equal(t, expectedTotal, total)
}

func TestListSessionsHandler_Handle_WithPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockSessionQueryRepository(ctrl)

	params := session_domain.ListSessionsParams{
		SortBy:   session_domain.ListSessionsSortByCreatedAt,
		SortKind: domain.SortAsc,
		Pagination: domain.Pagination{
			Number: 3,
			Size:   20,
		},
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	expectedSessions := []session_domain.ListSessionsDTO{
		{
			UserAgent: "Safari/14.0",
			IPAddress: "10.0.0.1",
			CreatedAt: now.Add(-5 * time.Hour),
			ExpiresAt: expiresAt,
		},
	}
	var expectedTotal uint64 = 45 // Total across all pages

	mockRepo.EXPECT().List(ctx, params).Return(expectedSessions, expectedTotal, nil)

	handler := query.NewListSessionsHandler(mockRepo)

	sessions, total, err := handler.Handle(ctx, params)

	assert.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, expectedTotal, total)
}

func TestListSessionsHandler_Handle_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockSessionQueryRepository(ctrl)

	params := session_domain.ListSessionsParams{
		SortBy:   session_domain.ListSessionsSortByCreatedAt,
		SortKind: domain.SortDesc,
		Pagination: domain.Pagination{
			Number: 1,
			Size:   10,
		},
	}

	mockRepo.EXPECT().List(ctx, params).Return(nil, uint64(0), errors.New("database connection failed"))

	handler := query.NewListSessionsHandler(mockRepo)

	sessions, total, err := handler.Handle(ctx, params)

	assert.Error(t, err)
	assert.Equal(t, "database connection failed", err.Error())
	assert.Nil(t, sessions)
	assert.Equal(t, uint64(0), total)
}

func TestListSessionsHandler_Handle_DifferentSortOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockSessionQueryRepository(ctrl)

	t.Run("sort by created_at ascending", func(t *testing.T) {
		params := session_domain.ListSessionsParams{
			SortBy:   session_domain.ListSessionsSortByCreatedAt,
			SortKind: domain.SortAsc,
			Pagination: domain.Pagination{
				Number: 1,
				Size:   10,
			},
		}

		mockRepo.EXPECT().List(ctx, params).Return([]session_domain.ListSessionsDTO{}, uint64(0), nil)

		handler := query.NewListSessionsHandler(mockRepo)
		sessions, total, err := handler.Handle(ctx, params)

		assert.NoError(t, err)
		assert.Empty(t, sessions)
		assert.Equal(t, uint64(0), total)
	})

	t.Run("sort by expires_at descending", func(t *testing.T) {
		params := session_domain.ListSessionsParams{
			SortBy:   session_domain.ListSessionsSortByExpiresAt,
			SortKind: domain.SortDesc,
			Pagination: domain.Pagination{
				Number: 1,
				Size:   10,
			},
		}

		mockRepo.EXPECT().List(ctx, params).Return([]session_domain.ListSessionsDTO{}, uint64(0), nil)

		handler := query.NewListSessionsHandler(mockRepo)
		sessions, total, err := handler.Handle(ctx, params)

		assert.NoError(t, err)
		assert.Empty(t, sessions)
		assert.Equal(t, uint64(0), total)
	})

	t.Run("sort by user_agent ascending", func(t *testing.T) {
		params := session_domain.ListSessionsParams{
			SortBy:   session_domain.ListSessionsSortByUserAgent,
			SortKind: domain.SortAsc,
			Pagination: domain.Pagination{
				Number: 1,
				Size:   10,
			},
		}

		mockRepo.EXPECT().List(ctx, params).Return([]session_domain.ListSessionsDTO{}, uint64(0), nil)

		handler := query.NewListSessionsHandler(mockRepo)
		sessions, total, err := handler.Handle(ctx, params)

		assert.NoError(t, err)
		assert.Empty(t, sessions)
		assert.Equal(t, uint64(0), total)
	})

	t.Run("sort by ip_address descending", func(t *testing.T) {
		params := session_domain.ListSessionsParams{
			SortBy:   session_domain.ListSessionsSortByIPAddress,
			SortKind: domain.SortDesc,
			Pagination: domain.Pagination{
				Number: 1,
				Size:   10,
			},
		}

		mockRepo.EXPECT().List(ctx, params).Return([]session_domain.ListSessionsDTO{}, uint64(0), nil)

		handler := query.NewListSessionsHandler(mockRepo)
		sessions, total, err := handler.Handle(ctx, params)

		assert.NoError(t, err)
		assert.Empty(t, sessions)
		assert.Equal(t, uint64(0), total)
	})
}

func TestListSessionsHandler_Handle_WithVariousUserAgents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockSessionQueryRepository(ctrl)

	params := session_domain.ListSessionsParams{
		SortBy:   session_domain.ListSessionsSortByNone,
		SortKind: domain.SortAsc,
		Pagination: domain.Pagination{
			Number: 1,
			Size:   10,
		},
	}

	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	expectedSessions := []session_domain.ListSessionsDTO{
		{
			UserAgent: "PostmanRuntime/7.26.8",
			IPAddress: "192.168.1.1",
			CreatedAt: now,
			ExpiresAt: expiresAt,
		},
		{
			UserAgent: "curl/7.68.0",
			IPAddress: "10.0.0.5",
			CreatedAt: now,
			ExpiresAt: expiresAt,
		},
		{
			UserAgent: "python-requests/2.25.1",
			IPAddress: "172.16.0.1",
			CreatedAt: now,
			ExpiresAt: expiresAt,
		},
	}
	var expectedTotal uint64 = 3

	mockRepo.EXPECT().List(ctx, params).Return(expectedSessions, expectedTotal, nil)

	handler := query.NewListSessionsHandler(mockRepo)

	sessions, total, err := handler.Handle(ctx, params)

	assert.NoError(t, err)
	assert.Len(t, sessions, 3)
	assert.Equal(t, "PostmanRuntime/7.26.8", sessions[0].UserAgent)
	assert.Equal(t, "curl/7.68.0", sessions[1].UserAgent)
	assert.Equal(t, "python-requests/2.25.1", sessions[2].UserAgent)
	assert.Equal(t, expectedTotal, total)
}
