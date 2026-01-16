package command_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brunoibarbosa/url-shortener/internal/app/auth/command"
	"github.com/brunoibarbosa/url-shortener/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRedirectGoogleHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	expectedState := "random-state-123"
	expectedURL := "https://accounts.google.com/o/oauth2/auth?state=random-state-123"

	mockProvider := mocks.NewMockOAuthProvider(ctrl)
	mockStateService := mocks.NewMockStateService(ctrl)

	mockStateService.EXPECT().GenerateState(ctx).Return(expectedState, nil)
	mockProvider.EXPECT().GetAuthURL(expectedState).Return(expectedURL)

	handler := command.NewRedirectGoogleHandler(mockProvider, mockStateService)

	url, err := handler.Handle(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, url)
}

func TestRedirectGoogleHandler_Handle_StateGenerationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	expectedError := errors.New("state generation failed")

	mockProvider := mocks.NewMockOAuthProvider(ctrl)
	mockStateService := mocks.NewMockStateService(ctrl)

	mockStateService.EXPECT().GenerateState(ctx).Return("", expectedError)

	handler := command.NewRedirectGoogleHandler(mockProvider, mockStateService)

	url, err := handler.Handle(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, url)
}
