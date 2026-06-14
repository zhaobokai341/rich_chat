package main

import (
	"net/http"
	"rich_chat/server_api/database"
	"rich_chat/server_api/service"
	"rich_chat/server_api/websocket"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
) // App holds all application dependencies (replaces global variables)
type App struct {
	Engine       *gin.Engine
	API          *WebServerAPI
	DBService    *database.DatabaseService
	Services     *service.Services
	WebSocketHub websocket.HubInterface
}

// WebServerAPI holds all service dependencies for HTTP handlers
type WebServerAPI struct {
	authService  service.AuthService
	userService  service.UserService
	chatService  service.ChatService
	tokenService service.TokenService
	websocketHub websocket.HubInterface
}

// NewApp creates and initializes a new application instance
func NewApp() (*App, error) {
	app := &App{}

	// Initialize Redis manager
	redisManager := redis_init()

	// Initialize database service
	var err error
	app.DBService, err = database.InitializeDatabaseService(
		database.Config{
			DB_HOST:                   DB_HOST,
			DB_PORT:                   DB_PORT,
			DB_USER:                   DB_USER,
			DB_PASS:                   DB_PASS,
			DB_NAME:                   DB_NAME,
			DB_SSL:                    DB_SSL,
			MAXOPENCONNS:              DB_MAX_OPEN_CONNS,
			MAXIDLECONNS:              DB_MAX_IDLE_CONNS,
			CONNMAXIDLETIME:           DB_CONN_MAX_IDLE_TIME,
			CONNMAXLIFETIME:           DB_CONN_MAX_LIFETIME,
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
		return nil, err
	}

	// Initialize service layer
	app.Services = service.NewServices(app.DBService, service.ServiceConfig{
		JWTSecret:         JWT_SECRET,
		JWTExpiration:     JWT_EXPIRE_TIME,
		MaxUsernameLength: ALLOW_MAX_LENGTH_OF_USERNAME,
		VerifyTokenTTL:    VERIFY_TOKEN_EXPIRE_TIME,
		MaxPasswordLength: ALLOW_MAX_LENGTH_OF_PASSWORD,
		MaxBioLength:      ALLOW_MAX_LENGTH_OF_BIO,
		MaxEmailLength:    ALLOW_MAX_LENGTH_OF_EMAIL,
		MaxNicknameLength: ALLOW_MAX_LENGTH_OF_NICKNAME,
	})

	// Initialize WebSocket hub
	app.WebSocketHub = websocket.NewHub()
	go app.WebSocketHub.Run()

	// Initialize API handlers
	app.API = &WebServerAPI{
		authService:  app.Services.AuthService,
		userService:  app.Services.UserService,
		chatService:  app.Services.ChatService,
		tokenService: app.Services.TokenService,
		websocketHub: app.WebSocketHub,
	}

	// Initialize Gin engine
	app.Engine = gin.Default()
	app.setupMiddleware()
	app.setupRoutes()

	return app, nil
}

// setupMiddleware configures all middleware
func (app *App) setupMiddleware() {
	deps := &MiddlewareDependencies{
		DBService: app.DBService,
		Services:  app.Services,
	}

	app.Engine.Use(languageMiddleware())
	app.Engine.Use(request_id())

	app.Engine.Use(
		gin.CustomRecovery(
			func(c *gin.Context, recovered interface{}) {
				method := c.Request.Method
				path := c.Request.URL.Path
				requestID, _ := c.Get("request_id")

				log.WithFields(log.Fields{
					"request_id": requestID,
					"error":      recovered,
				}).Errorf("PANIC: %s %s - Stack:\n%s",
					method, path, debug.Stack())

				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    http.StatusInternalServerError,
					"message": "Server internal error",
				})
			},
		),
	)

	app.Engine.Use(force_https())
	app.Engine.Use(check_ip_in_block(deps))
	app.Engine.Use(track_ip_visit(deps))
	app.Engine.Use(check_client_user_agent())

	app.Engine.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/ws/chat" {
			c.Next()
			return
		}
		gin.BasicAuth(gin.Accounts{
			AUTH_USERNAME: AUTH_PASSWORD,
		})(c)
	})
}

// setupRoutes configures all routes
func (app *App) setupRoutes() {
	deps := &MiddlewareDependencies{
		DBService: app.DBService,
		Services:  app.Services,
	}

	// Health check endpoints (no auth required)
	app.Engine.GET("/health", app.API.HealthCheck)
	app.Engine.GET("/ready", app.API.ReadinessCheck)

	app.Engine.GET("/", app.API.Index)

	authGroup := app.Engine.Group("/api/auth")
	{
		authGroup.GET("/token", app.API.GetVerifyToken)
		authGroup.POST("/login", app.API.Login, check_verify_token(deps))
		authGroup.POST("/register", app.API.Register, check_verify_token(deps))
	}

	userGroup := app.Engine.Group(
		"/api/users/:user_id",
		check_client_token(),
		check_user_is_exists(deps),
		check_verify_token(deps),
	)
	{
		userGroup.POST("/delete", app.API.DeleteUser)
		userGroup.GET("/profile", app.API.GetUserProfile)
		userGroup.PATCH("/profile", app.API.ChangeUserProfile)
		userGroup.PUT("/password", app.API.ChangeUserPassword)

		userGroup.GET("/basic", app.API.GetUserBasicInfo)

		userGroup.POST("/sessions", app.API.CreateChatSession)
		userGroup.GET("/sessions", app.API.GetUserSessions)

		userGroup.GET("/public-key", app.API.GetUserPublicKeyHandler)
		userGroup.POST("/keys", app.API.StoreUserKeyHandler)
	}

	websocketConfig := websocket.Config{
		WRITEWAIT:             WEBSOCKET_WRITE_WAIT,
		PONGWAIT:              WEBSOCKET_PONG_WAIT,
		PINGPERIOD:            WEBSOCKET_PING_PERIOD,
		MAXMESSAGESIZE:        WEBSOCKET_MAX_MESSAGE_SIZE,
		SEND_CHANNEL_BUFFER:   WEBSOCKET_SEND_CHANNEL_BUFFER,
		MAX_MESSAGES_PER_SEC:  WEBSOCKET_MAX_MESSAGES_PER_SEC,
		OFFLINE_MESSAGE_DELAY: WEBSOCKET_OFFLINE_MSG_DELAY,
	}
	websocketHandler := websocket.NewHandler(
		app.API.websocketHub,
		app.API.authService,
		app.API.userService,
		app.API.chatService,
		JWT_SECRET,
		websocketConfig,
	)
	app.Engine.GET("/ws/chat", func(c *gin.Context) {
		websocketHandler.WebSocketEndpoint(c)
	})
}

// Close cleans up application resources
func (app *App) Close() {
	if app.DBService != nil && app.DBService.GetDB() != nil {
		app.DBService.GetDB().Close()
	}
	if app.WebSocketHub != nil {
		app.WebSocketHub.Close()
	}
}
