package main

import (
	"fmt"
)

// UserService handles user management business logic
type UserService struct {
	apiClient      APIClient
	configMgr      ConfigManager
	tokenExtractor TokenExtractor
}

// NewUserService creates a new user service
func NewUserService(apiClient APIClient, configMgr ConfigManager, tokenExtractor TokenExtractor) *UserService {
	return &UserService{
		apiClient:      apiClient,
		configMgr:      configMgr,
		tokenExtractor: tokenExtractor,
	}
}

// DeleteAccount deletes the current user's account
func (s *UserService) DeleteAccount(password string) error {
	userID, exists := s.configMgr.GetUserID()
	if !exists {
		return fmt.Errorf("user_id_not_found")
	}

	verifyToken, err := s.apiClient.GetVerifyToken()
	if err != nil {
		return fmt.Errorf("getting_verify_token_failed: %w", err)
	}

	if err := s.apiClient.DeleteUser(userID, password, verifyToken); err != nil {
		return err
	}

	// Clear local credentials
	s.configMgr.ClearCredentials()
	if err := s.configMgr.SaveConfig(s.getUserData()); err != nil {
		return fmt.Errorf("clear_credentials_failed: %w", err)
	}

	return nil
}

// GetProfile retrieves the current user's profile
func (s *UserService) GetProfile() (*UserData, error) {
	userID, exists := s.configMgr.GetUserID()
	if !exists {
		return nil, fmt.Errorf("user_id_not_found")
	}

	verifyToken, err := s.apiClient.GetVerifyToken()
	if err != nil {
		return nil, fmt.Errorf("getting_verify_token_failed: %w", err)
	}

	resp, err := s.apiClient.GetUserProfile(userID, verifyToken)
	if err != nil {
		return nil, err
	}

	if resp.Data == nil {
		return nil, fmt.Errorf("failed_to_parse_user_info_response")
	}

	return resp.Data, nil
}

// UpdateProfile updates a specific field in the user's profile
func (s *UserService) UpdateProfile(key, value string) error {
	userID, exists := s.configMgr.GetUserID()
	if !exists {
		return fmt.Errorf("user_id_not_found")
	}

	verifyToken, err := s.apiClient.GetVerifyToken()
	if err != nil {
		return fmt.Errorf("getting_verify_token_failed: %w", err)
	}

	if err := s.apiClient.UpdateUserProfile(userID, key, value, verifyToken); err != nil {
		return err
	}

	return nil
}

// getUserData returns current user data from config manager
func (s *UserService) getUserData() map[string]interface{} {
	data, _ := s.configMgr.ReadConfig()
	if data == nil {
		return make(map[string]interface{})
	}
	return data
}
