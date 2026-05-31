# Client Architecture Documentation

## Overview

The client has been refactored following the same architectural patterns as `server_api` to improve readability, maintainability, and testability.

## Architecture

The client now follows a **layered architecture** with clear separation of concerns:

```
┌─────────────────────────────────────┐
│         UI Handler Layer            │  ← User interaction & flow control
├─────────────────────────────────────┤
│        Service Layer                │  ← Business logic
│  ┌──────────┐  ┌──────────────┐    │
│  │ AuthService│  │ UserService  │    │
│  └──────────┘  └──────────────┘    │
├─────────────────────────────────────┤
│        API Client Layer             │  ← HTTP request encapsulation
│  ┌──────────┐  ┌──────────────┐    │
│  │HTTPClient│  │RestAPIClient │    │
│  └──────────┘  └──────────────┘    │
├─────────────────────────────────────┤
│       Infrastructure Layer          │  ← Configuration, utilities
│  ┌──────────┐  ┌──────────────┐    │
│  │ConfigMgr │  │TokenExtractor│    │
│  └──────────┘  └──────────────┘    │
└─────────────────────────────────────┘
```

## File Structure

### Core Files

- **`config.go`** - Configuration constants (unchanged as requested)
  - Language settings
  - URL configuration
  - HTTP Basic Auth credentials
  - User agent and limits

- **`main.go`** - Application entry point and dependency injection
  - `Application` struct holds all dependencies
  - `NewApplication()` creates and wires all components
  - Clean initialization without global variables

### API Client Layer

- **`http_client.go`** - HTTP client wrapper
  - Wraps `resty.Client` for better abstraction
  - Provides testable interface
  - Manages headers and base configuration

- **`api_client.go`** - API operations interface and implementation
  - `APIClient` interface defines all API operations
  - `RestAPIClient` implements the interface using REST/HTTP
  - Handles all HTTP requests and response parsing
  - Unified error handling

### Service Layer

- **`auth_service.go`** - Authentication business logic
  - Login/register flows
  - Credential management
  - Session state tracking
  - Independent of HTTP details

- **`user_service.go`** - User management business logic
  - Profile retrieval and updates
  - Account deletion
  - Independent of HTTP details

### Configuration Layer

- **`config_manager.go`** - Configuration management
  - `ConfigManager` interface for config operations
  - `FileConfigManager` implements JSON file-based storage
  - Handles config directory creation
  - Thread-safe operations

### Utility Layer

- **`token_extractor.go`** - JWT token parsing
  - `TokenExtractor` interface
  - `JWTTokenExtractor` implementation
  - Extracts user ID from JWT tokens

- **`language_pack_wrapper.go`** - Localization wrapper
  - Simplifies language pack access
  - Provides clean API for translations

### UI Layer

- **`ui_handler.go`** - User interface and flow control
  - Manages user interactions
  - Handles menu navigation
  - Coordinates service calls
  - Formats output for users

- **`repeatitive_function_provide.go`** - Shared utilities
  - Styled printing functions
  - User input helper
  - Common formatting

## Key Improvements

### 1. **Eliminated Global Variables**
- All dependencies are injected through constructors
- No implicit global state
- Easier to test and reason about

### 2. **Interface Abstractions**
- `APIClient` interface for API operations
- `ConfigManager` interface for configuration
- `TokenExtractor` interface for token parsing
- Enables mocking for unit tests

### 3. **Separation of Concerns**
- HTTP details isolated in API client layer
- Business logic in service layer (no HTTP dependencies)
- UI logic separate from business logic
- Each layer has a single responsibility

### 4. **Improved Error Handling**
- Consistent error formatting
- Clear error messages
- Proper error wrapping for debugging

### 5. **Better Testability**
- All services can be tested independently
- Mock implementations possible via interfaces
- No need for real HTTP calls in unit tests

### 6. **Cleaner Code Organization**
- Files organized by responsibility
- Each file < 300 lines
- Clear naming conventions
- Easy to navigate

## Usage Example

```go
// Create application with all dependencies wired
app := NewApplication()

// Run the application
app.Run()
```

## Testing Strategy (Future Work)

Following the server_api pattern, you can add:

1. **Unit Tests** - Mock dependencies
   ```go
   // Mock APIClient
   type MockAPIClient struct {
       // Implement APIClient interface
   }
   
   // Mock ConfigManager
   type MockConfigManager struct {
       // Implement ConfigManager interface
   }
   ```

2. **Integration Tests** - Real API calls
3. **E2E Tests** - Full user flows

## Migration Notes

### What Changed
- ✅ Eliminated global variables (`lp`, `requests`)
- ✅ Added dependency injection
- ✅ Created interface abstractions
- ✅ Separated concerns into layers
- ✅ Improved error handling

### What Stayed the Same
- ✅ `config.go` unchanged (as requested)
- ✅ Same user-facing functionality
- ✅ Same API endpoints
- ✅ Same configuration format

### Removed Files
- ❌ `user_input.go` - Replaced by `ui_handler.go`
- ❌ `user_manager.go` - Functionality split into services

## Benefits

1. **Maintainability**: Clear structure makes it easy to find and modify code
2. **Testability**: Interfaces enable comprehensive unit testing
3. **Extensibility**: Easy to add new features without breaking existing code
4. **Readability**: Each component has a single, well-defined purpose
5. **Reusability**: Services can be reused across different UI implementations

## Comparison with server_api

| Aspect | server_api | client (refactored) |
|--------|-----------|-------------------|
| Architecture | Repository-Service-Handler | API Client-Service-UI Handler |
| Dependency Injection | ✅ | ✅ |
| Interface Abstraction | ✅ | ✅ |
| Layered Design | ✅ | ✅ |
| Testability | High | High |
| Global Variables | Minimal | None |

## Next Steps

1. Add unit tests for each service
2. Add integration tests for API client
3. Consider adding a mock API client for testing
4. Document public APIs with godoc comments
5. Add logging support similar to server_api
