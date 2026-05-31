# Service Layer Integration Example

## How to Integrate Services into main.go

### Step 1: Initialize Services in initialize()

**Current code (main.go):**
```go
var user_database *database.UserDatabase
var redis_manager *RedisManager
var repeatitive_function *RepeatitiveFunctionProvide
```

**New code:**
```go
var dbService *database.DatabaseService
var services *service.Services
```

### Step 2: Update initialize() Function

```go
func initialize() {
    // Load language pack
    lp = lang_pack_load.NewLanguagePack("server_api/main.json", LANGUAGE)
    lp.Load()

    // Initialize database service (new way)
    var err error
    dbService, err = database.InitializeDatabaseService(
        database.Config{
            DB_HOST:     DB_HOST,
            DB_PORT:     DB_PORT,
            DB_USER:     DB_USER,
            DB_PASS:     DB_PASS,
            DB_NAME:     DB_NAME,
            DB_SSL:      DB_SSL,
            MAXOPENCONNS: MAXOPENCONNS,
            MAXIDLECONNS: MAXIDLECONNS,
            CONNMAXLIFETIME: CONNMAXLIFETIME,
            CONNMAXIDLETIME: CONNMAXIDLETIME,
            MAX_LOGIN_ATTEMPTS: MAX_LOGIN_ATTEMPTS,
            LOCKOUT_DURATION: LOCKOUT_DURATION,
            IP_LIMIT_TIME: IP_LIMIT_TIME,
            IP_LIMIT_VISIT_TIMES: IP_LIMIT_VISIT_TIMES,
            IP_LIMIT_LOCKOUT_DURATION: IP_LIMIT_LOCKOUT_DURATION,
        },
        redis_manager,
    )
    if err != nil {
        log.Fatal("Failed to initialize database service: ", err)
    }

    // Initialize services (new!)
    services = service.NewServices(dbService, service.ServiceConfig{
        JWTSecret:         JWT_SECRET,
        JWTExpiration:     JWT_EXPIRE_TIME,
        MaxUsernameLength: ALLOW_MAX_LENGTH_OF_USERNAME,
        VerifyTokenTTL:    VERIFY_TOKEN_EXPIRE_TIME,
    })

    // Initialize web server
    web_server_engine = gin.Default()
    web_server_api = &WebServerApi{
        authService: services.AuthService,
        userService: services.UserService,
        tokenService: services.TokenService,
    }
    
    // ... rest of initialization
}
```

### Step 3: Update WebServerApi Structure

**Current:**
```go
type WebServerApi struct {
    database *database.UserDatabase
}
```

**New:**
```go
type WebServerApi struct {
    authService  service.AuthService
    userService  service.UserService
    tokenService service.TokenService
}
```

### Step 4: Update Handlers to Use Services

#### Example: Login Handler

**Before:**
```go
func (api *WebServerApi) Login(c *gin.Context) {
    username := c.PostForm("username")
    password := c.PostForm("password")
    verifyToken := c.PostForm("verify_token")

    // Verify the verification token
    if !repeatitive_function.validateVerifyToken(c, verifyToken) {
        return
    }

    // Check if username and password are provided
    if username == "" || password == "" {
        repeatitive_function.sendErrorResponse(...)
        return
    }

    // Check if username exists
    if !user_database.CheckUsernameIsExist(username) {
        repeatitive_function.sendErrorResponse(...)
        return
    }

    // Check if account is locked
    if repeatitive_function.checkAccountLockout(c, username) {
        return
    }

    // Check if password is correct
    user_id, pass := user_database.SelectUserIsExist(username, password)
    if !pass {
        user_database.TrackLoginAttempt(username, false)
        c.JSON(http.StatusUnauthorized, gin.H{
            "message": lp.G("username_or_password_is_invalid"),
        })
        return
    }

    // Track successful login
    user_database.TrackLoginAttempt(username, true)
    repeatitive_function.generateAndReturnToken(c, user_id, "User id %d login successful")
}
```

**After:**
```go
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
        switch err {
        case service.ErrInvalidInput:
            c.JSON(http.StatusBadRequest, gin.H{
                "message": lp.G("username_and_password_required"),
            })
        case service.ErrInvalidToken:
            c.JSON(http.StatusUnauthorized, gin.H{
                "message": lp.G("invalid_or_expired_verification_token"),
            })
        case service.ErrAccountLocked:
            c.JSON(http.StatusTooManyRequests, gin.H{
                "message": lp.G("account_locked_try_later"),
            })
        case service.ErrInvalidPassword:
            c.JSON(http.StatusUnauthorized, gin.H{
                "message": lp.G("username_or_password_is_invalid"),
            })
        default:
            log.WithFields(log.Fields{
                "error": err.Error(),
            }).Error("Login failed")
            c.JSON(http.StatusInternalServerError, gin.H{
                "message": lp.G("internal_server_error"),
            })
        }
        return
    }

    // Return success response
    c.JSON(http.StatusOK, gin.H{
        "user_token": resp.UserToken,
        "user_id":    resp.UserID,
    })
}
```

#### Example: Register Handler

**After:**
```go
func (api *WebServerApi) Register(c *gin.Context) {
    req := &service.RegisterRequest{
        Username:    c.PostForm("username"),
        Password:    c.PostForm("password"),
        VerifyToken: c.PostForm("verify_token"),
    }

    resp, err := api.authService.Register(c.Request.Context(), req)
    if err != nil {
        switch err {
        case service.ErrInvalidInput:
            c.JSON(http.StatusBadRequest, gin.H{
                "message": lp.G("username_and_password_required"),
            })
        case service.ErrInvalidToken:
            c.JSON(http.StatusUnauthorized, gin.H{
                "message": lp.G("invalid_or_expired_verification_token"),
            })
        case service.ErrUsernameAlreadyExists:
            c.JSON(http.StatusConflict, gin.H{
                "message": lp.G("username_already_exists"),
            })
        default:
            log.WithFields(log.Fields{
                "error": err.Error(),
            }).Error("Registration failed")
            c.JSON(http.StatusInternalServerError, gin.H{
                "message": lp.G("internal_server_error"),
            })
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "user_token": resp.UserToken,
        "user_id":    resp.UserID,
    })
}
```

#### Example: Get User Profile Handler

**After:**
```go
func (api *WebServerApi) GetUserProfile(c *gin.Context) {
    userID, err := strconv.Atoi(c.Param("user_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "message": lp.G("invalid_user_id_format"),
        })
        return
    }

    verifyToken := c.Query("verify_token")
    if err := api.tokenService.ValidateAndConsumeToken(verifyToken); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "message": lp.G("invalid_or_expired_verification_token"),
        })
        return
    }

    profile, err := api.userService.GetUserProfile(c.Request.Context(), userID)
    if err != nil {
        switch err {
        case service.ErrUserNotFound:
            c.JSON(http.StatusNotFound, gin.H{
                "message": lp.G("user_not_found"),
            })
        default:
            log.WithFields(log.Fields{
                "error": err.Error(),
            }).Error("Failed to get user profile")
            c.JSON(http.StatusInternalServerError, gin.H{
                "message": lp.G("internal_server_error"),
            })
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": lp.G("user_info"),
        "data":    profile,
    })
}
```

#### Example: Delete User Handler

**After:**
```go
func (api *WebServerApi) DeleteUser(c *gin.Context) {
    userID, err := strconv.Atoi(c.Param("user_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "message": lp.G("invalid_user_id_format"),
        })
        return
    }

    req := &service.DeleteUserRequest{
        UserID:      userID,
        Password:    c.PostForm("user_password"),
        VerifyToken: c.PostForm("verify_token"),
    }

    err = api.userService.DeleteUser(c.Request.Context(), req)
    if err != nil {
        switch err {
        case service.ErrInvalidInput:
            c.JSON(http.StatusBadRequest, gin.H{
                "message": lp.G("user_password_required"),
            })
        case service.ErrInvalidToken:
            c.JSON(http.StatusUnauthorized, gin.H{
                "message": lp.G("invalid_or_expired_verification_token"),
            })
        case service.ErrAccountLocked:
            c.JSON(http.StatusTooManyRequests, gin.H{
                "message": lp.G("account_locked_try_later"),
            })
        case service.ErrInvalidPassword:
            c.JSON(http.StatusUnauthorized, gin.H{
                "message": lp.G("invalid_user_id_or_password"),
            })
        case service.ErrUserNotFound:
            c.JSON(http.StatusNotFound, gin.H{
                "message": lp.G("user_not_found"),
            })
        default:
            log.WithFields(log.Fields{
                "error": err.Error(),
            }).Error("Failed to delete user")
            c.JSON(http.StatusInternalServerError, gin.H{
                "message": lp.G("internal_server_error"),
            })
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": lp.G("user_deleted_successfully"),
    })
}
```

### Step 5: Update main() Cleanup

**Current:**
```go
defer user_database.Database.Close()
defer redis_manager.redis.Close()
```

**New:**
```go
defer dbService.GetDB().Close()
defer redis_manager.redis.Close()
```

## Benefits of This Integration

1. **Cleaner Handlers**: Handlers only deal with HTTP concerns
2. **Testable Business Logic**: Services can be unit tested independently
3. **Standardized Errors**: Consistent error handling across all endpoints
4. **Context Support**: Services accept context for cancellation/timeout
5. **Dependency Injection**: All dependencies are explicit

## Migration Checklist

- [ ] Create service package ✅
- [ ] Define service interfaces ✅
- [ ] Implement AuthService ✅
- [ ] Implement UserService ✅
- [ ] Implement TokenService ✅
- [ ] Write service tests ✅
- [ ] Update main.go initialization
- [ ] Update WebServerApi structure
- [ ] Migrate Login handler
- [ ] Migrate Register handler
- [ ] Migrate GetUserProfile handler
- [ ] Migrate ChangeUserProfile handler
- [ ] Migrate DeleteUser handler
- [ ] Update GetVerifyToken handler
- [ ] Remove old global variables
- [ ] Test all endpoints
- [ ] Add integration tests

## Testing the New Code

```bash
# Run service tests
go test ./server_api/service/... -v

# Run with coverage
go test ./server_api/service/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

**Ready to integrate!** Start with one handler at a time to minimize risk.
