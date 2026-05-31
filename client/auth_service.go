package main

import (
	"fmt"
)

// AuthService handles authentication business logic
type AuthService struct {
	apiClient    APIClient
	configMgr    ConfigManager
	tokenExtractor TokenExtractor
}

// NewAuthService creates a new authentication service
func NewAuthService(apiClient APIClient, configMgr ConfigManager, tokenExtractor TokenExtractor) *AuthService {
	return &AuthService{
		apiClient:      apiClient,
		configMgr:      configMgr,
		tokenExtractor: tokenExtractor,
	}
}

// Login authenticates a user and saves credentials
func (s *AuthService) Login(username, password string) error {
	// Get verification token
	verifyToken, err := s.apiClient.GetVerifyToken()
	if err != nil {
		return fmt.Errorf("getting_verify_token_failed: %w", err)
	}

	// Perform login
	authResp, err := s.apiClient.Login(username, password, verifyToken)
	if err != nil {
		return err
	}

	// Save credentials
	s.configMgr.SetToken(authResp.UserToken)
	userIDStr := fmt.Sprintf("%d", authResp.UserID)
	s.configMgr.SetUserID(userIDStr)

	if err := s.configMgr.SaveConfig(s.getUserData()); err != nil {
		return fmt.Errorf("save_credentials_failed: %w", err)
	}

	return nil
}

// Register creates a new user account and saves credentials
func (s *AuthService) Register(username, password string) error {
	// Get verification token
	verifyToken, err := s.apiClient.GetVerifyToken()
	if err != nil {
		return fmt.Errorf("getting_verify_token_failed: %w", err)
	}

	// Perform registration
	authResp, err := s.apiClient.Register(username, password, verifyToken)
	if err != nil {
		return err
	}

	// Save credentials
	s.configMgr.SetToken(authResp.UserToken)
	userIDStr := fmt.Sprintf("%d", authResp.UserID)
	s.configMgr.SetUserID(userIDStr)

	if err := s.configMgr.SaveConfig(s.getUserData()); err != nil {
		return fmt.Errorf("save_credentials_failed: %w", err)
	}

	return nil
}

// Logout clears authentication credentials
func (s *AuthService) Logout() error {
	s.configMgr.ClearCredentials()
	if err := s.configMgr.SaveConfig(s.getUserData()); err != nil {
		return fmt.Errorf("clear_credentials_failed: %w", err)
	}
	return nil
}

// IsAuthenticated checks if user is currently authenticated
func (s *AuthService) IsAuthenticated() bool {
	_, tokenExists := s.configMgr.GetToken()
	_, userIDExists := s.configMgr.GetUserID()
	return tokenExists && userIDExists
}

// GetCredentials retrieves current authentication credentials
func (s *AuthService) GetCredentials() (token string, userID string, err error) {
	token, tokenOk := s.configMgr.GetToken()
	userID, userIDOk := s.configMgr.GetUserID()

	if !tokenOk || !userIDOk {
		return "", "", fmt.Errorf("user_id_or_token_not_string")
	}

	return token, userID, nil
}

// getUserData returns current user data from config manager
func (s *AuthService) getUserData() map[string]interface{} {
	data, _ := s.configMgr.ReadConfig()
	if data == nil {
		return make(map[string]interface{})
	}
	return data
}
