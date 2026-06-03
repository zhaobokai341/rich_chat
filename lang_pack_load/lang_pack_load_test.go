package lang_pack_load

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLanguagePack(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		language     string
		expectedPath string
	}{
		{
			name:         "client main.json with zh language",
			file:         "client/main.json",
			language:     "zh",
			expectedPath: "../lang_pack/client/main.json",
		},
		{
			name:         "server_api main.json with en language",
			file:         "server_api/main.json",
			language:     "en",
			expectedPath: "../lang_pack/server_api/main.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lp := NewLanguagePack(tt.file, tt.language)

			assert.NotNil(t, lp)
			assert.Equal(t, tt.expectedPath, lp.file)
			assert.Equal(t, tt.language, lp.language)
			assert.NotNil(t, lp.data)
		})
	}
}

func TestLanguagePack_Load(t *testing.T) {
	// Create temporary directory and file for testing
	tempDir, err := os.MkdirTemp("", "lang_pack_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test JSON file
	testData := map[string]map[string]string{
		"greeting": {
			"zh": "你好",
			"en": "Hello",
		},
		"farewell": {
			"zh": "再见",
			"en": "Goodbye",
		},
	}

	jsonFile := filepath.Join(tempDir, "test.json")
	jsonContent, _ := json.Marshal(testData)
	err = os.WriteFile(jsonFile, jsonContent, 0644)
	assert.NoError(t, err)

	// Test successful load
	lp := &LanguagePack{
		file:     jsonFile,
		language: "zh",
		data:     make(map[string]map[string]string),
	}

	lp.Load()

	assert.NotNil(t, lp.data)
	assert.Contains(t, lp.data, "greeting")
	assert.Contains(t, lp.data, "farewell")
	assert.Equal(t, "你好", lp.data["greeting"]["zh"])
	assert.Equal(t, "Hello", lp.data["greeting"]["en"])
}

func TestLanguagePack_Load_FileNotFound(t *testing.T) {
	// Test loading from non-existent file
	lp := &LanguagePack{
		file:     "/nonexistent/path/file.json",
		language: "zh",
		data:     make(map[string]map[string]string),
	}

	// Should not panic even if file doesn't exist
	lp.Load()

	// Data should be empty or nil
	assert.Empty(t, lp.data)
}

func TestLanguagePack_Load_InvalidJSON(t *testing.T) {
	// Create temporary file with invalid JSON
	tempFile, err := os.CreateTemp("", "invalid_json_*.json")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString("this is not valid json {{{")
	assert.NoError(t, err)
	tempFile.Close()

	lp := &LanguagePack{
		file:     tempFile.Name(),
		language: "zh",
		data:     make(map[string]map[string]string),
	}

	// Should not panic even with invalid JSON
	lp.Load()

	// Data should be empty
	assert.Empty(t, lp.data)
}

func TestLanguagePack_G(t *testing.T) {
	// Create a LanguagePack with test data
	lp := &LanguagePack{
		file:     "test.json",
		language: "zh",
		data: map[string]map[string]string{
			"greeting": {
				"zh": "你好",
				"en": "Hello",
			},
			"farewell": {
				"zh": "再见",
				"en": "Goodbye",
			},
			"empty_lang": {
				"en": "Only English",
			},
		},
	}

	tests := []struct {
		name     string
		key      string
		language string
		expected string
	}{
		{
			name:     "existing key in Chinese",
			key:      "greeting",
			language: "zh",
			expected: "你好",
		},
		{
			name:     "existing key in English",
			key:      "greeting",
			language: "en",
			expected: "Hello",
		},
		{
			name:     "non-existent key",
			key:      "nonexistent",
			language: "zh",
			expected: "",
		},
		{
			name:     "key exists but language does not",
			key:      "empty_lang",
			language: "zh",
			expected: "",
		},
		{
			name:     "empty key",
			key:      "",
			language: "zh",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the language for this test
			lp.language = tt.language

			result := lp.G(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLanguagePack_MultipleLanguages(t *testing.T) {
	// Test switching between languages
	lp := &LanguagePack{
		file:     "test.json",
		language: "zh",
		data: map[string]map[string]string{
			"welcome": {
				"zh": "欢迎",
				"en": "Welcome",
				"ja": "ようこそ",
			},
		},
	}

	// Test Chinese
	lp.language = "zh"
	assert.Equal(t, "欢迎", lp.G("welcome"))

	// Test English
	lp.language = "en"
	assert.Equal(t, "Welcome", lp.G("welcome"))

	// Test Japanese
	lp.language = "ja"
	assert.Equal(t, "ようこそ", lp.G("welcome"))

	// Test unsupported language
	lp.language = "fr"
	assert.Equal(t, "", lp.G("welcome"))
}

func TestLanguagePack_ComplexStructure(t *testing.T) {
	// Test with more complex nested structure
	lp := &LanguagePack{
		file:     "test.json",
		language: "zh",
		data: map[string]map[string]string{
			"error_messages": {
				"zh": "错误消息",
				"en": "Error message",
			},
			"success_messages": {
				"zh": "成功消息",
				"en": "Success message",
			},
			"validation": {
				"zh": "验证",
				"en": "Validation",
			},
		},
	}

	// Verify all keys are accessible
	assert.Equal(t, "错误消息", lp.G("error_messages"))
	assert.Equal(t, "成功消息", lp.G("success_messages"))
	assert.Equal(t, "验证", lp.G("validation"))

	// Switch to English
	lp.language = "en"
	assert.Equal(t, "Error message", lp.G("error_messages"))
	assert.Equal(t, "Success message", lp.G("success_messages"))
	assert.Equal(t, "Validation", lp.G("validation"))
}

func TestLanguagePack_EmptyData(t *testing.T) {
	// Test with empty data map
	lp := &LanguagePack{
		file:     "test.json",
		language: "zh",
		data:     make(map[string]map[string]string),
	}

	// All lookups should return empty string
	assert.Equal(t, "", lp.G("any_key"))
	assert.Equal(t, "", lp.G(""))
}
