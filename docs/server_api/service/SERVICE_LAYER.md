# Service Layer Refactoring

## Overview

The service layer has been created to separate business logic from HTTP handlers and data access layers. This follows the **Repository-Service-Handler** architecture pattern.

## Architecture

```
┌─────────────────────────────────────┐
│      HTTP Handlers (web_server)     │
│   - Parse requests                  │
│   - Call services                   │
│   - Format responses                │
└──────────────┬──────────────────────┘
               │ depends on interfaces
               ▼
┌─────────────────────────────────────┐
│       Service Layer (NEW!)          │
│   - AuthService                     │
│   - UserService                     │
│   - TokenService                    │
│                                     │
│ Business Logic:                     │
│   - Password verification           │
│   - Token generation                │
│   - Validation                      │
│   - Rate limiting checks            │
└──────────────┬──────────────────────┘
               │ depends on interfaces
               ▼
┌─────────────────────────────────────┐
│      Repository Layer               │
│   - UserRepository                  │
│   - RateLimitRepository             │
│   - TokenRepository                 │
└─────────────────────────────────────┘
```

## Key Components

### 1. Service Interfaces (`interfaces.go`)

Defines contracts for all business operations:

```go
type AuthService interface {
    Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
    Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
    GenerateToken(userID int) (string, error)
    ValidateVerifyToken(token string) error
}

type UserService interface {
    GetUserProfile(ctx context.Context, userID int) (*UserInfo, error)
    UpdateUserProfile(ctx context.Context, req *UserProfileUpdateRequest) error
    DeleteUser(ctx context.Context, req *DeleteUserRequest) error
}

type TokenService interface {
    GenerateJWT(userID int, expiration time.Duration) (string, error)
    GenerateVerificationToken() (string, error)
    ValidateAndConsumeToken(token string) error
}
```

### 2. AuthService (`auth_service.go`)

Handles authentication business logic:

**Responsibilities:**
- User login with password verification
- User registration with validation
- Account lockout checking
- Login attempt tracking
- JWT token generation after successful auth

**Example:**
```go
// Login flow
req := &LoginRequest{
    Username:    "john",
    Password:    "secret",
    VerifyToken: "abc123",
}

resp, err := authService.Login(ctx, req)
if err != nil {
    // Handle errors: ErrInvalidPassword, ErrAccountLocked, etc.
}
// resp.UserID, resp.UserToken available
```

### 3. UserService (`user_service.go`)

Handles user management business logic:

**Responsibilities:**
- User profile retrieval
- Profile updates with validation
- User deletion with password verification
- Account status checking

**Example:**
```go
// Get user profile
profile, err := userService.GetUserProfile(ctx, userID)
if err != nil {
    // Handle errors: ErrUserNotFound, etc.
}

// Update profile
req := &UserProfileUpdateRequest{
    UserID: userID,
    Key:    "nickname",
    Value:  "John Doe",
}
err := userService.UpdateUserProfile(ctx, req)
```

### 4. TokenService (`token_service.go`)

Handles token operations:

**Responsibilities:**
- JWT token generation with claims
- Verification token generation
- Token validation and consumption
- Token storage in Redis

**Example:**
```go
// Generate JWT
token, err := tokenService.GenerateJWT(userID, 24*time.Hour)

// Generate verification token
verifyToken, err := tokenService.GenerateVerificationToken()

// Store and validate
err = tokenService.StoreVerificationToken(verifyToken, 5*time.Minute)
err = tokenService.ValidateAndConsumeToken(verifyToken)
```

### 5. Service Factory (`factory.go`)

Creates and wires all services together:

```go
services := NewServices(dbService, ServiceConfig{
    JWTSecret:         "your-secret",
    JWTExpiration:     30 * 24 * time.Hour,
    MaxUsernameLength: 50,
    VerifyTokenTTL:    5 * time.Minute,
})

// Access services
authService := services.AuthService
userService := services.UserService
tokenService := services.TokenService
```

## Error Handling

Standardized error types for consistent handling:

```go
var (
    ErrUserNotFound          = errors.New("user not found")
    ErrInvalidPassword       = errors.New("invalid password")
    ErrAccountLocked         = errors.New("account is locked")
    ErrInvalidToken          = errors.New("invalid or expired token")
    ErrUsernameAlreadyExists = errors.New("username already exists")
    ErrInvalidInput          = errors.New("invalid input")
)
```

**Usage in handlers:**
```go
resp, err := authService.Login(ctx, req)
if err != nil {
    switch err {
    case ErrInvalidPassword:
        c.JSON(401, gin.H{"message": "Invalid credentials"})
    case ErrAccountLocked:
        c.JSON(429, gin.H{"message": "Account locked"})
    default:
        c.JSON(500, gin.H{"message": "Internal error"})
    }
    return
}
```

## Testing Strategy

### Unit Tests with Mocks

```go
func TestAuthService_Login_Success(t *testing.T) {
    // Create mocks
    mockUserRepo := new(MockUserRepository)
    mockRateLimitRepo := new(MockRateLimitRepository)
    mockTokenRepo := new(MockTokenRepository)
    
    // Setup expectations
    mockTokenRepo.On("VerifyAndConsumeToken", "token").Return(true, nil)
    mockUserRepo.On("FindByUsername", "user").Return(&User{...}, nil)
    
    // Create service
    authService := NewAuthService(mockUserRepo, mockRateLimitRepo, ...)
    
    // Test
    resp, err := authService.Login(ctx, &LoginRequest{...})
    
    // Assert
    assert.NoError(t, err)
    mockUserRepo.AssertExpectations(t)
}
```

### Integration Tests

```go
func TestAuthService_Integration(t *testing.T) {
    // Use real database and Redis
    dbService := setupTestDB()
    services := NewServices(dbService, testConfig)
    
    // Test with real dependencies
    resp, err := services.AuthService.Login(ctx, req)
    
    assert.NoError(t, err)
}
```

## Migration Guide

### Before (Business Logic in Handlers)

```go
// Handler had everything mixed together
func (api *WebServerApi) Login(c *gin.Context) {
    username := c.PostForm("username")
    password := c.PostForm("password")
    
    // Validation
    if username == "" || password == "" {
        // ...
    }
    
    // Check account lock
    if user_database.IsAccountLocked(username) {
        // ...
    }
    
    // Find user
    user_id, pass := user_database.SelectUserIsExist(username, password)
    if !pass {
        user_database.TrackLoginAttempt(username, false)
        // ...
    }
    
    // Track success
    user_database.TrackLoginAttempt(username, true)
    
    // Generate token
    jwt_token, _ := generateToken(user_id, JWT_EXPIRE_TIME)
    
    c.JSON(200, gin.H{"token": jwt_token})
}
```

### After (Clean Separation)

```go
// Handler only handles HTTP concerns
func (api *WebServerApi) Login(c *gin.Context) {
    // Parse request
    req := &service.LoginRequest{
        Username:    c.PostForm("username"),
        Password:    c.PostForm("password"),
        VerifyToken: c.PostForm("verify_token"),
    }
    
    // Call service
    resp, err := api.authService.Login(c.Request.Context(), req)
    if err != nil {
        // Handle standardized errors
        handleServiceError(c, err)
        return
    }
    
    // Return response
    c.JSON(200, gin.H{
        "user_token": resp.UserToken,
        "user_id":    resp.UserID,
    })
}
```

## Benefits

### Before Service Layer
- ❌ Business logic scattered across handlers
- ❌ Hard to test (need full HTTP stack)
- ❌ Duplicate code in multiple handlers
- ❌ Tight coupling to Gin framework
- ❌ No clear separation of concerns

### After Service Layer
- ✅ Business logic centralized in services
- ✅ Easy to unit test with mocks
- ✅ DRY principle - no duplication
- ✅ Framework agnostic (can swap HTTP framework)
- ✅ Clear separation: Handler → Service → Repository

## Usage Examples

See `service_test.go` for comprehensive examples:
- Login success/failure scenarios
- Registration with validation
- Token generation and validation
- User profile operations
- Error handling patterns

## Next Steps

1. **Migrate Handlers**: Update existing handlers to use services
2. **Add More Services**: Create services for other domains (chat, messages, etc.)
3. **Add Middleware**: Create service-level middleware (logging, metrics)
4. **Add Caching**: Implement service-level caching strategies
5. **Add Validation**: Centralize input validation in services

---

**Status**: ✅ Service layer created and ready for integration!
