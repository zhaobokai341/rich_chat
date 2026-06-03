package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileConfigManager(t *testing.T) {
	configDir := "/tmp/test_config"
	configFile := "config.json"

	manager := NewFileConfigManager(configDir, configFile)

	assert.NotNil(t, manager)
	assert.Equal(t, configDir, manager.configDir)
	assert.Equal(t, configFile, manager.configFile)
	assert.NotNil(t, manager.userData)
}

func TestFileConfigManager_ReadConfig(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "config_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		setupFunc     func(string) error
		expectedError bool
		checkData     func(map[string]interface{})
	}{
		{
			name: "read existing config",
			setupFunc: func(dir string) error {
				configPath := filepath.Join(dir, "test.json")
				return os.WriteFile(configPath, []byte(`{"token":"test-token","user_id":"123"}`), 0644)
			},
			expectedError: false,
			checkData: func(data map[string]interface{}) {
				assert.Equal(t, "test-token", data["token"])
				assert.Equal(t, "123", data["user_id"])
			},
		},
		{
			name: "create new config if not exists",
			setupFunc: func(dir string) error {
				// Don't create file - should be created automatically
				return nil
			},
			expectedError: false,
			checkData: func(data map[string]interface{}) {
				assert.Empty(t, data)
			},
		},
		{
			name: "invalid JSON format",
			setupFunc: func(dir string) error {
				configPath := filepath.Join(dir, "test.json")
				return os.WriteFile(configPath, []byte(`{invalid json`), 0644)
			},
			expectedError: true,
			checkData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := filepath.Join(tempDir, tt.name)
			err := os.MkdirAll(testDir, 0755)
			assert.NoError(t, err)

			if tt.setupFunc != nil {
				err := tt.setupFunc(testDir)
				assert.NoError(t, err)
			}

			manager := NewFileConfigManager(testDir, "test.json")
			data, err := manager.ReadConfig()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkData != nil {
					tt.checkData(data)
				}
			}
		})
	}
}

func TestFileConfigManager_SaveConfig(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "config_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := NewFileConfigManager(tempDir, "test.json")

	testData := map[string]interface{}{
		"token":    "test-token-123",
		"user_id":  "456",
		"username": "testuser",
	}

	err = manager.SaveConfig(testData)
	assert.NoError(t, err)

	// Verify file was created and contains correct data
	configPath := filepath.Join(tempDir, "test.json")
	fileData, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Contains(t, string(fileData), "test-token-123")
	assert.Contains(t, string(fileData), "456")

	// Verify in-memory data was updated
	assert.Equal(t, "test-token-123", manager.userData["token"])
	assert.Equal(t, "456", manager.userData["user_id"])
}

func TestFileConfigManager_GetToken(t *testing.T) {
	manager := NewFileConfigManager("/tmp/test", "test.json")

	tests := []struct {
		name           string
		setupData      map[string]interface{}
		expectedToken  string
		expectedExists bool
	}{
		{
			name: "token exists",
			setupData: map[string]interface{}{
				"token": "valid-token",
			},
			expectedToken:  "valid-token",
			expectedExists: true,
		},
		{
			name:           "token does not exist",
			setupData:      map[string]interface{}{},
			expectedToken:  "",
			expectedExists: false,
		},
		{
			name: "token is not string",
			setupData: map[string]interface{}{
				"token": 12345,
			},
			expectedToken:  "",
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.userData = tt.setupData
			token, exists := manager.GetToken()

			assert.Equal(t, tt.expectedToken, token)
			assert.Equal(t, tt.expectedExists, exists)
		})
	}
}

func TestFileConfigManager_GetUserID(t *testing.T) {
	manager := NewFileConfigManager("/tmp/test", "test.json")

	tests := []struct {
		name            string
		setupData       map[string]interface{}
		expectedUserID  string
		expectedExists  bool
	}{
		{
			name: "user_id exists",
			setupData: map[string]interface{}{
				"user_id": "123",
			},
			expectedUserID: "123",
			expectedExists: true,
		},
		{
			name:           "user_id does not exist",
			setupData:      map[string]interface{}{},
			expectedUserID: "",
			expectedExists: false,
		},
		{
			name: "user_id is not string",
			setupData: map[string]interface{}{
				"user_id": 999,
			},
			expectedUserID: "",
			expectedExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.userData = tt.setupData
			userID, exists := manager.GetUserID()

			assert.Equal(t, tt.expectedUserID, userID)
			assert.Equal(t, tt.expectedExists, exists)
		})
	}
}

func TestFileConfigManager_SetToken(t *testing.T) {
	manager := NewFileConfigManager("/tmp/test", "test.json")

	manager.SetToken("new-token-abc")

	assert.Equal(t, "new-token-abc", manager.userData["token"])
}

func TestFileConfigManager_SetUserID(t *testing.T) {
	manager := NewFileConfigManager("/tmp/test", "test.json")

	manager.SetUserID("789")

	assert.Equal(t, "789", manager.userData["user_id"])
}

func TestFileConfigManager_ClearCredentials(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "config_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := NewFileConfigManager(tempDir, "test.json")

	// Set initial credentials
	manager.userData = map[string]interface{}{
		"token":    "old-token",
		"user_id":  "999",
		"username": "testuser",
	}

	manager.ClearCredentials()

	// Verify credentials were removed
	_, tokenExists := manager.GetToken()
	_, userIDExists := manager.GetUserID()
	assert.False(t, tokenExists)
	assert.False(t, userIDExists)

	// Verify other data remains
	assert.Equal(t, "testuser", manager.userData["username"])
}

func TestFileConfigManager_ensureConfigDir(t *testing.T) {
	// Create temporary parent directory
	tempParent, err := os.MkdirTemp("", "config_parent")
	assert.NoError(t, err)
	defer os.RemoveAll(tempParent)

	tests := []struct {
		name          string
		dirPath       string
		setupFunc     func(string) error
		expectedError bool
	}{
		{
			name:    "directory does not exist - create it",
			dirPath: filepath.Join(tempParent, "new_dir"),
			setupFunc: func(path string) error {
				return nil // Don't create anything
			},
			expectedError: false,
		},
		{
			name:    "directory already exists",
			dirPath: filepath.Join(tempParent, "existing_dir"),
			setupFunc: func(path string) error {
				return os.MkdirAll(path, 0755)
			},
			expectedError: false,
		},
		{
			name:    "path is a file not directory",
			dirPath: filepath.Join(tempParent, "not_a_dir"),
			setupFunc: func(path string) error {
				return os.WriteFile(path, []byte("test"), 0644)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				err := tt.setupFunc(tt.dirPath)
				assert.NoError(t, err)
			}

			manager := &FileConfigManager{
				configDir: tt.dirPath,
				userData:  make(map[string]interface{}),
			}

			err := manager.ensureConfigDir()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify directory exists
				info, err := os.Stat(tt.dirPath)
				assert.NoError(t, err)
				assert.True(t, info.IsDir())
			}
		})
	}
}

func TestFileConfigManager_createEmptyConfig(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "config_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	manager := NewFileConfigManager(tempDir, "test.json")

	configPath := filepath.Join(tempDir, "empty.json")
	err = manager.createEmptyConfig(configPath)

	assert.NoError(t, err)

	// Verify file was created with empty JSON
	fileData, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(fileData))
}
