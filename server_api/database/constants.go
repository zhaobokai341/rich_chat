package database

import "time"

// Cache TTL constants for database layer
const (
	CACHE_TTL                = 300 // Cache TTL in seconds (5 minutes)
	CACHE_NULL_TTL           = 60  // Cache TTL for null values in seconds (1 minute)
	CACHE_USER_PROFILE_TTL   = 600 // User profile cache TTL (10 minutes)
	CACHE_USER_BASIC_TTL     = 900 // User basic info cache TTL (15 minutes)
	CACHE_USER_EXISTS_TTL    = 300 // User existence cache TTL (5 minutes)
	CACHE_IP_BLOCKED_TTL     = 60  // IP blocked status cache TTL (1 minute)
	CACHE_IP_NOT_BLOCKED_TTL = 60  // IP not blocked negative cache TTL (1 minute)
	CACHE_PASSWORD_HASH_TTL  = 120 // Password hash cache TTL (2 minutes) - short TTL for security
)

// Rate limiting constants
const (
	MAX_LOGIN_ATTEMPTS        = 5                // Maximum login attempts before lockout
	LOCKOUT_DURATION          = time.Minute * 15 // Account lockout duration after max attempts
	IP_LIMIT_TIME             = time.Minute * 10 // Time window for IP rate limiting
	IP_LIMIT_VISIT_TIMES      = 1000             // Maximum visits per IP within IP_LIMIT_TIME
	IP_LIMIT_LOCKOUT_DURATION = time.Minute * 10 // IP lockout duration
)
