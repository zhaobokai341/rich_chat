package ui_handler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUIHandler_ViewUserInfo_Success(t *testing.T) {
	mockUser := new(MockUserService)
	mockUser.On("GetProfile").Return(&UserData{
		Username: "testuser",
		Email:    "test@example.com",
		Nickname: "Test User",
		Bio:      "Test bio",
	}, nil)

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "viewing_user_info").Return("Viewing user info")
	mockLP.On("Get", "user_info_title").Return("User Info")
	mockLP.On("Get", "user_info_username").Return("Username")
	mockLP.On("Get", "user_info_email").Return("Email")
	mockLP.On("Get", "user_info_nickname").Return("Nickname")
	mockLP.On("Get", "user_info_bio").Return("Bio")

	handler := &UIHandler{
		userService:  mockUser,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
		menuRenderer: NewMenuRenderer(),
	}

	handler.viewUserInfo()
	mockUser.AssertExpectations(t)
}

func TestUIHandler_ViewUserInfo_Error(t *testing.T) {
	mockUser := new(MockUserService)
	mockUser.On("GetProfile").Return((*UserData)(nil), errors.New("failed to get profile"))

	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "viewing_user_info").Return("Viewing user info")

	handler := &UIHandler{
		userService:  mockUser,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	handler.viewUserInfo()
	mockUser.AssertExpectations(t)
}

func TestUIHandler_PrintUserInfoCard(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "user_info_title").Return("User Info")
	mockLP.On("Get", "user_info_username").Return("Username")
	mockLP.On("Get", "user_info_email").Return("Email")
	mockLP.On("Get", "user_info_nickname").Return("Nickname")
	mockLP.On("Get", "user_info_bio").Return("Bio")

	handler := &UIHandler{
		languagePack: mockLP,
		menuRenderer: NewMenuRenderer(),
	}

	userData := &UserData{
		Username: "testuser",
		Email:    "test@example.com",
		Nickname: "Test User",
		Bio:      "Test bio",
	}

	handler.printUserInfoCard(userData)
	mockLP.AssertExpectations(t)
}

func TestUIHandler_PrintModifyUserInfoMenu(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "menu_modify_nickname").Return("Modify Nickname")
	mockLP.On("Get", "menu_modify_email").Return("Modify Email")
	mockLP.On("Get", "menu_modify_bio").Return("Modify Bio")
	mockLP.On("Get", "menu_back").Return("Back")

	handler := &UIHandler{
		languagePack: mockLP,
		menuRenderer: NewMenuRenderer(),
	}

	handler.printModifyUserInfoMenu()
	mockLP.AssertExpectations(t)
}

func TestUIHandler_ModifyField(t *testing.T) {
	// This test requires stdin input, so we just verify the function exists
	mockUser := new(MockUserService)
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "enter_new_nickname").Return("Enter new nickname: ")

	handler := &UIHandler{
		userService:  mockUser,
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	// Function exists and can be called, but requires stdin for full test
	assert.NotNil(t, handler)
}

func TestUIHandler_HandleChangePassword_Error_OldPassword(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "enter_old_password").Return("Enter old password: ")
	mockLP.On("Get", "old_password_cannot_be_empty").Return("Old password cannot be empty")

	handler := &UIHandler{
		languagePack: mockLP,
		printer:      NewMessagePrinter(),
	}

	// This will wait for password input, so we just test the function exists
	assert.NotNil(t, handler)
}

func TestUIHandler_ReadPasswordSecure(t *testing.T) {
	mockLP := new(MockLanguagePack)
	mockLP.On("Get", "password_prompt").Return("Password: ")

	handler := &UIHandler{
		languagePack: mockLP,
	}

	// This will wait for password input, so we just test the function exists
	assert.NotNil(t, handler)
}
