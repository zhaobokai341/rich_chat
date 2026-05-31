# Quick Start - Refactored Database Layer

## 🚀 Get Started in 3 Steps

### Step 1: Understand the New Structure

```
server_api/database/
├── types.go                    # Type definitions (User, UserInfo)
├── interfaces.go               # Repository interfaces
├── user_repository.go          # PostgreSQL implementation
├── cached_user_repository.go   # Caching decorator
├── redis_adapter.go            # Redis integration
├── token_repository.go         # Token management
├── database_service.go         # Service orchestrator
├── repository_test.go          # Test examples
└── Documentation/
    ├── REFACTORING.md          # Architecture guide
    ├── USAGE_EXAMPLES.md       # Code examples
    └── REFACTORING_SUMMARY.md  # What changed
```

### Step 2: Initialize the Service

Replace this old code:
```go
user_database := database.DatabaseInit(cfg, redisManager)
```

With new code:
```go
dbService, err := database.InitializeDatabaseService(cfg, redisManager)
if err != nil {
    log.Fatal(err)
}
defer dbService.GetDB().Close()
```

### Step 3: Use Repositories

```go
// Get repositories
userRepo := dbService.GetUserRepository()
rateLimitRepo := dbService.GetRateLimitRepository()
tokenRepo := dbService.GetTokenRepository()

// Use them
exists, _ := userRepo.ExistsByUsername("john")
locked, _ := rateLimitRepo.CheckAccountLocked("john")
valid, _ := tokenRepo.VerifyAndConsumeToken(token)
```

## 📚 Read These Files (In Order)

1. **REFACTORING_SUMMARY.md** - What we did and why (5 min)
2. **USAGE_EXAMPLES.md** - How to use the new code (10 min)
3. **REFACTORING.md** - Deep dive into architecture (15 min)
4. **repository_test.go** - See testing patterns (10 min)

## 🎯 Key Concepts

### 1. Repository Pattern
```go
// Interface defines what operations are available
type UserRepository interface {
    FindByID(id int) (*User, error)
    Create(username, hash string) (int, error)
    // ...
}

// Implementation does the actual work
repo := NewPostgresUserRepository(db)
```

### 2. Dependency Injection
```go
// Old way - global variables ❌
var user_database *database.UserDatabase

// New way - explicit dependencies ✅
type Handler struct {
    userRepo UserRepository
}

func NewHandler(repo UserRepository) *Handler {
    return &Handler{userRepo: repo}
}
```

### 3. Caching Decorator
```go
// Base repository (no cache)
pgRepo := NewPostgresUserRepository(db)

// Add caching transparently
cachedRepo := NewCachedUserRepository(pgRepo, cache)

// Use it - caching happens automatically
user, _ := cachedRepo.FindByID(123)
```

## 🧪 Writing Tests

### Unit Test (Fast, No DB)
```go
func TestFindUser(t *testing.T) {
    // Create mock
    mockRepo := new(MockUserRepository)
    
    // Setup expectation
    mockRepo.On("FindByID", 123).Return(&User{ID: 123}, nil)
    
    // Test
    user, err := mockRepo.FindByID(123)
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, 123, user.ID)
}
```

### Integration Test (Real DB)
```go
func TestCreateUser(t *testing.T) {
    db := setupTestDB()
    defer db.Close()
    
    repo := NewPostgresUserRepository(db)
    
    userID, err := repo.CreateUser("test", "hash")
    
    assert.NoError(t, err)
    assert.Greater(t, userID, 0)
}
```

## ⚡ Migration Checklist

For each handler/function using `user_database`:

- [ ] Replace `user_database.Method()` with `userRepo.Method()`
- [ ] Update error handling (new methods return errors)
- [ ] Remove global variable references
- [ ] Add unit tests with mocks

Example:
```diff
- if user_database.CheckUsernameIsExist(username) {
+ exists, _ := userRepo.ExistsByUsername(username)
+ if exists {
```

## 🔍 Common Patterns

### Pattern 1: Check then Act
```go
exists, err := userRepo.ExistsByUsername(username)
if err != nil {
    return err
}
if !exists {
    return ErrUserNotFound
}
```

### Pattern 2: Create and Return
```go
userID, err := userRepo.CreateUser(username, hash)
if err != nil {
    return err
}
return userID, nil
```

### Pattern 3: Update with Cache Invalidation
```go
err := userRepo.UpdateProfile(userID, "nickname", "NewName")
if err != nil {
    return err
}
// Cache automatically invalidated by CachedUserRepository
```

## 💡 Tips

1. **Start Small**: Migrate one handler at a time
2. **Write Tests First**: Protect existing functionality
3. **Use IDE Features**: Find all references to `user_database`
4. **Keep Old Code**: Don't delete until new code is tested
5. **Read the Docs**: Examples are in USAGE_EXAMPLES.md

## ❓ FAQ

**Q: Do I need to change everything at once?**  
A: No! The old API still works. Migrate gradually.

**Q: How do I test without a real database?**  
A: Use the mock implementations in `repository_test.go`.

**Q: Where did the caching logic go?**  
A: It's in `CachedUserRepository` - transparent decorator.

**Q: Can I still use the old UserDatabase?**  
A: Yes! It's still there for backward compatibility.

**Q: What about my existing tests?**  
A: They should still work. Add new tests for new code.

## 🎓 Next Steps

After mastering the database layer:

1. **Refactor Service Layer** - Move business logic out of handlers
2. **Add More Tests** - Aim for 80% coverage
3. **Document Your Changes** - Update team wiki/docs
4. **Share Knowledge** - Teach teammates the new patterns

---

**Need Help?** Check these files:
- Architecture questions → `REFACTORING.md`
- Code examples → `USAGE_EXAMPLES.md`
- Testing help → `repository_test.go`
- What changed → `REFACTORING_SUMMARY.md`
