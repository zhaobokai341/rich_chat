package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileConfigManager_ReadConfig(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string) error
		expectedError bool
	}{
		{
			name: "read existing config",
			setupFunc: func(dir string) error {
				configPath := filepath.Join(dir, "config.json")
				return os.WriteFile(configPath, []byte(`{"token":"test-token","user_id":"1"}`), 0644)
			},
			expectedError: false,
		},
		{
			name: "read non-existent config creates empty",
			setupFunc: func(dir string) error {
				// Don't create file
				return nil
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			manager := NewFileConfigManager(tempDir, "config.json")

			if tt.setupFunc != nil {
				if err := tt.setupFunc(tempDir); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			data, err := manager.ReadConfig()
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if data == nil {
					t.Errorf("Expected config data but got nil")
				}
			}
		})
	}
}

func TestFileConfigManager_SaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewFileConfigManager(tempDir, "config.json")

	testData := map[string]interface{}{
		"token":   "test-token-123",
		"user_id": "42",
	}

	err := manager.SaveConfig(testData)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tempDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}

	// Verify data can be read back
	readData, err := manager.ReadConfig()
	if err != nil {
		t.Errorf("Failed to read saved config: %v", err)
	}

	if token, exists := readData["token"]; !exists || token != "test-token-123" {
		t.Errorf("Token not saved correctly")
	}
	if userID, exists := readData["user_id"]; !exists || userID != "42" {
		t.Errorf("UserID not saved correctly")
	}
}

func TestFileConfigManager_GetToken(t *testing.T) {
	tests := []struct {
		name        string
		setupData   map[string]interface{}
		expectToken string
		expectExist bool
	}{
		{
			name: "get existing token",
			setupData: map[string]interface{}{
				"token": "my-token",
			},
			expectToken: "my-token",
			expectExist: true,
		},
		{
			name:        "get non-existent token",
			setupData:   map[string]interface{}{},
			expectToken: "",
			expectExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConfigManager()
			for k, v := range tt.setupData {
				manager.data[k] = v
			}

			token, exists := manager.GetToken()
			if exists != tt.expectExist {
				t.Errorf("Expected existence %v but got %v", tt.expectExist, exists)
			}
			if token != tt.expectToken {
				t.Errorf("Expected token '%s' but got '%s'", tt.expectToken, token)
			}
		})
	}
}

func TestFileConfigManager_GetUserID(t *testing.T) {
	tests := []struct {
		name        string
		setupData   map[string]interface{}
		expectID    string
		expectExist bool
	}{
		{
			name: "get existing user ID",
			setupData: map[string]interface{}{
				"user_id": "123",
			},
			expectID:    "123",
			expectExist: true,
		},
		{
			name:        "get non-existent user ID",
			setupData:   map[string]interface{}{},
			expectID:    "",
			expectExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMockConfigManager()
			for k, v := range tt.setupData {
				manager.data[k] = v
			}

			userID, exists := manager.GetUserID()
			if exists != tt.expectExist {
				t.Errorf("Expected existence %v but got %v", tt.expectExist, exists)
			}
			if userID != tt.expectID {
				t.Errorf("Expected userID '%s' but got '%s'", tt.expectID, userID)
			}
		})
	}
}

func TestFileConfigManager_SetToken(t *testing.T) {
	manager := NewMockConfigManager()
	manager.SetToken("new-token")

	token, exists := manager.GetToken()
	if !exists {
		t.Errorf("Token should exist after setting")
	}
	if token != "new-token" {
		t.Errorf("Expected token 'new-token' but got '%s'", token)
	}
}

func TestFileConfigManager_SetUserID(t *testing.T) {
	manager := NewMockConfigManager()
	manager.SetUserID("999")

	userID, exists := manager.GetUserID()
	if !exists {
		t.Errorf("UserID should exist after setting")
	}
	if userID != "999" {
		t.Errorf("Expected userID '999' but got '%s'", userID)
	}
}

func TestFileConfigManager_ClearCredentials(t *testing.T) {
	manager := NewMockConfigManager()
	manager.SetToken("test-token")
	manager.SetUserID("1")

	manager.ClearCredentials()

	if _, exists := manager.GetToken(); exists {
		t.Errorf("Token should be cleared")
	}
	if _, exists := manager.GetUserID(); exists {
		t.Errorf("UserID should be cleared")
	}
}
