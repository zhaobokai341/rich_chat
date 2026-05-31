# Quick Reference - Client Testing

## Running Tests

### All Tests
```bash
cd client
go test -v ./...
```

### With Race Detector
```bash
go test -race -v ./...
```

### Specific Package
```bash
go test -v -run TestAuthService
```

### Specific Test Case
```bash
go test -v -run "TestAuthService_Login/successful_login"
```

### Coverage Report
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Test Files Overview

| File | Tests | Purpose |
|------|-------|---------|
| `mocks.go` | - | Mock implementations for all interfaces |
| `auth_service_test.go` | 16 | Login, register, logout, auth state |
| `user_service_test.go` | 12 | Profile CRUD, account deletion |
| `config_manager_test.go` | 10 | Config file operations |
| `token_extractor_test.go` | 5 | JWT token parsing |
| `api_client_test.go` | 5 | API health checks |
| **Total** | **48** | **All passing ✅** |

## Key Test Patterns

### Table-Driven Test Structure
```go
tests := []struct {
    name          string
    mockSetup     func(*MockAPIClient, *MockConfigManager)
    expectedError bool
}{
    // test cases
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Arrange
        mockAPI := &MockAPIClient{}
        mockConfig := NewMockConfigManager()
        
        // Setup mocks
        if tt.mockSetup != nil {
            tt.mockSetup(mockAPI, mockConfig)
        }
        
        // Act
        service := NewAuthService(mockAPI, mockConfig, extractor)
        err := service.Login("user", "pass")
        
        // Assert
        if tt.expectedError {
            assert.Error(t, err)
        } else {
            assert.NoError(t, err)
        }
    })
}
```

### Mock Setup Example
```go
mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
    api.LoginFunc = func(username, password, verifyToken string) (*AuthResponse, error) {
        return CreateSuccessAuthResponse(1, "mock-token"), nil
    }
}
```

## Common Test Scenarios

### Testing Success Path
```go
{
    name: "successful operation",
    mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
        api.SomeFunc = func(...) (Result, error) {
            return expectedResult, nil
        }
    },
    expectedError: false,
}
```

### Testing Error Path
```go
{
    name: "operation fails",
    mockSetup: func(api *MockAPIClient, config *MockConfigManager) {
        api.SomeFunc = func(...) (Result, error) {
            return nil, errors.New("expected error")
        }
    },
    expectedError: true,
}
```

### Testing State Changes
```go
// After operation, verify state changed
if _, exists := mockConfig.GetToken(); !exists {
    t.Errorf("Token should be saved")
}
```

## Mock Objects Available

### MockAPIClient
```go
mockAPI := &MockAPIClient{
    GetVerifyTokenFunc:   func() (string, error) { ... },
    LoginFunc:            func(...) (*AuthResponse, error) { ... },
    RegisterFunc:         func(...) (*AuthResponse, error) { ... },
    DeleteUserFunc:       func(...) error { ... },
    GetUserProfileFunc:   func(...) (*UserInfoResponse, error) { ... },
    UpdateUserProfileFunc: func(...) error { ... },
    CheckServerHealthFunc: func() (bool, error) { ... },
}
```

### MockConfigManager
```go
mockConfig := NewMockConfigManager()
// Or customize behavior
mockConfig := &MockConfigManager{
    GetTokenFunc: func() (string, bool) { ... },
    SaveConfigFunc: func(data map[string]interface{}) error { ... },
}
```

### MockTokenExtractor
```go
mockExtractor := &MockTokenExtractor{
    ExtractUserIDFunc: func(token string) (string, error) { ... },
}
```

## Helper Functions

```go
// Create success response
CreateSuccessAuthResponse(userID int, token string) *AuthResponse

// Create error response
CreateErrorAuthResponse(message string) (*AuthResponse, error)

// Create success user info
CreateSuccessUserInfoResponse() *UserInfoResponse
```

## Debugging Tests

### Verbose Output
```bash
go test -v -run TestName
```

### Show Coverage
```bash
go test -cover -v
```

### Run Specific Subtest
```bash
go test -v -run "TestName/subtest_name"
```

### Count Mode (disable cache)
```bash
go test -count=1 -v
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.26'
      - name: Run tests
        run: |
          cd client
          go test -v -race -coverprofile=coverage.out ./...
      - name: Check coverage
        run: |
          cd client
          go tool cover -func=coverage.out | grep total
```

## Best Practices

1. ✅ Always use table-driven tests
2. ✅ Mock all external dependencies
3. ✅ Test both success and error paths
4. ✅ Verify state changes after operations
5. ✅ Use descriptive test names
6. ✅ Keep tests fast (< 1ms each)
7. ✅ Ensure test isolation
8. ✅ Use subtests for organization

## Troubleshooting

### Test Fails with "nil pointer"
- Check that mocks are properly initialized
- Verify all required mock functions are set

### Test Runs Slow
- Ensure no real I/O operations
- Check for unnecessary sleeps or waits
- Use `-count=1` to disable cache

### Coverage Too Low
- Add more test cases for edge conditions
- Test error handling paths
- Consider integration tests for UI layer

---

**Remember**: Good tests are documentation, safety net, and design tool all in one! 🚀
