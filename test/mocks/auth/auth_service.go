package mocks

import (
	"context"

	authDTO "github.com/slodkiadrianek/Go-API-template/internal/auth/DTO"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, user authDTO.CreateUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockAuthService) Login(ctx context.Context, loginData authDTO.LoginUser, ipAddress, deviceInfo string) (string, []byte, error) {
	args := m.Called(ctx, loginData, ipAddress, deviceInfo)
	return args.Get(0).(string), args.Get(1).([]byte), args.Error(2)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken []byte) (string, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockAuthService) LogoutUser(ctx context.Context, refreshToken []byte) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) LogoutUserFromAllDevices(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (as *MockAuthService) ActivateAccount(ctx context.Context, userID int) error {
	args := as.Called(ctx, userID)
	return args.Error(0)
}
