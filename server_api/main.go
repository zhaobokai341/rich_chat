package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
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

var log_func = map[string]func(v ...interface{}){
	"info":  log.Info,
	"debug": log.Debug,
	"warn":  log.Warning,
	"error": log.Error,
}

// Generate log message with specified parameters
func log_output(log_type string, ip_address string, user_agent string, fmt_msg string, args ...interface{}) {
	prefix := fmt.Sprintf("From %s, User-Agent is %s. ", ip_address, user_agent)
	finalMsg := prefix + fmt.Sprintf(fmt_msg, args...)
	if fn, ok := log_func[log_type]; ok {
		fn(finalMsg)
	} else {
		log.Println(finalMsg)
	}
}

func log_with_ctx(c *gin.Context, log_type string, fmt_msg string, args ...interface{}) {
	log_output(log_type, c.Request.RemoteAddr, c.GetHeader("User-Agent"), fmt_msg, args...)
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
	web_server_engine.POST("/api/login", web_server_api.Login)
	web_server_engine.POST("/api/register", web_server_api.Register)
	web_server_engine.POST("/api/delete_user", web_server_api.DeleteUser)
}

func main() {
	log.SetReportCaller(true)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&logrus.JSONFormatter{})
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
