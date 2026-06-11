package main

import (
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

// Claims represents JWT claims structure
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func main() {
	log.SetReportCaller(true)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(
		&log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		},
	)
	log.Info("Initializing...")

	// Load configuration from environment variables
	LoadConfig()

	// Initialize application (replaces global variables)
	app, err := NewApp()
	if err != nil {
		log.Fatal("Failed to initialize application: ", err)
	}
	defer app.Close()

	log.Info("Starting server...")
	if err := app.Engine.Run(WEB_PORT); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
