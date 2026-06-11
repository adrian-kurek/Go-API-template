package mocks

import (
	"context"

	authDTO "github.com/slodkiadrianek/Go-API-template/internal/auth/DTO"
	"github.com/slodkiadrianek/Go-API-template/internal/user/model"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user authDTO.CreateUser, hashedPassword []byte) error {
	args := m.Called(ctx, user, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(model.User), args.Error(1)
}
