package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// APIClient defines the interface for API operations
type APIClient interface {
	GetVerifyToken() (string, error)
	Login(username, password, verifyToken string) (*AuthResponse, error)
	Register(username, password, verifyToken string) (*AuthResponse, error)
	DeleteUser(userID, token, password, verifyToken string) error
	GetUserProfile(userID, token, verifyToken string) (*UserInfoResponse, error)
	UpdateUserProfile(userID, token, key, value, verifyToken string) error
	ChangePassword(userID, token, oldPassword, newPassword, verifyToken string) error
	CheckServerHealth() (bool, error)
}

// AuthResponse represents authentication response from server
type AuthResponse struct {
	UserID    int    `json:"user_id"`
	UserToken string `json:"user_token"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

// UserInfoResponse represents user profile response
type UserInfoResponse struct {
	Data *UserData `json:"data"`
}

// UserData represents user profile data
type UserData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
}

// ErrorResponse represents API error response
type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// RestAPIClient implements APIClient using resty
type RestAPIClient struct {
	client       *HTTPClient
	baseURL      string
	languagePack *LanguagePackWrapper
}

// NewRestAPIClient creates a new REST API client
func NewRestAPIClient(client *HTTPClient, baseURL string, languagePack *LanguagePackWrapper) *RestAPIClient {
	return &RestAPIClient{
		client:       client,
		baseURL:      baseURL,
		languagePack: languagePack,
	}
}

// handleAPIResponse handles common API response patterns
func (c *RestAPIClient) handleAPIResponse(resp *resty.Response, successStatus int, operationFailed string) error {
	if resp.StatusCode() == 429 {
		return errors.New(c.languagePack.Get("account_locked_try_later"))
	}

	if resp.StatusCode() != successStatus {
		var errorResp ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			msg := errorResp.Message
			if msg == "" {
				msg = errorResp.Error
			}
			if msg != "" {
				if isAuthenticationFailure(msg) {
					return errors.New(c.languagePack.Get("authentication_failed"))
				}
				return fmt.Errorf(c.languagePack.Get(operationFailed), msg)
			}
		}
		return fmt.Errorf(c.languagePack.Get(operationFailed+"_with_status"), resp.StatusCode())
	}

	return nil
}

// isAuthenticationFailure checks if error message indicates authentication failure
func isAuthenticationFailure(msg string) bool {
	return msg == "Authentication failed, please check your credentials" ||
		msg == "认证失败，请检查您的凭据"
}

// GetVerifyToken retrieves a verification token from the server
func (c *RestAPIClient) GetVerifyToken() (string, error) {
	resp, err := c.client.R().
		SetQueryParam("language", getCurrentLanguage()).
		Get(fmt.Sprintf("%s/api/auth/token", c.baseURL))
	if err != nil {
		return "", fmt.Errorf(c.languagePack.Get("verify_token_request_failed"), err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf(c.languagePack.Get("http_status_code_error"), resp.StatusCode())
	}

	var tokenResp map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &tokenResp); err != nil {
		return "", errors.New(c.languagePack.Get("verify_token_parse_failed"))
	}

	verifyToken, exists := tokenResp["verify_token"]
	if !exists {
		return "", errors.New(c.languagePack.Get("no_token_received"))
	}

	verifyTokenStr, ok := verifyToken.(string)
	if !ok {
		return "", errors.New(c.languagePack.Get("invalid_token_format"))
	}

	return verifyTokenStr, nil
}

// Login authenticates a user
func (c *RestAPIClient) Login(username, password, verifyToken string) (*AuthResponse, error) {
	resp, err := c.client.R().
		SetFormData(map[string]string{
			"username":     username,
			"password":     password,
			"verify_token": verifyToken,
			"language":     getCurrentLanguage(),
		}).
		Post(fmt.Sprintf("%s/api/auth/login", c.baseURL))

	if err != nil {
		return nil, fmt.Errorf(c.languagePack.Get("login_request_failed"), err)
	}

	if err := c.handleAPIResponse(resp, 200, "login_failed"); err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := json.Unmarshal(resp.Body(), &authResp); err != nil {
		return nil, errors.New(c.languagePack.Get("failed_to_parse_login_response"))
	}

	return &authResp, nil
}

// Register creates a new user account
func (c *RestAPIClient) Register(username, password, verifyToken string) (*AuthResponse, error) {
	resp, err := c.client.R().
		SetFormData(map[string]string{
			"username":     username,
			"password":     password,
			"verify_token": verifyToken,
			"language":     getCurrentLanguage(),
		}).
		Post(fmt.Sprintf("%s/api/auth/register", c.baseURL))

	if err != nil {
		return nil, fmt.Errorf(c.languagePack.Get("register_request_failed"), err)
	}

	if resp.StatusCode() == 409 {
		return nil, errors.New(c.languagePack.Get("username_already_exists"))
	}

	if err := c.handleAPIResponse(resp, 200, "register_failed"); err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := json.Unmarshal(resp.Body(), &authResp); err != nil {
		return nil, errors.New(c.languagePack.Get("failed_to_parse_register_response"))
	}

	return &authResp, nil
}

// DeleteUser deletes a user account
func (c *RestAPIClient) DeleteUser(userID, token, password, verifyToken string) error {
	resp, err := c.client.R().
		SetHeaders(map[string]string{
			"user_token": token,
			"user_id":    userID,
		}).
		SetFormData(map[string]string{
			"user_password": password,
			"verify_token":  verifyToken,
			"language":      getCurrentLanguage(),
		}).
		Post(fmt.Sprintf("%s/api/users/%s/delete", c.baseURL, userID))

	if err != nil {
		return fmt.Errorf(c.languagePack.Get("delete_account_request_failed"), err)
	}

	return c.handleAPIResponse(resp, 200, "delete_account_failed")
}

// GetUserProfile retrieves user profile information
func (c *RestAPIClient) GetUserProfile(userID, token, verifyToken string) (*UserInfoResponse, error) {
	resp, err := c.client.R().
		SetHeaders(map[string]string{
			"user_token": token,
			"user_id":    userID,
		}).
		SetQueryParams(map[string]string{
			"verify_token": verifyToken,
			"language":     getCurrentLanguage(),
		}).
		Get(fmt.Sprintf("%s/api/users/%s/profile", c.baseURL, userID))

	if err != nil {
		return nil, fmt.Errorf(c.languagePack.Get("user_info_request_failed"), err)
	}

	if err := c.handleAPIResponse(resp, 200, "user_info_failed"); err != nil {
		return nil, err
	}

	var userInfoResp UserInfoResponse
	if err := json.Unmarshal(resp.Body(), &userInfoResp); err != nil {
		return nil, errors.New(c.languagePack.Get("failed_to_parse_user_info_response"))
	}

	return &userInfoResp, nil
}

// UpdateUserProfile updates a specific user profile field
func (c *RestAPIClient) UpdateUserProfile(userID, token, key, value, verifyToken string) error {
	resp, err := c.client.R().
		SetHeaders(map[string]string{
			"user_token": token,
			"user_id":    userID,
		}).
		SetFormData(map[string]string{
			"verify_token":    verifyToken,
			"user_info_key":   key,
			"user_info_value": value,
			"language":        getCurrentLanguage(),
		}).
		Patch(fmt.Sprintf("%s/api/users/%s/profile", c.baseURL, userID))

	if err != nil {
		return fmt.Errorf(c.languagePack.Get("user_info_change_request_failed"), err)
	}

	return c.handleAPIResponse(resp, 200, "user_info_change_failed")
}

// ChangePassword changes the user's password
func (c *RestAPIClient) ChangePassword(userID, token, oldPassword, newPassword, verifyToken string) error {
	resp, err := c.client.R().
		SetHeaders(map[string]string{
			"user_token": token,
			"user_id":    userID,
		}).
		SetFormData(map[string]string{
			"verify_token": verifyToken,
			"old_password": oldPassword,
			"new_password": newPassword,
			"language":     getCurrentLanguage(),
		}).
		Put(fmt.Sprintf("%s/api/users/%s/password", c.baseURL, userID))

	if err != nil {
		return fmt.Errorf(c.languagePack.Get("password_change_request_failed"), err)
	}

	return c.handleAPIResponse(resp, 200, "password_change_failed")
}

// CheckServerHealth verifies server availability
func (c *RestAPIClient) CheckServerHealth() (bool, error) {
	resp, err := c.client.R().
		SetQueryParam("language", getCurrentLanguage()).
		Get(c.baseURL)
	if err != nil {
		return false, fmt.Errorf(c.languagePack.Get("connection_error"), err)
	}

	if resp.StatusCode() != 200 {
		return false, fmt.Errorf(c.languagePack.Get("http_status_code_error"), resp.StatusCode())
	}

	if resp.String() != "Welcome to Rich Chat!" {
		return false, errors.New(c.languagePack.Get("is_not_rich_chat_server"))
	}

	return true, nil
}

// SetAuthHeaders sets authentication headers for all subsequent requests
func (c *RestAPIClient) SetAuthHeaders(token, userID string) {
	c.client.SetHeader("user_token", token)
	c.client.SetHeader("user_id", userID)
}

// ClearAuthHeaders clears authentication headers
func (c *RestAPIClient) ClearAuthHeaders() {
	c.client.SetHeader("user_token", "")
	c.client.SetHeader("user_id", "")
}

// getCurrentLanguage returns the current language setting
func getCurrentLanguage() string {
	return LANGUAGE
}
