package main

import "fmt"

const ( // basic config
	LANGUAGE     = "zh"        // Language (zh/en)
	URL_SCHEMA   = "http"      // URL protocol (http/https)
	URL_DOMAIN   = "localhost" // URL domain
	URL_PORT     = 2316        // URL port
	URL_USERNAME = "admin"     // HTTP Basic Auth username
	URL_PASSWORD = "password"  // HTTP Basic Auth password
)

// Constructed URL root
var url_root = fmt.Sprintf("%s://%s:%s@%s:%d", URL_SCHEMA, URL_USERNAME, URL_PASSWORD, URL_DOMAIN, URL_PORT)

const (
	CONFIG_DIR  = "config"      // config directory
	CONFIG_FILE = "config.json" // config file
)

const (
	USER_AGENT                   = "rich_chat 1.0.0"
	ALLOW_MAX_LENGTH_OF_USERNAME = 50
)
