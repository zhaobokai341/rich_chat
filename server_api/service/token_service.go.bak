package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"

	"rich_chat/server_api/database"
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
	Issuer     string
}

// TokenServiceImpl implements TokenService
type TokenServiceImpl struct {
	jwtConfig JWTConfig
	tokenRepo database.TokenRepository
}

// NewTokenService creates a new token service
func NewTokenService(jwtConfig JWTConfig, tokenRepo database.TokenRepository) *TokenServiceImpl {
	return &TokenServiceImpl{
		jwtConfig: jwtConfig,
		tokenRepo: tokenRepo,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for a user
func (s *TokenServiceImpl) GenerateJWT(userID int, expiration time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiration)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.jwtConfig.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtConfig.Secret))
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to generate JWT token")
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id": userID,
	}).Debug("JWT token generated successfully")

	return tokenString, nil
}

// GenerateVerificationToken generates a random verification token
func (s *TokenServiceImpl) GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to generate verification token")
		return "", fmt.Errorf("failed to generate verification token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// StoreVerificationToken stores a verification token
func (s *TokenServiceImpl) StoreVerificationToken(token string, ttl time.Duration) error {
	err := s.tokenRepo.StoreVerifyToken(token, ttl)
	if err != nil {
		log.WithFields(log.Fields{
			"token": token,
			"error": err.Error(),
		}).Error("Failed to store verification token")
		return fmt.Errorf("failed to store verification token: %w", err)
	}

	log.WithFields(log.Fields{
		"token": token,
		"ttl":   ttl,
	}).Debug("Verification token stored successfully")

	return nil
}

// ValidateAndConsumeToken validates and consumes a verification token
func (s *TokenServiceImpl) ValidateAndConsumeToken(token string) error {
	if token == "" {
		return ErrInvalidToken
	}

	valid, err := s.tokenRepo.VerifyAndConsumeToken(token)
	if err != nil {
		log.WithFields(log.Fields{
			"token": token,
			"error": err.Error(),
		}).Error("Failed to verify token")
		return fmt.Errorf("failed to verify token: %w", err)
	}

	if !valid {
		log.WithFields(log.Fields{
			"token": token,
		}).Warning("Invalid or expired verification token")
		return ErrInvalidToken
	}

	return nil
}
