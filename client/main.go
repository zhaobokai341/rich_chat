package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Application holds all dependencies for the client application
type Application struct {
	apiClient    APIClient
	configMgr    ConfigManager
	authService  *AuthService
	userService  *UserService
	uiHandler    *UIHandler
	languagePack *LanguagePackWrapper
}

// NewApplication creates and initializes a new application instance
func NewApplication() *Application {
	// Initialize language pack
	lp := NewLanguagePackWrapper("client/main.json", LANGUAGE)

	// Initialize HTTP client
	httpClient := NewHTTPClient(USER_AGENT)

	// Initialize API client
	baseURL := url_root
	apiClient := NewRestAPIClient(httpClient, baseURL, lp)

	// Initialize config manager
	configMgr := NewFileConfigManager(CONFIG_DIR, CONFIG_FILE)

	// Initialize token extractor
	tokenExtractor := NewJWTTokenExtractor()

	// Initialize services
	authService := NewAuthService(apiClient, configMgr, tokenExtractor)
	userService := NewUserService(apiClient, configMgr, tokenExtractor)

	// Initialize UI handler
	uiHandler := NewUIHandler(authService, userService, apiClient, configMgr, lp)

	return &Application{
		apiClient:    apiClient,
		configMgr:    configMgr,
		authService:  authService,
		userService:  userService,
		uiHandler:    uiHandler,
		languagePack: lp,
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
	app := NewApplication()
	app.Run()
}
