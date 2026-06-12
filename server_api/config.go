package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Configuration variables
const (
	WEB_PORT         = ":2316"                             // http port
	DEFAULT_LANGUAGE = "zh"                                // default language (zh/en)
	LANGUAGE_PACK    = "../lang_pack/server_api/main.json" // language pack file path
	HTTPS_FORCE      = false                               // force https (RECOMMENDED SET TO TRUE ON PRODUCTION)
	VERSION          = "1.0.0"                             // product version, DO NOT CHANGE
)

var (
	JWT_SECRET    string // JWT secret - loaded from environment
	AUTH_USERNAME string // HTTP Basic Auth username - loaded from environment
	AUTH_PASSWORD string // HTTP Basic Auth password - loaded from environment
)

// Database configuration variables
var (
	DB_HOST string // database host - loaded from environment
	DB_PORT int    // database port - loaded from environment
	DB_USER string // database user - loaded from environment
	DB_PASS string // database password - loaded from environment
	DB_NAME string // database name - loaded from environment
	DB_SSL  string // database ssl mode - loaded from environment
)

// Redis configuration variables
var (
	REDIS_HOST     string // Redis host - loaded from environment
	REDIS_PORT     int    // Redis port - loaded from environment
	REDIS_PASSWORD string // Redis password - loaded from environment
	REDIS_DB       int    // Redis database number - loaded from environment
)

// JWT configuration
const (
	JWT_EXPIRE_TIME          = time.Hour * 24     // 24 hours (changed from 30 days for security)
	JWT_REFRESH_EXPIRE_TIME  = time.Hour * 24 * 7 // 7 days for refresh token
	VERIFY_TOKEN_EXPIRE_TIME = time.Minute * 5    // Verification token expire time (5 minutes)
)

// RSA/E2EE configuration
const (
	RSA_KEY_SIZE = 4096 // RSA key size in bits (increased from 2048 for security)
)

// Rate limiting configuration
const (
	MAX_LOGIN_ATTEMPTS        = 5                // Maximum login attempts before lockout
	LOCKOUT_DURATION          = time.Minute * 15 // Account lockout duration after max attempts
	IP_LIMIT_TIME             = time.Minute * 10 // Time window for IP rate limiting
	IP_LIMIT_VISIT_TIMES      = 1000             // Maximum visits per IP within IP_LIMIT_TIME
	IP_LIMIT_LOCKOUT_DURATION = time.Minute * 10 // IP lockout duration
)

// Cache configuration
const (
	CACHE_TTL                = 300 // Cache TTL in seconds (5 minutes)
	CACHE_NULL_TTL           = 60  // Cache TTL for null values in seconds (1 minute)
	CACHE_USER_PROFILE_TTL   = 600 // User profile cache TTL (10 minutes)
	CACHE_USER_BASIC_TTL     = 900 // User basic info cache TTL (15 minutes)
	CACHE_USER_EXISTS_TTL    = 300 // User existence cache TTL (5 minutes)
	CACHE_IP_BLOCKED_TTL     = 60  // IP blocked status cache TTL (1 minute)
	CACHE_IP_NOT_BLOCKED_TTL = 60  // IP not blocked negative cache TTL (1 minute)
)

// User input validation constants
const (
	ALLOW_MAX_LENGTH_OF_USERNAME = 50  // max length of username
	ALLOW_MAX_LENGTH_OF_NICKNAME = 50  // max length of nickname
	ALLOW_MAX_LENGTH_OF_PASSWORD = 100 // max length of password
	ALLOW_MAX_LENGTH_OF_BIO      = 500 // max length of bio
	ALLOW_MAX_LENGTH_OF_EMAIL    = 100 // max length of email
)

// Other constants
const (
	ALLOW_USER_AGENT = "rich_chat" // allow user agent visit api
)

// WebSocket configuration constants
const (
	WEBSOCKET_WRITE_WAIT       = 10 * time.Second // Time allowed to write a message to the peer
	WEBSOCKET_PONG_WAIT        = 60 * time.Second // Time allowed to read the next pong message from the peer
	WEBSOCKET_PING_PERIOD      = 54 * time.Second // Send pings to peer (should be less than pong wait)
	WEBSOCKET_MAX_MESSAGE_SIZE = 5120             // Maximum message size allowed from peer (5KB)
)

// Database connection pool configuration
const (
	DB_MAX_OPEN_CONNS     = 25              // maximum number of open connections to the database
	DB_MAX_IDLE_CONNS     = 10              // maximum number of idle connections to the database
	DB_CONN_MAX_LIFETIME  = 5 * time.Minute // maximum amount of time a connection may be reused
	DB_CONN_MAX_IDLE_TIME = 5 * time.Minute // maximum amount of time a connection may be idle
)

// LoadConfig loads configuration from environment variables
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
	// Load JWT and Auth config
	JWT_SECRET = getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	AUTH_USERNAME = getEnv("AUTH_USERNAME", "admin")
	AUTH_PASSWORD = getEnv("AUTH_PASSWORD", "change-this-password")

	// Load Database config
	DB_HOST = getEnv("DB_HOST", "localhost")
	DB_PORT = getEnvInt("DB_PORT", 5432)
	DB_USER = getEnv("DB_USER", "postgres")
	DB_PASS = getEnv("DB_PASS", "")
	DB_NAME = getEnv("DB_NAME", "rich_chat")
	DB_SSL = getEnv("DB_SSL", "require") // Default to require for security

	// Load Redis config
	REDIS_HOST = getEnv("REDIS_HOST", "localhost")
	REDIS_PORT = getEnvInt("REDIS_PORT", 6379)
	REDIS_PASSWORD = getEnv("REDIS_PASSWORD", "")
	REDIS_DB = getEnvInt("REDIS_DB", 0)
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as integer with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
