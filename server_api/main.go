package main

import (
	"net/http"
	"rich_chat/lang_pack_load"
	"rich_chat/server_api/database"
	"rich_chat/server_api/service"
	"rich_chat/server_api/websocket"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

// WebServerApi holds all service dependencies for HTTP handlers
type WebServerApi struct {
	authService  service.AuthService
	userService  service.UserService
	tokenService service.TokenService
	websocketHub *websocket.Hub
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

// Initialize all components with dependency injection
func initialize() {
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
		MaxPasswordLength: ALLOW_MAX_LENGTH_OF_PASSWORD,
		MaxBioLength:      ALLOW_MAX_LENGTH_OF_BIO,
		MaxEmailLength:    ALLOW_MAX_LENGTH_OF_EMAIL,
	})

	// Initialize WebSocket hub
	websocketHub := websocket.NewHub()

	// Start the hub in a separate goroutine
	go websocketHub.Run()

	// Initialize web server API with injected services
	web_server_engine = gin.Default()
	web_server_api = &WebServerApi{
		authService:  services.AuthService,
		userService:  services.UserService,
		tokenService: services.TokenService,
		websocketHub: websocketHub,
	}

	// Add middleware to extract language from request
	web_server_engine.Use(languageMiddleware())

	// Middleware
	web_server_engine.Use(
		gin.CustomRecovery(
			func(c *gin.Context, recovered interface{}) {
				// Get request method and path
				method := c.Request.Method
				path := c.Request.URL.Path

				// Output error message
				log.Errorf("PANIC: %s %s - Error: %v\nStack:\n%s",
					method, path, recovered, debug.Stack())

				// Response with internal server error
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "Server internal error",
				})
			},
		),
	)
	web_server_engine.Use(force_https())
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
	web_server_engine.PUT("/api/users/:user_id/password", web_server_api.ChangeUserPassword)

	// WebSocket endpoint for real-time chat
	websocketConfig := websocket.Config{
		WRITEWAIT:      WEBSOCKET_WRITE_WAIT,
		PONGWAIT:       WEBSOCKET_PONG_WAIT,
		PINGPERIOD:     WEBSOCKET_PING_PERIOD,
		MAXMESSAGESIZE: WEBSOCKET_MAX_MESSAGE_SIZE,
	}
	websocketHandler := websocket.NewHandler(web_server_api.websocketHub, web_server_api.authService, web_server_api.userService, JWT_SECRET, websocketConfig)
	web_server_engine.GET("/ws/chat", func(c *gin.Context) {
		websocketHandler.WebSocketEndpoint(c)
	})
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
	if err := web_server_engine.Run(WEB_PORT); err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}

// languageMiddleware extracts language from request and stores it in context
func languageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract language from query param, form data, or header
		lang := c.Query("language")
		if lang == "" {
			lang = c.PostForm("language")
		}
		if lang == "" {
			lang = c.GetHeader("X-Language")
		}
		if lang == "" {
			lang = DEFAULT_LANGUAGE // fallback to default
		}

		// Store the language in context for later use
		c.Set("language", lang)

		c.Next()
	}
}

// getLanguagePackFromContext returns a language pack based on the language in the request context
func getLanguagePackFromContext(c *gin.Context) *lang_pack_load.LanguagePack {
	lang, exists := c.Get("language")
	if !exists {
		lang = DEFAULT_LANGUAGE
	}

	language, ok := lang.(string)
	if !ok {
		language = DEFAULT_LANGUAGE
	}

	lp := lang_pack_load.NewLanguagePack("server_api/main.json", language)
	lp.Load()
	return lp
}
