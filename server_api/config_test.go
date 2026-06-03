package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable exists",
			key:          "TEST_VAR",
			value:        "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "environment variable does not exist",
			key:          "NONEXISTENT_VAR",
			value:        "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
		{
			name:         "empty environment variable uses default",
			key:          "EMPTY_VAR",
			value:        "",
			defaultValue: "fallback",
			expected:     "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue int
		expected     int
	}{
		{
			name:         "valid integer environment variable",
			key:          "TEST_INT",
			value:        "42",
			defaultValue: 100,
			expected:     42,
		},
		{
			name:         "environment variable does not exist",
			key:          "NONEXISTENT_INT",
			value:        "",
			defaultValue: 5432,
			expected:     5432,
		},
		{
			name:         "invalid integer uses default",
			key:          "INVALID_INT",
			value:        "not_a_number",
			defaultValue: 6379,
			expected:     6379,
		},
		{
			name:         "zero value",
			key:          "ZERO_INT",
			value:        "0",
			defaultValue: 10,
			expected:     0,
		},
		{
			name:         "negative value",
			key:          "NEGATIVE_INT",
			value:        "-5",
			defaultValue: 10,
			expected:     -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvInt(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Save original environment
	originalJWTSecret := os.Getenv("JWT_SECRET")
	originalDBHost := os.Getenv("DB_HOST")
	originalDBPort := os.Getenv("DB_PORT")
	originalRedisHost := os.Getenv("REDIS_HOST")
	originalRedisPort := os.Getenv("REDIS_PORT")

	// Restore after test
	defer func() {
		if originalJWTSecret != "" {
			os.Setenv("JWT_SECRET", originalJWTSecret)
		} else {
			os.Unsetenv("JWT_SECRET")
		}
		if originalDBHost != "" {
			os.Setenv("DB_HOST", originalDBHost)
		} else {
			os.Unsetenv("DB_HOST")
		}
		if originalDBPort != "" {
			os.Setenv("DB_PORT", originalDBPort)
		} else {
			os.Unsetenv("DB_PORT")
		}
		if originalRedisHost != "" {
			os.Setenv("REDIS_HOST", originalRedisHost)
		} else {
			os.Unsetenv("REDIS_HOST")
		}
		if originalRedisPort != "" {
			os.Setenv("REDIS_PORT", originalRedisPort)
		} else {
			os.Unsetenv("REDIS_PORT")
		}
	}()

	// Set test environment variables
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("DB_HOST", "test-db-host")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("REDIS_HOST", "test-redis-host")
	os.Setenv("REDIS_PORT", "6380")

	// Load config
	LoadConfig()

	// Verify loaded values
	assert.Equal(t, "test-jwt-secret", JWT_SECRET)
	assert.Equal(t, "test-db-host", DB_HOST)
	assert.Equal(t, 5433, DB_PORT)
	assert.Equal(t, "test-redis-host", REDIS_HOST)
	assert.Equal(t, 6380, REDIS_PORT)
}

func TestLoadConfigWithDefaults(t *testing.T) {
	// Unset all environment variables to test defaults
	envVars := []string{
		"JWT_SECRET", "AUTH_USERNAME", "AUTH_PASSWORD",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASS", "DB_NAME", "DB_SSL",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
	}

	originalValues := make(map[string]string)
	for _, envVar := range envVars {
		originalValues[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}

	// Backup and remove .env file if it exists
	originalEnvContent := []byte{}
	envFileExists := false
	if content, err := os.ReadFile(".env"); err == nil {
		originalEnvContent = content
		envFileExists = true
		os.Remove(".env")
	}

	// Restore after test
	defer func() {
		// Restore environment variables
		for envVar, value := range originalValues {
			if value != "" {
				os.Setenv(envVar, value)
			} else {
				os.Unsetenv(envVar)
			}
		}

		// Restore .env file if it existed
		if envFileExists {
			if err := os.WriteFile(".env", originalEnvContent, 0644); err != nil {
				t.Logf("Warning: failed to restore .env file: %v", err)
			}
		}
	}()

	// Load config with no environment variables set
	LoadConfig()

	// Verify default values
	assert.Equal(t, "your-secret-key-change-in-production", JWT_SECRET)
	assert.Equal(t, "admin", AUTH_USERNAME)
	assert.Equal(t, "change-this-password", AUTH_PASSWORD)
	assert.Equal(t, "localhost", DB_HOST)
	assert.Equal(t, 5432, DB_PORT)
	assert.Equal(t, "postgres", DB_USER)
	assert.Equal(t, "", DB_PASS)
	assert.Equal(t, "rich_chat", DB_NAME)
	assert.Equal(t, "require", DB_SSL)
	assert.Equal(t, "localhost", REDIS_HOST)
	assert.Equal(t, 6379, REDIS_PORT)
	assert.Equal(t, "", REDIS_PASSWORD)
	assert.Equal(t, 0, REDIS_DB)
}
