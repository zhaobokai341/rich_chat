package ui_handler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUIHandler_PrintChatMenu(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "chat_menu_title").Return("Chat Menu")
	mockLP.On("Get", "chat_my_sessions").Return("My Sessions")
	mockLP.On("Get", "chat_new_chat").Return("New Chat")
	mockLP.On("Get", "chat_generate_keys").Return("Generate Keys")
	mockLP.On("Get", "chat_back_to_main").Return("Back to Main")

	handler := &UIHandler{
		languagePack: mockLP,
		menuRenderer: NewMenuRenderer(),
	}

	handler.printChatMenu()
	mockLP.AssertExpectations(t)
}

func TestUIHandler_PrintSessionList(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "my_sessions_title").Return("My Sessions")

	handler := &UIHandler{
		languagePack: mockLP,
		menuRenderer: NewMenuRenderer(),
	}

	sessions := []*SessionInfo{
		{SessionID: 1, PartnerID: 100, PartnerName: "Alice"},
		{SessionID: 2, PartnerID: 101, PartnerName: ""},
		{SessionID: 3, PartnerID: 102, PartnerName: "Bob"},
	}

	handler.printSessionList(sessions)
	mockLP.AssertExpectations(t)
}

func TestUIHandler_PrintSessionList_Empty(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "my_sessions_title").Return("My Sessions")

	handler := &UIHandler{
		languagePack: mockLP,
		menuRenderer: NewMenuRenderer(),
	}

	handler.printSessionList([]*SessionInfo{})
	mockLP.AssertExpectations(t)
}

func TestUIHandler_GetPartnerName_Success(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("GetVerifyToken").Return("token123", nil)

	mockChatAPI := new(MockChatAPIClient)
	mockChatAPI.On("GetUserBasicInfo", 1, "token", "token123", 100).Return(&UserBasicInfo{
		ID:       100,
		Username: "testuser",
		Nickname: "Test User",
	}, nil)

	handler := &UIHandler{
		apiClient:     mockAPI,
		chatAPIClient: mockChatAPI,
	}

	name := handler.getPartnerName(1, "token", 100)
	assert.Equal(t, "Test User", name)
	mockAPI.AssertExpectations(t)
	mockChatAPI.AssertExpectations(t)
}

func TestUIHandler_GetPartnerName_FallbackToUsername(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("GetVerifyToken").Return("token123", nil)

	mockChatAPI := new(MockChatAPIClient)
	mockChatAPI.On("GetUserBasicInfo", 1, "token", "token123", 100).Return(&UserBasicInfo{
		ID:       100,
		Username: "testuser",
		Nickname: "",
	}, nil)

	handler := &UIHandler{
		apiClient:     mockAPI,
		chatAPIClient: mockChatAPI,
	}

	name := handler.getPartnerName(1, "token", 100)
	assert.Equal(t, "testuser", name)
	mockAPI.AssertExpectations(t)
	mockChatAPI.AssertExpectations(t)
}

func TestUIHandler_GetPartnerName_Error(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("GetVerifyToken").Return("", errors.New("failed"))

	handler := &UIHandler{
		apiClient:     mockAPI,
		chatAPIClient: new(MockChatAPIClient),
	}

	name := handler.getPartnerName(1, "token", 100)
	assert.Equal(t, "User 100", name)
	mockAPI.AssertExpectations(t)
}

func TestUIHandler_GetPartnerName_APIError(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("GetVerifyToken").Return("token123", nil)

	mockChatAPI := new(MockChatAPIClient)
	mockChatAPI.On("GetUserBasicInfo", 1, "token", "token123", 100).Return((*UserBasicInfo)(nil), errors.New("not found"))

	handler := &UIHandler{
		apiClient:     mockAPI,
		chatAPIClient: mockChatAPI,
	}

	name := handler.getPartnerName(1, "token", 100)
	assert.Equal(t, "User 100", name)
	mockAPI.AssertExpectations(t)
	mockChatAPI.AssertExpectations(t)
}

func TestUIHandler_PrintChatConnected(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "chat_connected_title").Return("Chat Connected")
	mockLP.On("Get", "chat_connected_message").Return("Connected to %s")

	handler := &UIHandler{
		languagePack: mockLP,
		menuRenderer: NewMenuRenderer(),
	}

	handler.printChatConnected("Alice")
	mockLP.AssertExpectations(t)
}

func TestUIHandler_HandleGenerateKeys_Success(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("GetVerifyToken").Return("token123", nil)

	mockChat := new(MockChatService)
	mockChat.On("GenerateAndUploadKeys", 1, "token", "token123").Return(nil)

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "generating_keys").Return("Generating keys...")
	mockLP.On("Get", "keys_generated_successfully").Return("Keys generated successfully")

	handler := &UIHandler{
		apiClient:    mockAPI,
		chatService:  mockChat,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	handler.handleGenerateKeys(1, "token")
	mockAPI.AssertExpectations(t)
	mockChat.AssertExpectations(t)
}

func TestUIHandler_HandleGenerateKeys_APIError(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("GetVerifyToken").Return("", errors.New("failed"))

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "generating_keys").Return("Generating keys...")
	mockLP.On("Get", "getting_verify_token_failed").Return("Getting verify token failed")

	handler := &UIHandler{
		apiClient:    mockAPI,
		chatService:  new(MockChatService),
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	handler.handleGenerateKeys(1, "token")
	mockAPI.AssertExpectations(t)
}

func TestUIHandler_HandleGenerateKeys_ChatError(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("GetVerifyToken").Return("token123", nil)

	mockChat := new(MockChatService)
	mockChat.On("GenerateAndUploadKeys", 1, "token", "token123").Return(errors.New("upload failed"))

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "generating_keys").Return("Generating keys...")

	handler := &UIHandler{
		apiClient:    mockAPI,
		chatService:  mockChat,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	handler.handleGenerateKeys(1, "token")
	mockAPI.AssertExpectations(t)
	mockChat.AssertExpectations(t)
}

func TestParseUserID_ChatMenu(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedID  int
		expectError bool
	}{
		{
			name:        "valid user ID",
			input:       "42",
			expectedID:  42,
			expectError: false,
		},
		{
			name:        "invalid user ID - letters",
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

func TestUIHandler_HandleGenerateNewKey_Success(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("GetVerifyToken").Return("token123", nil)

	mockChat := new(MockChatService)
	mockChat.On("GenerateAndUploadKeys", 1, "token", "token123").Return(nil)
	mockChat.On("LoadEncryptionKey", 1).Return(nil)

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "generating_keys_for_chat").Return("Generating keys...")
	mockLP.On("Get", "private_key_security_warning").Return("Security warning: %s")

	handler := &UIHandler{
		apiClient:    mockAPI,
		chatService:  mockChat,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	result := handler.handleGenerateNewKey(1, "token", "/path/to/key.pem")
	assert.True(t, result)
	mockAPI.AssertExpectations(t)
	mockChat.AssertExpectations(t)
	mockLP.AssertExpectations(t)
}

func TestUIHandler_HandleGenerateNewKey_APIError(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("GetVerifyToken").Return("", errors.New("failed"))

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "generating_keys_for_chat").Return("Generating keys...")
	mockLP.On("Get", "getting_verify_token_failed").Return("Getting verify token failed")

	handler := &UIHandler{
		apiClient:    mockAPI,
		chatService:  new(MockChatService),
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	result := handler.handleGenerateNewKey(1, "token", "/path/to/key.pem")
	assert.False(t, result)
	mockAPI.AssertExpectations(t)
	mockLP.AssertExpectations(t)
}

func TestUIHandler_HandleGenerateNewKey_GenerateError(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("GetVerifyToken").Return("token123", nil)

	mockChat := new(MockChatService)
	mockChat.On("GenerateAndUploadKeys", 1, "token", "token123").Return(errors.New("generation failed"))

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "generating_keys_for_chat").Return("Generating keys...")
	mockLP.On("Get", "key_generation_failed_continue_anyway").Return("Key generation failed")

	handler := &UIHandler{
		apiClient:    mockAPI,
		chatService:  mockChat,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	result := handler.handleGenerateNewKey(1, "token", "/path/to/key.pem")
	assert.True(t, result)
	mockAPI.AssertExpectations(t)
	mockChat.AssertExpectations(t)
	mockLP.AssertExpectations(t)
}

func TestUIHandler_HandleGenerateNewKey_LoadKeyError(t *testing.T) {
	mockAPI := new(MockAPIClient)
	mockAPI.On("GetVerifyToken").Return("token123", nil)

	mockChat := new(MockChatService)
	mockChat.On("GenerateAndUploadKeys", 1, "token", "token123").Return(nil)
	mockChat.On("LoadEncryptionKey", 1).Return(errors.New("load failed"))

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "generating_keys_for_chat").Return("Generating keys...")
	mockLP.On("Get", "private_key_security_warning").Return("Security warning: %s")
	mockLP.On("Get", "key_load_failed: load failed").Return("Key load failed: load failed")

	handler := &UIHandler{
		apiClient:    mockAPI,
		chatService:  mockChat,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	result := handler.handleGenerateNewKey(1, "token", "/path/to/key.pem")
	assert.True(t, result)
	mockAPI.AssertExpectations(t)
	mockChat.AssertExpectations(t)
	mockLP.AssertExpectations(t)
}
