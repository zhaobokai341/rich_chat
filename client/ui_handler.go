package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
)

// UIHandler handles user interface interactions and flow control
type UIHandler struct {
	authService  *AuthService
	userService  *UserService
	apiClient    APIClient
	configMgr    ConfigManager
	languagePack *LanguagePackWrapper
}

// NewUIHandler creates a new UI handler
func NewUIHandler(authService *AuthService, userService *UserService, apiClient APIClient, configMgr ConfigManager, lp *LanguagePackWrapper) *UIHandler {
	return &UIHandler{
		authService:  authService,
		userService:  userService,
		apiClient:    apiClient,
		configMgr:    configMgr,
		languagePack: lp,
	}
}

// Start begins the main application loop
func (h *UIHandler) Start() {
	// Check server health
	if !h.checkServer() {
		return
	}

	for {
		// Load configuration
		if _, err := h.configMgr.ReadConfig(); err != nil {
			h.printError("reading_config_failed")
			os.Exit(1)
		}

		// Check if user is authenticated
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

// handleUnauthenticated shows login/register menu
func (h *UIHandler) handleUnauthenticated() {
	h.printInfo("not_logged_in")
	choice, err := input(h.languagePack.Get("choose_login_method"))
	if err != nil {
		h.printError("failed_to_read_choice")
		return
	}
	choice = strings.TrimSpace(choice)

	// 1. Login
	// 2. Register
	// 3. Exit
	switch choice {
	case "1":
		h.handleLogin()
	case "2":
		h.handleRegister()
	case "3":
		h.printInfo("exit")
		os.Exit(0)
	default:
		h.printError("invalid_choice")
	}
}

// handleAuthenticated shows main menu for authenticated users
func (h *UIHandler) handleAuthenticated() {
	token, userID, err := h.authService.GetCredentials()
	if err != nil {
		h.printWarning("user_id_or_token_not_string")
		h.authService.Logout()
		return
	}

	// Set authentication headers
	httpClient := h.apiClient.(*RestAPIClient).client
	httpClient.SetHeader("user_token", token)
	httpClient.SetHeader("user_id", userID)

	h.showMainMenu()
}

// showMainMenu displays and handles the main menu
func (h *UIHandler) showMainMenu() {
	for {
		choice, err := input(h.languagePack.Get("choose_action"))
		if err != nil {
			h.printError("failed_to_read_choice")
			return
		}
		choice = strings.TrimSpace(choice)

		// 1. Exit
		// 2. View and modify user info
		// 3. Change password
		// 4. Logout
		// 5. Delete account
		switch choice {
		case "1":
			h.printInfo("exit")
			os.Exit(0)
		case "2":
			h.viewAndModifyUserInfo()
		case "4":
			h.handleLogout()
			return
		case "5":
			h.handleDeleteAccount()
			return
		default:
			h.printError("invalid_choice")
		}
	}
}

// handleLogin manages the login flow
func (h *UIHandler) handleLogin() {
	username, err := input(h.languagePack.Get("username_prompt"))
	if err != nil {
		h.printError("failed_to_read_username")
		return
	}

	username = strings.TrimSpace(username)
	if username == "" {
		h.printError("username_cannot_be_empty")
		return
	}

	password, err := h.readPassword()
	if err != nil {
		h.printError("failed_to_read_password")
		return
	}

	if password == "" {
		h.printError("password_cannot_be_empty")
		return
	}

	h.printInfo("logging_in")
	if err := h.authService.Login(username, password); err != nil {
		h.printError(err)
		return
	}

	h.printSuccess("login_successful")
}

// handleRegister manages the registration flow
func (h *UIHandler) handleRegister() {
	username, err := input(h.languagePack.Get("username_prompt"))
	if err != nil {
		h.printError("failed_to_read_username")
		return
	}

	username = strings.TrimSpace(username)
	if username == "" {
		h.printError("username_cannot_be_empty")
		return
	}

	if len(username) > ALLOW_MAX_LENGTH_OF_USERNAME {
		h.printError("username_too_long")
		return
	}

	password, err := h.getPasswordWithConfirmation()
	if err != nil {
		return
	}

	h.printInfo("registering")
	if err := h.authService.Register(username, password); err != nil {
		h.printError(err)
		return
	}

	h.printSuccess("register_successful")
}

// handleLogout manages the logout flow
func (h *UIHandler) handleLogout() {
	if err := h.authService.Logout(); err != nil {
		h.printError(err)
		return
	}

	// Clear HTTP headers
	httpClient := h.apiClient.(*RestAPIClient).client
	httpClient.SetHeader("user_token", "")
	httpClient.SetHeader("user_id", "")

	h.printSuccess("logout_successful")
}

// handleDeleteAccount manages the account deletion flow
func (h *UIHandler) handleDeleteAccount() {
	confirm, err := input(h.languagePack.Get("delete_account_confirm"))
	if err != nil {
		h.printError("failed_to_read_input")
		return
	}
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm != "y" && confirm != "yes" {
		h.printInfo("exit")
		return
	}

	password, err := h.readPassword()
	if err != nil {
		h.printError("failed_to_read_password")
		return
	}

	if password == "" {
		h.printError("password_cannot_be_empty")
		return
	}

	h.printInfo("deleting_account")
	if err := h.userService.DeleteAccount(password); err != nil {
		h.printError(err)
		return
	}

	// Clear HTTP headers
	httpClient := h.apiClient.(*RestAPIClient).client
	httpClient.SetHeader("user_token", "")
	httpClient.SetHeader("user_id", "")

	h.printSuccess("delete_account_successful")
}

// viewAndModifyUserInfo displays and allows modification of user info
func (h *UIHandler) viewAndModifyUserInfo() {
	h.viewUserInfo()
	h.modifyUserInfo()
}

// viewUserInfo displays current user information
func (h *UIHandler) viewUserInfo() {
	h.printInfo("viewing_user_info")
	userData, err := h.userService.GetProfile()
	if err != nil {
		h.printError(err)
		return
	}

	fmt.Println(h.languagePack.Get("user_info_title"))
	fmt.Printf(h.languagePack.Get("user_info_username")+"\n", userData.Username)
	fmt.Printf(h.languagePack.Get("user_info_nickname")+"\n", userData.Nickname)
	fmt.Printf(h.languagePack.Get("user_info_bio")+"\n", userData.Bio)
	fmt.Println()
}

// modifyUserInfo allows user to modify profile fields
func (h *UIHandler) modifyUserInfo() {
	for {
		choice, err := input(h.languagePack.Get("choose_info_to_modify"))
		if err != nil {
			h.printError("failed_to_read_input")
			return
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			newNickname, err := input(h.languagePack.Get("enter_new_nickname"))
			if err != nil {
				h.printError("failed_to_read_input")
				return
			}
			newNickname = strings.TrimSpace(newNickname)
			if newNickname == "" {
				h.printError("nickname_cannot_be_empty")
				continue
			}
			if err := h.userService.UpdateProfile("nickname", newNickname); err != nil {
				h.printError(err)
				return
			}
			h.printSuccess("user_info_changed_successfully")
			return
		case "2":
			newBio, err := input(h.languagePack.Get("enter_new_bio"))
			if err != nil {
				h.printError("failed_to_read_input")
				return
			}
			newBio = strings.TrimSpace(newBio)
			if newBio == "" {
				h.printError("bio_cannot_be_empty")
				continue
			}
			if err := h.userService.UpdateProfile("bio", newBio); err != nil {
				h.printError(err)
				return
			}
			h.printSuccess("user_info_changed_successfully")
			return
		case "3":
			return
		default:
			h.printError("invalid_choice")
		}
	}
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

// getPasswordWithConfirmation reads and confirms a password
func (h *UIHandler) getPasswordWithConfirmation() (string, error) {
	var password string
	for {
		var err error
		password, err = h.readPassword()
		if err != nil {
			h.printError("failed_to_read_password")
			return "", err
		}

		if password == "" {
			h.printError("password_cannot_be_empty")
			continue
		}

		if len(password) < 8 {
			h.printWarning("password_too_short")
		}

		fmt.Print(h.languagePack.Get("confirm_password"))
		bytePassword2, err := term.ReadPassword(os.Stdin.Fd())
		if err != nil {
			h.printError("failed_to_read_password")
			return "", err
		}
		fmt.Println()

		password2 := string(bytePassword2)
		if password2 == "" {
			h.printError("password_cannot_be_empty")
			continue
		}

		if password == password2 {
			break
		}
		h.printError("passwords_do_not_match")
	}
	return password, nil
}

// printInfo prints an informational message
func (h *UIHandler) printInfo(key string) {
	print("info", h.languagePack.Get(key))
}

// printWarning prints a warning message
func (h *UIHandler) printWarning(key string) {
	print("warning", h.languagePack.Get(key))
}

// printError prints an error message
func (h *UIHandler) printError(text any) {
	var key string
	switch v := text.(type) {
	case string:
		key = h.languagePack.Get(v)
	case error:
		key = text.(error).Error()
	}
	print("error", key)
}

// printSuccess prints a success message
func (h *UIHandler) printSuccess(key string) {
	print("success", h.languagePack.Get(key))
}
