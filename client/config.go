package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// basic config
var (
	LANGUAGE      string // Default Language (zh/en)
	LANGUAGE_PACK string // Language pack file
	URL_SCHEMA    string // URL protocol (http/https)
	URL_DOMAIN    string // URL domain
	URL_PORT      int    // URL port
	URL_USERNAME  string // HTTP Basic Auth username
	URL_PASSWORD  string // HTTP Basic Auth password
)

// config file config
var (
	CONFIG_DIR  string // config directory
	CONFIG_FILE string // config file
)

// other config
var (
	USER_AGENT                   string // user agent string
	ALLOW_MAX_LENGTH_OF_USERNAME int    // max length of username
)

// Constructed URL root
var url_root string

// LoadConfig loads configuration from environment variables
func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Load URL config
	LANGUAGE = getEnv("LANGUAGE", "en")
	LANGUAGE_PACK = getEnv("LANGUAGE_PACK", "client/main.json")
	URL_SCHEMA = getEnv("URL_SCHEMA", "http")
	URL_DOMAIN = getEnv("URL_DOMAIN", "localhost")
	URL_PORT = getEnvInt("URL_PORT", 2316)
	URL_USERNAME = getEnv("URL_USERNAME", "admin")
	URL_PASSWORD = getEnv("URL_PASSWORD", "password")

	// Load application config
	CONFIG_DIR = getEnv("CONFIG_DIR", "config")
	CONFIG_FILE = getEnv("CONFIG_FILE", "config.json")
	USER_AGENT = getEnv("USER_AGENT", "rich_chat 1.0.0")
	ALLOW_MAX_LENGTH_OF_USERNAME = getEnvInt("ALLOW_MAX_LENGTH_OF_USERNAME", 50)

	// Construct URL root
	url_root = fmt.Sprintf("%s://%s:%s@%s:%d", URL_SCHEMA, URL_USERNAME, URL_PASSWORD, URL_DOMAIN, URL_PORT)
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
