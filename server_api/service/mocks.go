package service

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// MockTokenService is a mock implementation of TokenService for testing
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateJWT(userID int, expiration time.Duration) (string, error) {
	args := m.Called(userID, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) GenerateVerificationToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) StoreVerificationToken(token string, ttl time.Duration) error {
	args := m.Called(token, ttl)
	return args.Error(0)
}

func (m *MockTokenService) ValidateAndConsumeToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}
