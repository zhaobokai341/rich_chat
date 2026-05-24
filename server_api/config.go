package main

import (
	"time"
)

const ( // basic config
	WEB_PORT      = ":2316"    // http port
	LANGUAGE      = "zh"       // language (zh/en)
	AUTH_USERNAME = "admin"    // HTTP Basic Auth username
	AUTH_PASSWORD = "password" // HTTP Basic Auth password
	JWT_SECRET    = "secret"   // JWT secret
	VERSION       = "1.0.0"    // product version, DO NOT CHANGE
)

const (
	DB_HOST = "localhost" // database host
	DB_PORT = 5432        // database port
	DB_USER = "postgres"  // database user
	DB_PASS = "123456"    // database password
	DB_NAME = "rich_chat" // database name
	DB_SSL  = "disable"   // database ssl mode (disable, enable, verify-full)
)

const (
	REDIS_HOST     = "localhost" // Redis host
	REDIS_PORT     = 5350        // Redis port
	REDIS_PASSWORD = "123456"    // Redis password
	REDIS_DB       = 0           // Redis database number
	CACHE_TTL      = 300         // Cache TTL in seconds (5 minutes)
	CACHE_NULL_TTL = 60          // Cache TTL for null values in seconds (1 minute)
)

const (
	ALLOW_USER_AGENT             = "rich_chat"         // allow user agent visit api
	JWT_EXPIRE_TIME              = time.Hour * 24 * 30 // 30 days, about 1 month
	ALLOW_MAX_LENGTH_OF_USERNAME = 50                  // max length of username
	VERIFY_TOKEN_EXPIRE_TIME     = time.Minute * 5     // Verification token expire time (5 minutes)
	MAX_LOGIN_ATTEMPTS           = 5                   // Maximum login attempts before lockout
	LOCKOUT_DURATION             = time.Minute * 15    // Account lockout duration after max attempts
)
