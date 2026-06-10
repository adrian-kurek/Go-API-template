package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Set(ctx context.Context, key, data string, ttl time.Duration) error {
	args := m.Called(ctx, key, data, ttl)
	return args.Error(0)
}

func (m *MockCacheService) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) Close() error {
	args := m.Called()
	return args.Error(0)
}
