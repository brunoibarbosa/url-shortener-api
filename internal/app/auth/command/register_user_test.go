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

func TestRegisterUserHandler_Handle_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	email := "newuser@example.com"
	password := "password123"
	name := "John Doe"
	passwordHash := "hashed_password"

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockPasswordEncrypter := mocks.NewMockUserPasswordEncrypter(ctrl)

	mockPasswordEncrypter.EXPECT().HashPassword(password).Return(passwordHash, nil)
	mockUserRepo.EXPECT().Exists(ctx, email).Return(false, nil)

	mockTx.EXPECT().WithinTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	)

	mockUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, u *user_domain.User) error {
			u.ID = uuid.New()
			u.CreatedAt = time.Now()
			return nil
		},
	)
	mockProviderRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockProfileRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	handler := command.NewRegisterUserHandler(
		mockTx,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockPasswordEncrypter,
	)

	cmd := command.RegisterUserCommand{
		Email:    email,
		Password: password,
		Name:     name,
	}

	result, err := handler.Handle(ctx, cmd)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, name, result.Profile.Name)
	assert.NotEqual(t, uuid.Nil, result.ID)
}

func TestRegisterUserHandler_Handle_EmailAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	email := "existing@example.com"
	password := "password123"
	name := "John Doe"
	passwordHash := "hashed_password"

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockPasswordEncrypter := mocks.NewMockUserPasswordEncrypter(ctrl)

	mockPasswordEncrypter.EXPECT().HashPassword(password).Return(passwordHash, nil)
	mockUserRepo.EXPECT().Exists(ctx, email).Return(true, nil)

	handler := command.NewRegisterUserHandler(
		mockTx,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockPasswordEncrypter,
	)

	cmd := command.RegisterUserCommand{
		Email:    email,
		Password: password,
		Name:     name,
	}

	result, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, user_domain.ErrEmailAlreadyExists, err)
	assert.Nil(t, result)
}

func TestRegisterUserHandler_Handle_HashPasswordError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	email := "user@example.com"
	password := "password123"
	name := "John Doe"
	expectedError := errors.New("hash error")

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockPasswordEncrypter := mocks.NewMockUserPasswordEncrypter(ctrl)

	mockPasswordEncrypter.EXPECT().HashPassword(password).Return("", expectedError)

	handler := command.NewRegisterUserHandler(
		mockTx,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockPasswordEncrypter,
	)

	cmd := command.RegisterUserCommand{
		Email:    email,
		Password: password,
		Name:     name,
	}

	result, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestRegisterUserHandler_Handle_TransactionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	email := "user@example.com"
	password := "password123"
	name := "John Doe"
	passwordHash := "hashed_password"
	expectedError := errors.New("transaction error")

	mockTx := mocks.NewMockTransactionManager(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockProviderRepo := mocks.NewMockUserProviderRepository(ctrl)
	mockProfileRepo := mocks.NewMockUserProfileRepository(ctrl)
	mockPasswordEncrypter := mocks.NewMockUserPasswordEncrypter(ctrl)

	mockPasswordEncrypter.EXPECT().HashPassword(password).Return(passwordHash, nil)
	mockUserRepo.EXPECT().Exists(ctx, email).Return(false, nil)
	mockTx.EXPECT().WithinTransaction(ctx, gomock.Any()).Return(expectedError)

	handler := command.NewRegisterUserHandler(
		mockTx,
		mockUserRepo,
		mockProviderRepo,
		mockProfileRepo,
		mockPasswordEncrypter,
	)

	cmd := command.RegisterUserCommand{
		Email:    email,
		Password: password,
		Name:     name,
	}

	result, err := handler.Handle(ctx, cmd)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}
