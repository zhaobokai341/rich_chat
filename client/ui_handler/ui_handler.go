package ui_handler

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

// input reads a line of input from the user
func input(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return input, nil
}

// UIHandler handles user interface interactions and flow control
type UIHandler struct {
	authService   AuthService
	userService   UserService
	apiClient     APIClient
	chatAPIClient ChatAPIClient
	chatService   ChatService
	configMgr     ConfigManager
	languagePack  LanguagePack
	printer       *MessagePrinter
	menuRenderer  *MenuRenderer
	configDir     string
}

// NewUIHandler creates a new UI handler
func NewUIHandler(
	authService AuthService,
	userService UserService,
	apiClient APIClient,
	chatAPIClient ChatAPIClient,
	chatService ChatService,
	configMgr ConfigManager,
	languagePack LanguagePack,
	configDir string,
) *UIHandler {
	return &UIHandler{
		authService:   authService,
		userService:   userService,
		apiClient:     apiClient,
		chatAPIClient: chatAPIClient,
		chatService:   chatService,
		configMgr:     configMgr,
		languagePack:  languagePack,
		printer:       NewMessagePrinter(),
		menuRenderer:  NewMenuRenderer(),
		configDir:     configDir,
	}
}

// Start begins the main application loop
func (h *UIHandler) Start() {
	if !h.checkServer() {
		return
	}

	for {
		if _, err := h.configMgr.ReadConfig(); err != nil {
			h.printError("reading_config_failed")
			os.Exit(1)
		}

		if !h.authService.IsAuthenticated() {
			h.handleUnauthenticated()
		} else {
			h.handleAuthenticated()
		}
	}
}

// checkServer verifies server availability
func (h *UIHandler) checkServer() bool {
	h.printInfo("connecting_server")
	available, err := h.apiClient.CheckServerHealth()
	if err != nil {
		h.printError("connection_error")
		return false
	}
	if !available {
		h.printError("server_health_check_failed")
		return false
	}
	h.printSuccess("connected_server")
	return true
}

// handleAuthenticated shows main menu for authenticated users
func (h *UIHandler) handleAuthenticated() {
	token, userID, err := h.authService.GetCredentials()
	if err != nil {
		h.printWarning("user_id_or_token_not_string")
		_ = h.authService.Logout()
		return
	}

	// Set authentication headers for all subsequent API requests
	h.apiClient.SetAuthHeaders(token, userID)

	h.showMainMenu(token, userID)
}

// showMainMenu displays and handles the main menu
func (h *UIHandler) showMainMenu(token, userID string) {
	for {
		h.printMainMenu()

		choice, err := input(h.languagePack.Get("choose_action"))
		if err != nil {
			h.printError("failed_to_read_choice")
			return
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			h.printInfo("exit")
			os.Exit(0)
		case "2":
			h.viewAndModifyUserInfo()
		case "3":
			h.handleChangePassword()
		case "4":
			h.handleLogout()
			return
		case "5":
			h.handleDeleteAccount()
			return
		case "6":
			h.handleChat(token, userID)
		default:
			h.printError("invalid_choice")
		}
	}
}

// printMainMenu displays the main menu with lipgloss styling
func (h *UIHandler) printMainMenu() {
	title := h.menuRenderer.RenderTitle(h.languagePack.Get("main_menu_title"))
	items := []string{
		h.languagePack.Get("menu_exit"),
		h.languagePack.Get("menu_view_modify_user_info"),
		h.languagePack.Get("menu_change_password"),
		h.languagePack.Get("menu_logout"),
		h.languagePack.Get("menu_delete_account"),
		h.languagePack.Get("menu_chat"),
	}
	menu := h.menuRenderer.RenderMenu(items)
	fmt.Println(title)
	fmt.Println(menu)
}

// readPassword securely reads a password from terminal
func (h *UIHandler) readPassword() (string, error) {
	fmt.Print(h.languagePack.Get("password_prompt"))
	bytePassword, err := term.ReadPassword(os.Stdin.Fd())
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(bytePassword), nil
}

// printInfo prints an informational message
func (h *UIHandler) printInfo(key string) {
	h.printer.PrintInfo(h.languagePack.Get(key))
}

// printWarning prints a warning message
func (h *UIHandler) printWarning(key string) {
	h.printer.PrintWarning(h.languagePack.Get(key))
}

// printError prints an error message
func (h *UIHandler) printError(text any) {
	var message string
	switch v := text.(type) {
	case string:
		message = h.languagePack.Get(v)
	case error:
		message = v.Error()
	default:
		message = fmt.Sprintf("%v", text)
	}
	h.printer.PrintError(message)
}

// printSuccess prints a success message
func (h *UIHandler) printSuccess(key string) {
	h.printer.PrintSuccess(h.languagePack.Get(key))
}

// MenuRenderer handles menu rendering with lipgloss
type MenuRenderer struct {
	titleStyle lipgloss.Style
	itemStyle  lipgloss.Style
	boxStyle   lipgloss.Style
}

// NewMenuRenderer creates a new menu renderer
func NewMenuRenderer() *MenuRenderer {
	return &MenuRenderer{
		titleStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#d0f112")).
			Bold(true).
			Underline(true),

		itemStyle: lipgloss.NewStyle().
			PaddingLeft(2),

		boxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5CACEE")).
			Padding(1, 2),
	}
}

// RenderTitle renders a menu title
func (mr *MenuRenderer) RenderTitle(title string) string {
	return mr.titleStyle.Render(title)
}

// RenderMenu renders a menu with items
func (mr *MenuRenderer) RenderMenu(items []string) string {
	var lines []string
	for i, item := range items {
		lines = append(lines, mr.itemStyle.Render(fmt.Sprintf("%d. %s", i+1, item)))
	}
	content := strings.Join(lines, "\n")
	return mr.boxStyle.Render(content)
}
