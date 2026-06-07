package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// TokenExtractor defines the interface for extracting user ID from JWT tokens
type TokenExtractor interface {
	ExtractUserID(token string) (string, error)
}

// JWTTokenExtractor implements TokenExtractor for JWT tokens
type JWTTokenExtractor struct {
	languagePack *LanguagePackWrapper
}

// NewJWTTokenExtractor creates a new JWT token extractor
func NewJWTTokenExtractor() *JWTTokenExtractor {
	return &JWTTokenExtractor{}
}

// NewJWTTokenExtractorWithLanguagePack creates a new JWT token extractor with language pack
func NewJWTTokenExtractorWithLanguagePack(languagePack *LanguagePackWrapper) *JWTTokenExtractor {
	return &JWTTokenExtractor{
		languagePack: languagePack,
	}
}

// ExtractUserID extracts the user_id from a JWT token
func (e *JWTTokenExtractor) ExtractUserID(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		msg := getErrorMessage(e.languagePack, "invalid_token_format")
		return "", fmt.Errorf("%s", msg)
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		msg := getErrorMessage(e.languagePack, "failed_to_decode_token_payload")
		return "", fmt.Errorf("%s: %v", msg, err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		msg := getErrorMessage(e.languagePack, "failed_to_parse_token_claims")
		return "", fmt.Errorf("%s: %v", msg, err)
	}

	userID, ok := claims["user_id"]
	if !ok {
		msg := getErrorMessage(e.languagePack, "user_id_not_found_in_token")
		return "", fmt.Errorf("%s", msg)
	}

	// Convert to string
	switch v := userID.(type) {
	case float64:
		return fmt.Sprintf("%.0f", v), nil
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// getErrorMessage safely gets error message from language pack, falling back to key if languagePack is nil
func getErrorMessage(lp *LanguagePackWrapper, key string) string {
	if lp == nil {
		return key // Return the key itself if language pack is not available
	}
	return lp.Get(key)
}
