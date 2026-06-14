package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Create HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:    WEB_PORT,
		Handler: app.Engine,
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Starting server on %s", WEB_PORT)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start: ", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Create a deadline for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server (waits for active connections to finish)
	if err := srv.Shutdown(ctx); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Server forced to shutdown")
	}

	log.Info("Server exited")
}
