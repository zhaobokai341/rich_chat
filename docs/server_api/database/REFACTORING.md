# Database Layer Refactoring Guide

## Overview

The database layer has been refactored to improve testability and maintainability using:
- **Repository Pattern**: Separates data access logic from business logic
- **Dependency Injection**: Makes dependencies explicit and replaceable
- **Interface Abstraction**: Enables mocking for unit tests
- **Decorator Pattern**: Adds caching without modifying core repository logic

## Architecture

```
┌─────────────────────────────────────────┐
│         Application Layer               │
│     (Handlers, Services)                │
└──────────────┬──────────────────────────┘
               │ depends on interfaces
               ▼
┌─────────────────────────────────────────┐
│      Repository Interfaces              │
│  - UserRepository                       │
│  - RateLimitRepository                  │
│  - TokenRepository                      │
│  - CacheService                         │
└──────────────┬──────────────────────────┘
               │ implemented by
               ▼
┌─────────────────────────────────────────┐
│    Concrete Implementations             │
│  - PostgresUserRepository               │
│  - CachedUserRepository (decorator)     │
│  - RedisRateLimitRepository             │
│  - RedisTokenRepository                 │
│  - RedisCacheAdapter                    │
└─────────────────────────────────────────┘
```

## Key Components

### 1. Interfaces (`interfaces.go`)

Defines contracts for all data access operations:

```go
type UserRepository interface {
    CreateUser(username, passwordHash string) (int, error)
    FindByID(id int) (*User, error)
    FindByUsername(username string) (*User, error)
    // ... more methods
}

type CacheService interface {
    Get(key string) (string, bool)
    Set(key, value string)
    SetWithTTL(key, value string, ttlSeconds int)
    // ... more methods
}
```

### 2. PostgreSQL Implementation (`user_repository.go`)

Pure SQL operations without caching:

```go
type PostgresUserRepository struct {
    db *sqlx.DB
}

func (r *PostgresUserRepository) FindByID(id int) (*User, error) {
    var user User
    err := r.db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
    return &user, err
}
```

### 3. Caching Decorator (`cached_user_repository.go`)

Adds caching layer without modifying the base repository:

```go
type CachedUserRepository struct {
    repo  UserRepository
    cache CacheService
}

func (c *CachedUserRepository) FindByID(id int) (*User, error) {
    // Check cache first
    if cached := c.cache.Get(key); found {
        return parseFromCache(cached)
    }
    
    // Cache miss - query database
    user, err := c.repo.FindByID(id)
    if err == nil {
        c.cache.Set(key, serialize(user))
    }
    return user, err
}
```

### 4. Redis Adapter (`redis_adapter.go`)

Adapts existing RedisManager to CacheService interface:

```go
type RedisCacheAdapter struct {
    manager RedisManager
}

func (a *RedisCacheAdapter) Get(key string) (string, bool) {
    return a.manager.GetCache(key)
}
```

### 5. Database Service (`database_service.go`)

Orchestrates all repositories with dependency injection:

```go
type DatabaseService struct {
    userRepo      UserRepository
    rateLimitRepo RateLimitRepository
    tokenRepo     TokenRepository
    cache         CacheService
}

func NewDatabaseService(db *sqlx.DB, cache CacheService, config Config) *DatabaseService {
    pgRepo := NewPostgresUserRepository(db)
    cachedRepo := NewCachedUserRepository(pgRepo, cache)
    // ... initialize other repos
    
    return &DatabaseService{
        userRepo: cachedRepo,
        // ...
    }
}
```

## Testing Strategy

### Unit Tests (Mock Dependencies)

```go
func TestCachedUserRepository_FindByID_CacheHit(t *testing.T) {
    mockRepo := new(MockUserRepository)
    mockCache := new(MockCacheService)
    cachedRepo := NewCachedUserRepository(mockRepo, mockCache)
    
    // Setup cache hit
    mockCache.On("Get", "user:exists:123").Return("true", true)
    
    // Act
    exists, _ := cachedRepo.ExistsByID(123)
    
    // Assert - database should NOT be called
    assert.True(t, exists)
    mockRepo.AssertNotCalled(t, "ExistsByID")
}
```

### Integration Tests (Real Database)

```go
func TestPostgresUserRepository_CreateUser(t *testing.T) {
    // Use testcontainers or test database
    db := setupTestDB()
    defer db.Close()
    
    repo := NewPostgresUserRepository(db)
    
    userID, err := repo.CreateUser("testuser", "$2a$10$...")
    
    assert.NoError(t, err)
    assert.Greater(t, userID, 0)
}
```

## Migration Guide

### For New Code

Use the new `DatabaseService`:

```go
// Initialize
service, err := InitializeDatabaseService(cfg, redisManager)
if err != nil {
    log.Fatal(err)
}
defer service.GetDB().Close()

// Use repositories
userRepo := service.GetUserRepository()
user, err := userRepo.FindByID(123)

rateLimitRepo := service.GetRateLimitRepository()
locked, _ := rateLimitRepo.CheckAccountLocked("username")
```

### For Existing Code

The old `UserDatabase` API still works for backward compatibility. Gradually migrate to use the new service:

```go
// Old way (still works)
user_database := database.DatabaseInit(cfg, redisManager)
user_database.CheckUsernameIsExist(username)

// New way (recommended)
service := database.InitializeDatabaseService(cfg, redisManager)
userRepo := service.GetUserRepository()
exists, _ := userRepo.ExistsByUsername(username)
```

## Benefits

### Before Refactoring
- ❌ Tight coupling to global variables
- ❌ Cannot mock database/Redis for testing
- ❌ Mixed concerns (SQL + caching + business logic)
- ❌ No unit tests possible

### After Refactoring
- ✅ Explicit dependencies via constructor injection
- ✅ Easy to mock with interfaces
- ✅ Separated concerns (repository, cache, service layers)
- ✅ Comprehensive unit tests with mocks
- ✅ Swappable implementations (e.g., different cache backends)

## Running Tests

```bash
# Run all tests
go test ./server_api/database/... -v

# Run only unit tests
go test ./server_api/database/... -run "^Test.*" -v

# Run with coverage
go test ./server_api/database/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Next Steps

1. **Migrate Handlers**: Update handlers to use `DatabaseService` instead of `UserDatabase`
2. **Add More Tests**: Write tests for all repository methods
3. **Extract Business Logic**: Move password verification, token generation to service layer
4. **Add Integration Tests**: Test with real PostgreSQL using testcontainers
5. **Document API**: Create API documentation for all repository interfaces
