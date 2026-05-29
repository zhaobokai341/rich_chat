# Database Refactoring - Usage Examples

## Quick Start

### 1. Initialize the New Database Service

```go
package main

import (
    "rich_chat/server_api/database"
)

func main() {
    // Load configuration
    cfg := database.Config{
        DB_HOST: "localhost",
        DB_PORT: 5432,
        DB_USER: "postgres",
        DB_PASS: "password",
        DB_NAME: "rich_chat",
        DB_SSL: "disable",
        MAXOPENCONNS: 25,
        MAXIDLECONNS: 10,
        CONNMAXLIFETIME: 5 * time.Minute,
        CONNMAXIDLETIME: 5 * time.Minute,
        MAX_LOGIN_ATTEMPTS: 5,
        LOCKOUT_DURATION: 15 * time.Minute,
        IP_LIMIT_TIME: 10 * time.Minute,
        IP_LIMIT_VISIT_TIMES: 1000,
        IP_LIMIT_LOCKOUT_DURATION: 10 * time.Minute,
    }
    
    // Initialize Redis manager (existing code)
    redisManager := initializeRedis()
    
    // Create new database service
    dbService, err := database.InitializeDatabaseService(cfg, redisManager)
    if err != nil {
        log.Fatal(err)
    }
    defer dbService.GetDB().Close()
    
    // Use the service...
}
```

### 2. Using UserRepository

```go
// Get user repository
userRepo := dbService.GetUserRepository()

// Create a new user
userID, err := userRepo.CreateUser("john_doe", "$2a$10$hashed_password")
if err != nil {
    log.Error("Failed to create user:", err)
}

// Find user by ID
user, err := userRepo.FindByID(userID)
if err != nil {
    log.Error("User not found:", err)
}

// Find user by username
user, err = userRepo.FindByUsername("john_doe")
if err != nil {
    log.Error("User not found:", err)
}

// Check if user exists
exists, err := userRepo.ExistsByUsername("john_doe")
if err != nil {
    log.Error("Error checking user:", err)
}

// Get user profile
profile, err := userRepo.GetUserProfile(userID)
if err != nil {
    log.Error("Failed to get profile:", err)
}
fmt.Printf("Username: %s, Nickname: %s\n", profile.Username, profile.Nickname)

// Update user profile
err = userRepo.UpdateProfile(userID, "nickname", "John")
if err != nil {
    log.Error("Failed to update profile:", err)
}

// Delete user
err = userRepo.DeleteUser(userID)
if err != nil {
    log.Error("Failed to delete user:", err)
}
```

### 3. Using RateLimitRepository

```go
// Get rate limit repository
rateLimitRepo := dbService.GetRateLimitRepository()

// Track login attempt
err := rateLimitRepo.TrackLoginAttempt("john_doe", false) // failed login
if err != nil {
    log.Error("Failed to track login attempt:", err)
}

// Check if account is locked
locked, err := rateLimitRepo.CheckAccountLocked("john_doe")
if err != nil {
    log.Error("Failed to check lock status:", err)
}
if locked {
    fmt.Println("Account is locked due to too many failed attempts")
}

// Track IP visit
count, err := rateLimitRepo.TrackIPVisit("192.168.1.1")
if err != nil {
    log.Error("Failed to track IP:", err)
}
fmt.Printf("IP visited %d times\n", count)

// Block an IP
err = rateLimitRepo.BlockIP("192.168.1.1", "Suspicious activity", 10*time.Minute)
if err != nil {
    log.Error("Failed to block IP:", err)
}

// Check if IP is blocked
blocked, err := rateLimitRepo.CheckIPBlocked("192.168.1.1")
if err != nil {
    log.Error("Failed to check IP block status:", err)
}
```

### 4. Using TokenRepository

```go
// Get token repository
tokenRepo := dbService.GetTokenRepository()

// Store a verification token
token := generateRandomToken() // your token generation logic
err := tokenRepo.StoreVerifyToken(token, 5*time.Minute)
if err != nil {
    log.Error("Failed to store token:", err)
}

// Verify and consume token
valid, err := tokenRepo.VerifyAndConsumeToken(token)
if err != nil {
    log.Error("Failed to verify token:", err)
}
if !valid {
    fmt.Println("Invalid or expired token")
} else {
    fmt.Println("Token is valid")
}
```

## Testing Examples

### Unit Test with Mocks

```go
package database_test

import (
    "testing"
    "rich_chat/server_api/database"
    "github.com/stretchr/testify/mock"
)

func TestUserService_Login(t *testing.T) {
    // Create mocks
    mockUserRepo := new(MockUserRepository)
    mockCache := new(MockCacheService)
    
    // Setup expectations
    mockUserRepo.On("FindByUsername", "john_doe").Return(&database.User{
        ID: 1,
        Username: "john_doe",
        PasswordHash: "$2a$10$...",
    }, nil)
    
    // Create cached repository
    cachedRepo := database.NewCachedUserRepository(mockUserRepo, mockCache)
    
    // Test
    user, err := cachedRepo.FindByUsername("john_doe")
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, 1, user.ID)
    mockUserRepo.AssertExpectations(t)
}
```

### Integration Test with Real Database

```go
func TestPostgresUserRepository_Integration(t *testing.T) {
    // Skip in short mode
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Setup test database connection
    db := setupTestDB(t)
    defer db.Close()
    
    // Create repository
    repo := database.NewPostgresUserRepository(db)
    
    // Test user creation
    userID, err := repo.CreateUser("testuser", "$2a$10$hashed")
    assert.NoError(t, err)
    assert.Greater(t, userID, 0)
    
    // Test user retrieval
    user, err := repo.FindByID(userID)
    assert.NoError(t, err)
    assert.Equal(t, "testuser", user.Username)
    
    // Cleanup
    repo.DeleteUser(userID)
}
```

## Migration from Old API

### Before (Old API)

```go
// Old way - using UserDatabase directly
user_database := database.DatabaseInit(cfg, redisManager)

// Check username
if user_database.CheckUsernameIsExist(username) {
    // ...
}

// Login
user_id, pass := user_database.SelectUserIsExist(username, password)

// Get profile
user_info, err := user_database.GetUserProfile(user_id)
```

### After (New API)

```go
// New way - using DatabaseService
dbService := database.InitializeDatabaseService(cfg, redisManager)
userRepo := dbService.GetUserRepository()

// Check username
exists, _ := userRepo.ExistsByUsername(username)
if exists {
    // ...
}

// Login (note: password verification should be in service layer)
user, _ := userRepo.FindByUsername(username)
// Compare password hash here

// Get profile
profile, err := userRepo.GetUserProfile(user.ID)
```

## Benefits of New Architecture

1. **Testability**: Easy to mock dependencies
   ```go
   mockRepo := new(MockUserRepository)
   service := NewUserService(mockRepo)
   ```

2. **Flexibility**: Swap implementations easily
   ```go
   // Use PostgreSQL
   repo := NewPostgresUserRepository(db)
   
   // Or use in-memory for testing
   repo := NewInMemoryUserRepository()
   ```

3. **Separation of Concerns**: Each layer has clear responsibility
   - Repository: Data access
   - Cache: Caching logic
   - Service: Business logic

4. **Explicit Dependencies**: No hidden global state
   ```go
   service := NewDatabaseService(db, cache, config)
   // All dependencies are visible
   ```

## Best Practices

1. **Always use interfaces in function parameters**
   ```go
   func ProcessUser(repo UserRepository, userID int) {
       // Can accept any implementation
   }
   ```

2. **Keep repositories focused on data access**
   ```go
   // Good: Repository only handles data
   user, err := repo.FindByID(id)
   
   // Bad: Don't put business logic in repository
   if user.IsPremium() { ... } // Should be in service layer
   ```

3. **Use dependency injection**
   ```go
   // Good: Dependencies injected
   type Service struct {
       repo UserRepository
   }
   
   // Bad: Global variables
   var globalRepo UserRepository
   ```

4. **Write tests for each layer**
   - Unit tests: Mock all dependencies
   - Integration tests: Use real database/cache
   - E2E tests: Test complete flows
