package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRestAPIClient(t *testing.T) {
	languagePack := NewLanguagePackWrapper("client/main.json", "zh")
	httpClient := NewHTTPClient("test-agent")

	client := NewRestAPIClient(httpClient, "http://test.com", languagePack)

	assert.NotNil(t, client)
	assert.Equal(t, "http://test.com", client.baseURL)
	assert.NotNil(t, client.client)
	assert.NotNil(t, client.languagePack)
}

func TestRestAPIClient_GetVerifyToken(t *testing.T) {
	languagePack := NewLanguagePackWrapper("client/main.json", "zh")
	httpClient := NewHTTPClient("test-agent")

	client := NewRestAPIClient(httpClient, "http://test.invalid", languagePack)

	_, err := client.GetVerifyToken()
	assert.Error(t, err)
}

func TestRestAPIClient_Login(t *testing.T) {
	languagePack := NewLanguagePackWrapper("client/main.json", "zh")
	httpClient := NewHTTPClient("test-agent")

	client := NewRestAPIClient(httpClient, "http://test.invalid", languagePack)

	_, err := client.Login("test", "password", "token")
	assert.Error(t, err)
}

func TestRestAPIClient_Register(t *testing.T) {
	languagePack := NewLanguagePackWrapper("client/main.json", "zh")
	httpClient := NewHTTPClient("test-agent")

	client := NewRestAPIClient(httpClient, "http://test.invalid", languagePack)

	_, err := client.Register("test", "password", "token")
	assert.Error(t, err)
}

func TestRestAPIClient_DeleteUser(t *testing.T) {
	languagePack := NewLanguagePackWrapper("client/main.json", "zh")
	httpClient := NewHTTPClient("test-agent")

	client := NewRestAPIClient(httpClient, "http://test.invalid", languagePack)

	err := client.DeleteUser("1", "password", "token")
	assert.Error(t, err)
}

func TestRestAPIClient_GetUserProfile(t *testing.T) {
	languagePack := NewLanguagePackWrapper("client/main.json", "zh")
	httpClient := NewHTTPClient("test-agent")

	client := NewRestAPIClient(httpClient, "http://test.invalid", languagePack)

	_, err := client.GetUserProfile("1", "token")
	assert.Error(t, err)
}

func TestRestAPIClient_UpdateUserProfile(t *testing.T) {
	languagePack := NewLanguagePackWrapper("client/main.json", "zh")
	httpClient := NewHTTPClient("test-agent")

	client := NewRestAPIClient(httpClient, "http://test.invalid", languagePack)

	err := client.UpdateUserProfile("1", "nickname", "newname", "token")
	assert.Error(t, err)
}

func TestRestAPIClient_ChangePassword(t *testing.T) {
	languagePack := NewLanguagePackWrapper("client/main.json", "zh")
	httpClient := NewHTTPClient("test-agent")

	client := NewRestAPIClient(httpClient, "http://test.invalid", languagePack)

	err := client.ChangePassword("1", "oldpass", "newpass", "token")
	assert.Error(t, err)
}

func TestRestAPIClient_CheckServerHealth(t *testing.T) {
	languagePack := NewLanguagePackWrapper("client/main.json", "zh")
	httpClient := NewHTTPClient("test-agent")

	client := NewRestAPIClient(httpClient, "http://test.invalid", languagePack)

	_, err := client.CheckServerHealth()
	assert.Error(t, err)
}

func TestAuthResponseStructure(t *testing.T) {
	resp := &AuthResponse{
		UserID:    1,
		UserToken: "test-token",
		Message:   "success",
	}

	assert.Equal(t, 1, resp.UserID)
	assert.Equal(t, "test-token", resp.UserToken)
	assert.Equal(t, "success", resp.Message)
}

func TestUserInfoResponseStructure(t *testing.T) {
	resp := &UserInfoResponse{
		Data: &UserData{
			Username: "testuser",
			Email:    "test@example.com",
			Nickname: "Test",
			Bio:      "Hello",
		},
	}

	assert.NotNil(t, resp.Data)
	assert.Equal(t, "testuser", resp.Data.Username)
	assert.Equal(t, "test@example.com", resp.Data.Email)
}
