package main

import (
	"fmt"
)

// AuthService handles authentication business logic
type AuthService struct {
	apiClient      APIClient
	configMgr      ConfigManager
	tokenExtractor TokenExtractor
	languagePack   *LanguagePackWrapper
}

// NewAuthService creates a new authentication service
func NewAuthService(apiClient APIClient, configMgr ConfigManager, tokenExtractor TokenExtractor, languagePack *LanguagePackWrapper) *AuthService {
	return &AuthService{
		apiClient:      apiClient,
		configMgr:      configMgr,
		tokenExtractor: tokenExtractor,
		languagePack:   languagePack,
	}
}

// Login authenticates a user and saves credentials
func (s *AuthService) Login(username, password string) error {
	// Get verification token
	verifyToken, err := s.apiClient.GetVerifyToken()
	if err != nil {
		msg := s.languagePack.Get("getting_verify_token_failed")
		return fmt.Errorf("%s: %w", msg, err)
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
		msg := s.languagePack.Get("save_credentials_failed")
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

// Register creates a new user account and saves credentials
func (s *AuthService) Register(username, password string) error {
	// Get verification token
	verifyToken, err := s.apiClient.GetVerifyToken()
	if err != nil {
		msg := s.languagePack.Get("getting_verify_token_failed")
		return fmt.Errorf("%s: %w", msg, err)
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
		msg := s.languagePack.Get("save_credentials_failed")
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

// Logout clears authentication credentials
func (s *AuthService) Logout() error {
	s.configMgr.ClearCredentials()
	if err := s.configMgr.SaveConfig(s.getUserData()); err != nil {
		msg := s.languagePack.Get("clear_credentials_failed")
		return fmt.Errorf("%s: %w", msg, err)
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
		msg := s.languagePack.Get("user_id_or_token_not_string")
		return "", "", fmt.Errorf("%s", msg)
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
