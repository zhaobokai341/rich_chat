package ui_handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUIHandler(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockUser := new(MockUserService)
	mockAPI := new(MockAPIClient)
	mockChatAPI := new(MockChatAPIClient)
	mockChat := new(MockChatService)
	mockConfig := new(MockConfigManager)
	mockLP := new(MockLanguagePack)

	handler := NewUIHandler(mockAuth, mockUser, mockAPI, mockChatAPI, mockChat, mockConfig, mockLP, "test_config")

	assert.NotNil(t, handler)
	assert.Equal(t, mockAuth, handler.authService)
	assert.Equal(t, mockUser, handler.userService)
	assert.Equal(t, mockAPI, handler.apiClient)
	assert.Equal(t, mockChatAPI, handler.chatAPIClient)
	assert.Equal(t, mockChat, handler.chatService)
	assert.Equal(t, mockConfig, handler.configMgr)
	assert.Equal(t, mockLP, handler.languagePack)
	assert.Equal(t, "test_config", handler.configDir)
	assert.NotNil(t, handler.printer)
	assert.NotNil(t, handler.menuRenderer)
}

func TestUIHandler_PrintInfo(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "test_key").Return("Test message")

	handler := &UIHandler{
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	handler.printInfo("test_key")
	mockLP.AssertExpectations(t)
}

func TestUIHandler_PrintWarning(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "test_key").Return("Test warning")

	handler := &UIHandler{
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	handler.printWarning("test_key")
	mockLP.AssertExpectations(t)
}

func TestUIHandler_PrintSuccess(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "test_key").Return("Test success")

	handler := &UIHandler{
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	handler.printSuccess("test_key")
	mockLP.AssertExpectations(t)
}

func TestUIHandler_PrintError_String(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "error_key").Return("Error message")

	handler := &UIHandler{
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	handler.printError("error_key")
	mockLP.AssertExpectations(t)
}

func TestUIHandler_PrintError_Error(t *testing.T) {
	handler := &UIHandler{
		languagePack: new(MockLanguagePack),
		printer:      NewMessagePrinter(),
	}

	handler.printError(assert.AnError)
}

func TestUIHandler_PrintError_Default(t *testing.T) {
	handler := &UIHandler{
		languagePack: new(MockLanguagePack),
		printer:      NewMessagePrinter(),
	}

	handler.printError(123)
}

func TestMenuRenderer_RenderTitle(t *testing.T) {
	mr := NewMenuRenderer()
	title := mr.RenderTitle("Test Title")
	assert.NotEmpty(t, title)
}

func TestMenuRenderer_RenderMenu(t *testing.T) {
	mr := NewMenuRenderer()
	items := []string{"Item 1", "Item 2", "Item 3"}
	menu := mr.RenderMenu(items)
	assert.NotEmpty(t, menu)
	assert.Contains(t, menu, "1. Item 1")
	assert.Contains(t, menu, "2. Item 2")
	assert.Contains(t, menu, "3. Item 3")
}

func TestMenuRenderer_RenderMenu_Empty(t *testing.T) {
	mr := NewMenuRenderer()
	menu := mr.RenderMenu([]string{})
	assert.NotEmpty(t, menu)
}

func TestUIHandler_CheckServer_Success(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("CheckServerHealth").Return(true, nil)

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "connecting_server").Return("Connecting to server...")
	mockLP.On("Get", "connected_server").Return("Connected to server")

	handler := &UIHandler{
		apiClient:    mockAPI,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	result := handler.checkServer()
	assert.True(t, result)
	mockAPI.AssertExpectations(t)
}

func TestUIHandler_CheckServer_Error(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("CheckServerHealth").Return(false, assert.AnError)

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "connecting_server").Return("Connecting to server...")
	mockLP.On("Get", "connection_error").Return("Connection error")

	handler := &UIHandler{
		apiClient:    mockAPI,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	result := handler.checkServer()
	assert.False(t, result)
	mockAPI.AssertExpectations(t)
}

func TestUIHandler_CheckServer_HealthCheckFailed(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("CheckServerHealth").Return(false, nil)

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "connecting_server").Return("Connecting to server...")
	mockLP.On("Get", "server_health_check_failed").Return("Server health check failed")

	handler := &UIHandler{
		apiClient:    mockAPI,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	result := handler.checkServer()
	assert.False(t, result)
	mockAPI.AssertExpectations(t)
}

func TestUIHandler_HandleAuthenticated_CredentialsError(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockAuth.On("GetCredentials").Return("", "", assert.AnError)
	mockAuth.On("Logout").Return(nil)

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "user_id_or_token_not_string").Return("User ID or token not string")

	handler := &UIHandler{
		authService:  mockAuth,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	handler.handleAuthenticated()
	mockAuth.AssertExpectations(t)
}

func TestParseUserID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedID  int
		expectError bool
	}{
		{
			name:        "valid user ID",
			input:       "123",
			expectedID:  123,
			expectError: false,
		},
		{
			name:        "invalid user ID - not a number",
			input:       "abc",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "invalid user ID - zero",
			input:       "0",
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "invalid user ID - negative",
			input:       "-1",
			expectedID:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := parseUserID(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}
