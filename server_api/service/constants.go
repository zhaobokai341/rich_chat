package service

import "time"

// Service layer constants
const (
	// JWT configuration
	JWT_EXPIRE_TIME         = time.Hour * 24     // 24 hours (changed from 30 days for security)
	JWT_REFRESH_EXPIRE_TIME = time.Hour * 24 * 7 // 7 days for refresh token

	// RSA/E2EE configuration
	RSA_KEY_SIZE = 4096 // RSA key size in bits (increased from 2048 for security)

	// Rate limiting
	MAX_LOGIN_ATTEMPTS = 5                // Maximum login attempts before lockout
	LOCKOUT_DURATION   = time.Minute * 15 // Account lockout duration

	// User input validation
	ALLOW_MAX_LENGTH_OF_USERNAME = 50
	ALLOW_MAX_LENGTH_OF_NICKNAME = 50
	ALLOW_MAX_LENGTH_OF_PASSWORD = 100
	ALLOW_MAX_LENGTH_OF_BIO      = 500
	ALLOW_MAX_LENGTH_OF_EMAIL    = 100
)
