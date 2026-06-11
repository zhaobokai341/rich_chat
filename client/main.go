package main

import (
	"github.com/charmbracelet/lipgloss"

	"rich_chat/client/ui_handler"
)

// Application holds all dependencies for the client application
type Application struct {
	apiClient     APIClient
	chatAPIClient *ChatAPIClient
	configMgr     ConfigManager
	authService   *AuthService
	userService   *UserService
	chatService   *ChatService
	uiHandler     *ui_handler.UIHandler
	languagePack  *LanguagePackWrapper
}

// NewApplication creates and initializes a new application instance
func NewApplication() *Application {
	// Initialize language pack with default language
	lp := NewLanguagePackWrapper(LANGUAGE_PACK, LANGUAGE)

	// Initialize HTTP client
	httpClient := NewHTTPClient(USER_AGENT)

	// Initialize API client
	baseURL := url_root
	apiClient := NewRestAPIClient(httpClient, baseURL, lp)

	// Initialize chat API client
	chatAPIClient := NewChatAPIClient(httpClient, baseURL, lp)

	// Initialize config manager
	configMgr := NewFileConfigManager(CONFIG_DIR, CONFIG_FILE, lp)

	// Initialize token extractor
	tokenExtractor := NewJWTTokenExtractorWithLanguagePack(lp)

	// Initialize services
	authService := NewAuthService(apiClient, configMgr, tokenExtractor, lp)
	userService := NewUserService(apiClient, configMgr, tokenExtractor, lp)
	chatService := NewChatService(chatAPIClient, configMgr, lp)

	// Initialize UI handler with adapters
	uiHandler := ui_handler.NewUIHandler(
		authService,
		&userServiceAdapter{service: userService},
		apiClient,
		&chatAPIClientAdapter{client: chatAPIClient},
		&chatServiceAdapter{service: chatService},
		configMgr,
		lp,
		CONFIG_DIR,
	)

	return &Application{
		apiClient:     apiClient,
		chatAPIClient: chatAPIClient,
		configMgr:     configMgr,
		authService:   authService,
		userService:   userService,
		chatService:   chatService,
		uiHandler:     uiHandler,
		languagePack:  lp,
	}
}

// Run starts the application
func (app *Application) Run() {
	title_style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d0f112")).
		Bold(true).
		Underline(true)

	println(title_style.Render(app.languagePack.Get("welcome")))
	app.uiHandler.Start()
	print("info", app.languagePack.Get("exit"))
}

func main() {
	LoadConfig()
	app := NewApplication()
	app.Run()
}
