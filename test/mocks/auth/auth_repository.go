package mocks

import (
	"context"

	"github.com/slodkiadrianek/Go-API-template/internal/auth/model"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) InsertRefreshToken(ctx context.Context, ipAddress, deviceInfo, refreshToken string, userID int) error {
	args := m.Called(ctx, ipAddress, deviceInfo, refreshToken, userID)
	return args.Error(0)
}

func (m *MockAuthRepository) GetRefreshTokenByTokenHash(ctx context.Context, refreshToken string) (model.TokenWithUserEmailToRefreshToken, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(model.TokenWithUserEmailToRefreshToken), args.Error(1)
}

func (m *MockAuthRepository) UpdateLastTimeUsedToken(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthRepository) RemoveTokenFromDB(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthRepository) RemoveTokensFromDBByUserID(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthRepository) ActivateAccount(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
