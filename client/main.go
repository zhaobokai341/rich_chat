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
	// Initialize language pack with default language
	lp := NewLanguagePackWrapper("client/main.json", DEFAULT_LANGUAGE)

	// Initialize HTTP client
	httpClient := NewHTTPClient(USER_AGENT)

	// Initialize API client
	baseURL := url_root
	apiClient := NewRestAPIClient(httpClient, baseURL, lp)

	// Initialize config manager
	configMgr := NewFileConfigManager(CONFIG_DIR, CONFIG_FILE, lp)

	// Initialize token extractor
	tokenExtractor := NewJWTTokenExtractorWithLanguagePack(lp)

	// Initialize services
	authService := NewAuthService(apiClient, configMgr, tokenExtractor, lp)
	userService := NewUserService(apiClient, configMgr, tokenExtractor, lp)

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
