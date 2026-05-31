package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

// APIClient defines the interface for API operations
type APIClient interface {
	GetVerifyToken() (string, error)
	Login(username, password, verifyToken string) (*AuthResponse, error)
	Register(username, password, verifyToken string) (*AuthResponse, error)
	DeleteUser(userID, password, verifyToken string) error
	GetUserProfile(userID, verifyToken string) (*UserInfoResponse, error)
	UpdateUserProfile(userID, key, value, verifyToken string) error
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

// GetVerifyToken retrieves a verification token from the server
func (c *RestAPIClient) GetVerifyToken() (string, error) {
	resp, err := c.client.R().Get(fmt.Sprintf("%s/api/auth/token", c.baseURL))
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
		}).
		Post(fmt.Sprintf("%s/api/auth/login", c.baseURL))

	if err != nil {
		return nil, fmt.Errorf(c.languagePack.Get("login_request_failed"), err)
	}

	if resp.StatusCode() == 429 {
		return nil, errors.New(c.languagePack.Get("account_locked_try_later"))
	}

	if resp.StatusCode() != 200 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			msg := errorResp.Message
			if msg == "" {
				msg = errorResp.Error
			}
			if msg != "" {
				return nil, fmt.Errorf(c.languagePack.Get("login_failed"), msg)
			}
		}
		return nil, fmt.Errorf(c.languagePack.Get("login_failed_with_status"), resp.StatusCode())
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
		}).
		Post(fmt.Sprintf("%s/api/auth/register", c.baseURL))

	if err != nil {
		return nil, fmt.Errorf(c.languagePack.Get("register_request_failed"), err)
	}

	if resp.StatusCode() == 409 {
		return nil, errors.New(c.languagePack.Get("username_already_exists"))
	}

	if resp.StatusCode() == 429 {
		return nil, errors.New(c.languagePack.Get("account_locked_try_later"))
	}

	if resp.StatusCode() != 200 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			msg := errorResp.Message
			if msg == "" {
				msg = errorResp.Error
			}
			if msg != "" {
				return nil, fmt.Errorf(c.languagePack.Get("register_failed"), msg)
			}
		}
		return nil, fmt.Errorf(c.languagePack.Get("register_failed_with_status"), resp.StatusCode())
	}

	var authResp AuthResponse
	if err := json.Unmarshal(resp.Body(), &authResp); err != nil {
		return nil, errors.New(c.languagePack.Get("failed_to_parse_register_response"))
	}

	return &authResp, nil
}

// DeleteUser deletes a user account
func (c *RestAPIClient) DeleteUser(userID, password, verifyToken string) error {
	resp, err := c.client.R().
		SetFormData(map[string]string{
			"user_password": password,
			"verify_token":  verifyToken,
		}).
		Post(fmt.Sprintf("%s/api/users/%s/delete", c.baseURL, userID))

	if err != nil {
		return fmt.Errorf(c.languagePack.Get("delete_account_request_failed"), err)
	}

	if resp.StatusCode() == 429 {
		return errors.New(c.languagePack.Get("account_locked_try_later"))
	}

	if resp.StatusCode() != 200 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			msg := errorResp.Message
			if msg == "" {
				msg = errorResp.Error
			}
			if msg != "" {
				return fmt.Errorf(c.languagePack.Get("delete_account_failed"), msg)
			}
		}
		return fmt.Errorf(c.languagePack.Get("delete_account_failed_with_status"), resp.StatusCode())
	}

	return nil
}

// GetUserProfile retrieves user profile information
func (c *RestAPIClient) GetUserProfile(userID, verifyToken string) (*UserInfoResponse, error) {
	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"verify_token": verifyToken,
		}).
		Get(fmt.Sprintf("%s/api/users/%s/profile", c.baseURL, userID))

	if err != nil {
		return nil, fmt.Errorf(c.languagePack.Get("user_info_request_failed"), err)
	}

	if resp.StatusCode() == 429 {
		return nil, errors.New(c.languagePack.Get("account_locked_try_later"))
	}

	if resp.StatusCode() != 200 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			msg := errorResp.Message
			if msg == "" {
				msg = errorResp.Error
			}
			if msg != "" {
				return nil, fmt.Errorf(c.languagePack.Get("user_info_failed"), msg)
			}
		}
		return nil, fmt.Errorf(c.languagePack.Get("user_info_failed_with_status"), resp.StatusCode())
	}

	var userInfoResp UserInfoResponse
	if err := json.Unmarshal(resp.Body(), &userInfoResp); err != nil {
		return nil, errors.New(c.languagePack.Get("failed_to_parse_user_info_response"))
	}

	return &userInfoResp, nil
}

// UpdateUserProfile updates a specific user profile field
func (c *RestAPIClient) UpdateUserProfile(userID, key, value, verifyToken string) error {
	resp, err := c.client.R().
		SetFormData(map[string]string{
			"verify_token":    verifyToken,
			"user_info_key":   key,
			"user_info_value": value,
		}).
		Patch(fmt.Sprintf("%s/api/users/%s/profile", c.baseURL, userID))

	if err != nil {
		return fmt.Errorf(c.languagePack.Get("user_info_change_request_failed"), err)
	}

	if resp.StatusCode() == 429 {
		return errors.New(c.languagePack.Get("account_locked_try_later"))
	}

	if resp.StatusCode() != 200 {
		var errorResp ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			msg := errorResp.Message
			if msg == "" {
				msg = errorResp.Error
			}
			if msg != "" {
				return fmt.Errorf(c.languagePack.Get("user_info_change_failed"), msg)
			}
		}
		return fmt.Errorf(c.languagePack.Get("user_info_change_failed_with_status"), resp.StatusCode())
	}

	return nil
}

// CheckServerHealth verifies server availability
func (c *RestAPIClient) CheckServerHealth() (bool, error) {
	resp, err := c.client.R().Get(c.baseURL)
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
