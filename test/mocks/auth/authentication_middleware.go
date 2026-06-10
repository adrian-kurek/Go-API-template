package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"

	userModel "github.com/slodkiadrianek/Go-API-template/internal/user/model"
)

type MockAuthorizationMiddleware struct {
	mock.Mock
}

func (m *MockAuthorizationMiddleware) GenerateRefreshToken() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockAuthorizationMiddleware) HashToken(token []byte) string {
	args := m.Called(token)
	return args.Get(0).(string)
}

func (m *MockAuthorizationMiddleware) GenerateAccessToken(user userModel.User) (string, error) {
	args := m.Called(user)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockAuthorizationMiddleware) VerifyToken(r *http.Request) (*http.Request, error) {
	args := m.Called(r)
	return args.Get(0).(*http.Request), args.Error(1)
}

func (m *MockAuthorizationMiddleware) BlacklistUser(r *http.Request) error {
	args := m.Called(r)
	return args.Error(0)
}
