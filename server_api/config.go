package main

import (
	"time"
)

const ( // basic config
	WEB_PORT      = ":2316"    // http port
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
	ALLOW_USER_AGENT             = "rich_chat"         // allow user agent visit api
	JWT_EXPIRE_TIME              = time.Hour * 24 * 30 // 30 days, about 1 month
	ALLOW_MAX_LENGTH_OF_USERNAME = 50                  // max length of username
)
