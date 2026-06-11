package main

import (
	"fmt"
)

// UserService handles user management business logic
type UserService struct {
	apiClient      APIClient
	configMgr      ConfigManager
	tokenExtractor TokenExtractor
	languagePack   *LanguagePackWrapper
}

// NewUserService creates a new user service
func NewUserService(apiClient APIClient, configMgr ConfigManager, tokenExtractor TokenExtractor, languagePack *LanguagePackWrapper) *UserService {
	return &UserService{
		apiClient:      apiClient,
		configMgr:      configMgr,
		tokenExtractor: tokenExtractor,
		languagePack:   languagePack,
	}
}

// DeleteAccount deletes the current user's account
func (s *UserService) DeleteAccount(password string) error {
	userID, exists := s.configMgr.GetUserID()
	if !exists {
		msg := s.languagePack.Get("user_id_not_found")
		return fmt.Errorf("%s", msg)
	}

	token, tokenExists := s.configMgr.GetToken()
	if !tokenExists {
		msg := s.languagePack.Get("user_token_not_found")
		return fmt.Errorf("%s", msg)
	}

	verifyToken, err := s.apiClient.GetVerifyToken()
	if err != nil {
		msg := s.languagePack.Get("getting_verify_token_failed")
		return fmt.Errorf("%s: %w", msg, err)
	}

	if err := s.apiClient.DeleteUser(userID, token, password, verifyToken); err != nil {
		return err
	}

	// Clear local credentials
	s.configMgr.ClearCredentials()
	if err := s.configMgr.SaveConfig(s.getUserData()); err != nil {
		msg := s.languagePack.Get("clear_credentials_failed")
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

// GetProfile retrieves the current user's profile
func (s *UserService) GetProfile() (*UserData, error) {
	userID, exists := s.configMgr.GetUserID()
	if !exists {
		msg := s.languagePack.Get("user_id_not_found")
		return nil, fmt.Errorf("%s", msg)
	}

	token, tokenExists := s.configMgr.GetToken()
	if !tokenExists {
		msg := s.languagePack.Get("user_token_not_found")
		return nil, fmt.Errorf("%s", msg)
	}

	verifyToken, err := s.apiClient.GetVerifyToken()
	if err != nil {
		msg := s.languagePack.Get("getting_verify_token_failed")
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	resp, err := s.apiClient.GetUserProfile(userID, token, verifyToken)
	if err != nil {
		return nil, err
	}

	if resp.Data == nil {
		msg := s.languagePack.Get("failed_to_parse_user_info_response")
		return nil, fmt.Errorf("%s", msg)
	}

	return resp.Data, nil
}

// UpdateProfile updates a specific field in the user's profile
func (s *UserService) UpdateProfile(key, value string) error {
	userID, exists := s.configMgr.GetUserID()
	if !exists {
		msg := s.languagePack.Get("user_id_not_found")
		return fmt.Errorf("%s", msg)
	}

	token, tokenExists := s.configMgr.GetToken()
	if !tokenExists {
		msg := s.languagePack.Get("user_token_not_found")
		return fmt.Errorf("%s", msg)
	}

	verifyToken, err := s.apiClient.GetVerifyToken()
	if err != nil {
		msg := s.languagePack.Get("getting_verify_token_failed")
		return fmt.Errorf("%s: %w", msg, err)
	}

	if err := s.apiClient.UpdateUserProfile(userID, token, key, value, verifyToken); err != nil {
		return err
	}

	return nil
}

// ChangePassword changes the user's password
func (s *UserService) ChangePassword(oldPassword, newPassword string) error {
	userID, exists := s.configMgr.GetUserID()
	if !exists {
		msg := s.languagePack.Get("user_id_not_found")
		return fmt.Errorf("%s", msg)
	}

	token, tokenExists := s.configMgr.GetToken()
	if !tokenExists {
		msg := s.languagePack.Get("user_token_not_found")
		return fmt.Errorf("%s", msg)
	}

	verifyToken, err := s.apiClient.GetVerifyToken()
	if err != nil {
		msg := s.languagePack.Get("getting_verify_token_failed")
		return fmt.Errorf("%s: %w", msg, err)
	}

	if err := s.apiClient.ChangePassword(userID, token, oldPassword, newPassword, verifyToken); err != nil {
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
