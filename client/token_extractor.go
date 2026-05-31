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
type JWTTokenExtractor struct{}

// NewJWTTokenExtractor creates a new JWT token extractor
func NewJWTTokenExtractor() *JWTTokenExtractor {
	return &JWTTokenExtractor{}
}

// ExtractUserID extracts the user_id from a JWT token
func (e *JWTTokenExtractor) ExtractUserID(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode token payload: %v", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse token claims: %v", err)
	}

	userID, ok := claims["user_id"]
	if !ok {
		return "", fmt.Errorf("user_id not found in token")
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
