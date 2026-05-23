package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const (
	WEB_PORT      = ":2316"
	AUTH_USERNAME = "admin"
	AUTH_PASSWORD = "password"
	JWT_SECRET    = "secret"
	VERSION       = "1.0.0"
)

type WebServerApi struct {
	database *sqlx.DB
}
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

var web_server_engine *gin.Engine
var web_server_api *WebServerApi
var user_database *UserDatabase

// Generate a JWT token for a given user ID
func generateToken(user_id int) (string, error) {
	expirationTime := time.Now().Add(30 * 24 * time.Hour)

	claims := &Claims{
		UserID: user_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rich_chat",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWT_SECRET))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func initialize() {
	web_server_engine = gin.Default()
	web_server_api = &WebServerApi{}

	web_server_engine.Use(gin.BasicAuth(gin.Accounts{
		AUTH_USERNAME: AUTH_PASSWORD,
	}))
	web_server_engine.Use(safe_check())

	// Routes
	web_server_engine.GET("/", web_server_api.Index)
	web_server_engine.POST("/login", web_server_api.Login)
	web_server_engine.POST("/register", web_server_api.Register)
}

func main() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})
	log.Info("Initializing...")
	initialize()
	user_database = database_init()
	defer user_database.database.Close()
	log.Info("Starting server...")
	web_server_engine.Run(WEB_PORT)
}
