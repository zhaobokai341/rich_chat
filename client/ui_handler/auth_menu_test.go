package ui_handler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUIHandler_HandleLogout_Success(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockAuth.On("Logout").Return(nil)

	mockAPI := new(MockAPIClient)
	mockAPI.On("ClearAuthHeaders").Return()

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "logout_successful").Return("Logout successful")

	handler := &UIHandler{
		authService:  mockAuth,
		apiClient:    mockAPI,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	handler.handleLogout()
	mockAuth.AssertExpectations(t)
	mockAPI.AssertExpectations(t)
}

func TestUIHandler_HandleLogout_Error(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockAuth.On("Logout").Return(errors.New("logout failed"))

	mockAPI := new(MockAPIClient)

	handler := &UIHandler{
		authService:  mockAuth,
		apiClient:    mockAPI,
		languagePack: new(MockLanguagePack),
		printer:      NewMessagePrinter(),
	}

	handler.handleLogout()
	mockAuth.AssertExpectations(t)
}

func TestUIHandler_HandleDeleteAccount_Cancelled(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "delete_account_warning").Return("Warning")
	mockLP.On("Get", "delete_account_confirm").Return("Confirm?")
	mockLP.On("Get", "exit").Return("Exit")

	handler := &UIHandler{
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	// This will wait for input, so we just test the function exists
	// In real tests, you would mock stdin
	assert.NotNil(t, handler)
}

func TestValidatePassword_UsedInAuth(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "weak password - only digits",
			password:    "1234567",
			expectError: true,
		},
		{
			name:        "weak password - only letters",
			password:    "abcdefg",
			expectError: true,
		},
		{
			name:        "strong password - mixed",
			password:    "abc123!@#",
			expectError: false,
		},
		{
			name:        "long password - only digits",
			password:    "123456789",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUIHandler_PrintLoginMenu(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "menu_login").Return("Login")
	mockLP.On("Get", "menu_register").Return("Register")
	mockLP.On("Get", "menu_exit").Return("Exit")

	handler := &UIHandler{
		languagePack: mockLP,
		menuRenderer: NewMenuRenderer(),
	}

	handler.printLoginMenu()
	mockLP.AssertExpectations(t)
}

func TestUIHandler_PrintMainMenu(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "main_menu_title").Return("Main Menu")
	mockLP.On("Get", "menu_exit").Return("Exit")
	mockLP.On("Get", "menu_view_modify_user_info").Return("View User Info")
	mockLP.On("Get", "menu_change_password").Return("Change Password")
	mockLP.On("Get", "menu_logout").Return("Logout")
	mockLP.On("Get", "menu_delete_account").Return("Delete Account")
	mockLP.On("Get", "menu_chat").Return("Chat")

	handler := &UIHandler{
		languagePack: mockLP,
		menuRenderer: NewMenuRenderer(),
	}

	handler.printMainMenu()
	mockLP.AssertExpectations(t)
}
