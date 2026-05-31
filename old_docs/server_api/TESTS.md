# Unit Tests

This directory contains comprehensive unit tests for the `server_api/database` and `server_api/service` packages.

## Test Coverage

- **database package**: 37.1% coverage
- **service package**: 77.8% coverage
- **Overall**: 48.6% coverage

## Running Tests

### Run all tests
```bash
go test ./server_api/database/... ./server_api/service/... -v
```

### Run tests with coverage
```bash
go test ./server_api/database/... ./server_api/service/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Run specific test
```bash
go test ./server_api/database/... -run TestCachedUserRepository -v
go test ./server_api/service/... -run TestAuthServiceImpl_Login -v
```

## Test Structure

### Database Package Tests (`repository_test.go`)

Tests for repository layer implementations:

1. **CachedUserRepository Tests**
   - `TestCachedUserRepository_CreateUser`: User creation with cache invalidation
   - `TestCachedUserRepository_FindByID`: User retrieval by ID with caching
   - `TestCachedUserRepository_FindByUsername`: User retrieval by username with caching
   - `TestCachedUserRepository_ExistsByID`: User existence check by ID
   - `TestCachedUserRepository_GetUserProfile`: User profile retrieval with caching
   - `TestCachedUserRepository_DeleteUser`: User deletion with cache cleanup

2. **RedisTokenRepository Tests**
   - `TestRedisTokenRepository_StoreVerifyToken`: Verification token storage
   - `TestRedisTokenRepository_VerifyAndConsumeToken`: Token verification and consumption

3. **RedisRateLimitRepository Tests**
   - `TestRedisRateLimitRepository_TrackLoginAttempt`: Login attempt tracking
   - `TestRedisRateLimitRepository_CheckAccountLocked`: Account lock status checking
   - `TestRedisRateLimitRepository_TrackIPVisit`: IP visit tracking
   - `TestRedisRateLimitRepository_BlockIP`: IP blocking functionality

### Service Package Tests (`service_test.go`)

Tests for service layer business logic:

1. **AuthService Tests**
   - `TestAuthServiceImpl_Login`: User authentication with various scenarios
     - Successful login
     - Invalid input
     - Invalid verification token
     - Account locked
     - User not found
     - Invalid password
   
   - `TestAuthServiceImpl_Register`: User registration with validation
     - Successful registration
     - Invalid input
     - Username too long
     - Username already exists
     - Duplicate key error

2. **UserService Tests**
   - `TestUserServiceImpl_GetUserProfile`: Profile retrieval
   - `TestUserServiceImpl_UpdateUserProfile`: Profile updates
   - `TestUserServiceImpl_DeleteUser`: Account deletion with password verification
   - `TestUserServiceImpl_CheckAccountLocked`: Account lock checking
   - `TestUserServiceImpl_CheckUserExists`: User existence checking

3. **TokenService Tests**
   - `TestTokenServiceImpl_GenerateJWT`: JWT token generation
   - `TestTokenServiceImpl_GenerateVerificationToken`: Verification token generation
   - `TestTokenServiceImpl_StoreVerificationToken`: Token storage
   - `TestTokenServiceImpl_ValidateAndConsumeToken`: Token validation and consumption

## Testing Approach

### Mock-Based Testing

All tests use mock objects to isolate the code under test:
- `MockUserRepository`: Mocks database user operations
- `MockRateLimitRepository`: Mocks rate limiting operations
- `MockTokenRepository`: Mocks token operations
- `MockCacheService`: Mocks caching operations
- `MockTokenService`: Mocks token service operations

### Test Patterns

1. **Table-Driven Tests**: Most tests use table-driven approach for better maintainability
2. **Fresh Mocks per Subtest**: Each subtest creates fresh mock instances to avoid state pollution
3. **Comprehensive Scenarios**: Tests cover success cases, error cases, and edge cases
4. **Assertion Validation**: All mock expectations are verified using `AssertExpectations(t)`

### Key Testing Principles

1. **Isolation**: Each test is independent and doesn't rely on other tests
2. **Deterministic**: Tests produce consistent results regardless of execution order
3. **Fast**: Tests complete quickly without external dependencies
4. **Clear Intent**: Test names clearly describe what is being tested

## Adding New Tests

When adding new tests:

1. Follow the table-driven test pattern
2. Create fresh mocks for each subtest if needed
3. Cover both success and failure scenarios
4. Use descriptive test case names
5. Assert all mock expectations
6. Handle wrapped errors properly using `errors.Is()` or `Contains()`

## Example Test Structure

```go
func TestExample(t *testing.T) {
    tests := []struct {
        name          string
        // input fields
        expectedError error
        setupMocks    func(*MockType1, *MockType2)
    }{
        {
            name: "successful operation",
            // input values
            expectedError: nil,
            setupMocks: func(mock1 *MockType1, mock2 *MockType2) {
                mock1.On("Method", args).Return(result, nil)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock1 := new(MockType1)
            mock2 := new(MockType2)
            
            tt.setupMocks(mock1, mock2)
            
            // Call method under test
            err := methodUnderTest()
            
            // Assertions
            if tt.expectedError != nil {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
            
            mock1.AssertExpectations(t)
            mock2.AssertExpectations(t)
        })
    }
}
```

## Best Practices

1. **Use bcrypt for password testing**: Always generate real hashes using `bcrypt.GenerateFromPassword()`
2. **Handle wrapped errors**: Use `assert.Contains()` for wrapped errors instead of direct equality
3. **Clean mock expectations**: Ensure all expected calls are set up before test execution
4. **Test error paths**: Don't just test happy paths; test all error conditions
5. **Document test intent**: Use clear test and subtest names that explain the scenario
