package main

import (
	"rich_chat/lang_pack_load"
	"rich_chat/server_api/database"
	"rich_chat/server_api/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

// TODO: force HTTPS

// WebServerApi holds all service dependencies for HTTP handlers
type WebServerApi struct {
	authService  service.AuthService
	userService  service.UserService
	tokenService service.TokenService
}

// Claims represents JWT claims structure
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

var web_server_engine *gin.Engine
var web_server_api *WebServerApi
var dbService *database.DatabaseService
var services *service.Services
var lp *lang_pack_load.LanguagePack

var log_func = map[string]func(v ...interface{}){
	"info":  log.Info,
	"debug": log.Debug,
	"warn":  log.Warning,
	"error": log.Error,
}

// Initialize all components with dependency injection
func initialize() {
	// Load language pack
	lp = lang_pack_load.NewLanguagePack("server_api/main.json", LANGUAGE)
	lp.Load()

	// Initialize Redis manager
	redisManager := redis_init()

	// Initialize database service (new architecture)
	var err error
	dbService, err = database.InitializeDatabaseService(
		database.Config{
			DB_HOST:                   DB_HOST,
			DB_PORT:                   DB_PORT,
			DB_USER:                   DB_USER,
			DB_PASS:                   DB_PASS,
			DB_NAME:                   DB_NAME,
			DB_SSL:                    DB_SSL,
			MAXOPENCONNS:              MAXOPENCONNS,
			MAXIDLECONNS:              MAXIDLECONNS,
			CONNMAXIDLETIME:           CONNMAXIDLETIME,
			CONNMAXLIFETIME:           CONNMAXLIFETIME,
			MAX_LOGIN_ATTEMPTS:        MAX_LOGIN_ATTEMPTS,
			LOCKOUT_DURATION:          LOCKOUT_DURATION,
			IP_LIMIT_TIME:             IP_LIMIT_TIME,
			IP_LIMIT_VISIT_TIMES:      IP_LIMIT_VISIT_TIMES,
			IP_LIMIT_LOCKOUT_DURATION: IP_LIMIT_LOCKOUT_DURATION,
		},
		database.RedisManager{
			GetCache:         redisManager.GetCache,
			GetIntValue:      redisManager.GetIntValue,
			SetCache:         redisManager.SetCache,
			SetNullCache:     redisManager.SetNullCache,
			SetCacheWithTTL:  redisManager.SetCacheWithTTL,
			SetLockoutCache:  redisManager.SetLockoutCache,
			IncrementCounter: redisManager.IncrementCounter,
			SetKeyExpiration: redisManager.SetKeyExpiration,
			DeleteCache:      redisManager.DeleteCache,
		},
	)
	if err != nil {
		log.Fatal("Failed to initialize database service: ", err)
	}

	// Initialize service layer (new!)
	services = service.NewServices(dbService, service.ServiceConfig{
		JWTSecret:         JWT_SECRET,
		JWTExpiration:     JWT_EXPIRE_TIME,
		MaxUsernameLength: ALLOW_MAX_LENGTH_OF_USERNAME,
		VerifyTokenTTL:    VERIFY_TOKEN_EXPIRE_TIME,
	})

	// Initialize web server API with injected services
	web_server_engine = gin.Default()
	web_server_api = &WebServerApi{
		authService:  services.AuthService,
		userService:  services.UserService,
		tokenService: services.TokenService,
	}

	// Middleware
	web_server_engine.Use(safe_check())
	web_server_engine.Use(gin.BasicAuth(gin.Accounts{
		AUTH_USERNAME: AUTH_PASSWORD,
	}))

	// Routes - RESTful API design
	web_server_engine.GET("/", web_server_api.Index)

	// Authentication endpoints
	web_server_engine.GET("/api/auth/token", web_server_api.GetVerifyToken)
	web_server_engine.POST("/api/auth/login", web_server_api.Login)
	web_server_engine.POST("/api/auth/register", web_server_api.Register)

	// User management endpoints
	web_server_engine.POST("/api/users/:user_id/delete", web_server_api.DeleteUser)
	web_server_engine.GET("/api/users/:user_id/profile", web_server_api.GetUserProfile)
	web_server_engine.PATCH("/api/users/:user_id/profile", web_server_api.ChangeUserProfile)
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

	initialize()
	defer dbService.GetDB().Close()

	log.Info("Starting server...")
	web_server_engine.Run(WEB_PORT)
}
