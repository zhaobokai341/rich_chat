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
	WEB_PORT = ":2316" // http port
	LANGUAGE = "zh"    // language (zh/en)
	VERSION  = "1.0.0" // product version, DO NOT CHANGE
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

const (
	MAXOPENCONNS    = 25              // maximum number of open connections to the database
	MAXIDLECONNS    = 10              // maximum number of idle connections to the database
	CONNMAXLIFETIME = 5 * time.Minute // maximum amount of time a connection may be reused
	CONNMAXIDLETIME = 5 * time.Minute // maximum amount of time a connection may be idle
)

// Redis configuration variables
var (
	REDIS_HOST     string // Redis host - loaded from environment
	REDIS_PORT     int    // Redis port - loaded from environment
	REDIS_PASSWORD string // Redis password - loaded from environment
	REDIS_DB       int    // Redis database number - loaded from environment
)

const (
	CACHE_TTL      = 300 // Cache TTL in seconds (5 minutes)
	CACHE_NULL_TTL = 60  // Cache TTL for null values in seconds (1 minute)
)

// Rate limiting configuration variables
const (
	IP_LIMIT_TIME             = time.Minute * 10 // Time window for IP rate limiting (1 minute)
	IP_LIMIT_VISIT_TIMES      = 100              // Maximum visits per IP within IP_LIMIT_TIME
	IP_LIMIT_LOCKOUT_DURATION = time.Minute * 10 // IP lockout duration after IP_LIMIT_VISIT_TIMES
)

// Other constants
const (
	ALLOW_USER_AGENT             = "rich_chat"         // allow user agent visit api
	JWT_EXPIRE_TIME              = time.Hour * 24 * 30 // 30 days, about 1 month
	ALLOW_MAX_LENGTH_OF_USERNAME = 50                  // max length of username
	VERIFY_TOKEN_EXPIRE_TIME     = time.Minute * 5     // Verification token expire time (5 minutes)
	MAX_LOGIN_ATTEMPTS           = 5                   // Maximum login attempts before lockout
	LOCKOUT_DURATION             = time.Minute * 15    // Account lockout duration after max attempts
)

// LoadConfig loads configuration from environment variables
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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
