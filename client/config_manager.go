package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	ReadConfig() (map[string]interface{}, error)
	SaveConfig(data map[string]interface{}) error
	GetToken() (string, bool)
	GetUserID() (string, bool)
	SetToken(token string)
	SetUserID(userID string)
	ClearCredentials()
}

// FileConfigManager implements ConfigManager using JSON file storage
type FileConfigManager struct {
	configDir  string
	configFile string
	userData   map[string]interface{}
}

// NewFileConfigManager creates a new file-based config manager
func NewFileConfigManager(configDir, configFile string) *FileConfigManager {
	return &FileConfigManager{
		configDir:  configDir,
		configFile: configFile,
		userData:   make(map[string]interface{}),
	}
}

// ReadConfig loads configuration from file
func (m *FileConfigManager) ReadConfig() (map[string]interface{}, error) {
	// Ensure config directory exists
	if err := m.ensureConfigDir(); err != nil {
		return nil, err
	}

	fullPath := fmt.Sprintf("%s/%s", m.configDir, m.configFile)

	// Create config file if it doesn't exist
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err := m.createEmptyConfig(fullPath); err != nil {
			return nil, err
		}
	}

	// Read and parse config file
	fileData, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read_config_file_error: %w", err)
	}

	if err := json.Unmarshal(fileData, &m.userData); err != nil {
		return nil, fmt.Errorf("parse_config_file_error: %w", err)
	}

	return m.userData, nil
}

// SaveConfig writes configuration to file
func (m *FileConfigManager) SaveConfig(data map[string]interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("map_convert_to_str_error: %w", err)
	}

	fullPath := fmt.Sprintf("%s/%s", m.configDir, m.configFile)
	if err := os.WriteFile(fullPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("write_config_file_error: %w", err)
	}

	m.userData = data
	return nil
}

// GetToken retrieves the stored token
func (m *FileConfigManager) GetToken() (string, bool) {
	token, exists := m.userData["token"]
	if !exists {
		return "", false
	}
	tokenStr, ok := token.(string)
	return tokenStr, ok
}

// GetUserID retrieves the stored user ID
func (m *FileConfigManager) GetUserID() (string, bool) {
	userID, exists := m.userData["user_id"]
	if !exists {
		return "", false
	}
	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// SetToken stores the token
func (m *FileConfigManager) SetToken(token string) {
	m.userData["token"] = token
}

// SetUserID stores the user ID
func (m *FileConfigManager) SetUserID(userID string) {
	m.userData["user_id"] = userID
}

// ClearCredentials removes authentication credentials
func (m *FileConfigManager) ClearCredentials() {
	delete(m.userData, "token")
	delete(m.userData, "user_id")
	m.SaveConfig(m.userData)
}

// ensureConfigDir creates the config directory if it doesn't exist
func (m *FileConfigManager) ensureConfigDir() error {
	dirInfo, err := os.Stat(m.configDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(m.configDir, 0755); err != nil {
				return fmt.Errorf("create_config_dir_error: %w", err)
			}
			return nil
		}
		return fmt.Errorf("stat_config_dir_error: %w", err)
	}

	if !dirInfo.IsDir() {
		return fmt.Errorf("config_dir_is_not_dir")
	}

	return nil
}

// createEmptyConfig creates an empty JSON config file
func (m *FileConfigManager) createEmptyConfig(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create_config_file_error: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString("{}"); err != nil {
		return fmt.Errorf("write_config_file_error: %w", err)
	}

	return nil
}
