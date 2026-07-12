package ui_handler

import (
	"fmt"
	"os"
	"strings"
)

// handleUnauthenticated shows login/register menu
func (h *UIHandler) handleUnauthenticated() {
	h.printInfo("not_logged_in")
	h.printLoginMenu()

	choice, err := input(h.languagePack.Get("choose_login_method"))
	if err != nil {
		h.printError("failed_to_read_choice")
		return
	}
	choice = strings.TrimSpace(choice)

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

// printLoginMenu displays the login/register menu
func (h *UIHandler) printLoginMenu() {
	items := []string{
		h.languagePack.Get("menu_login"),
		h.languagePack.Get("menu_register"),
		h.languagePack.Get("menu_exit"),
	}
	menu := h.menuRenderer.RenderMenu(items)
	fmt.Println(menu)
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

	// Clear authentication headers
	h.apiClient.ClearAuthHeaders()

	h.printSuccess("logout_successful")
}

// handleDeleteAccount manages the account deletion flow
func (h *UIHandler) handleDeleteAccount() {
	h.printWarning("delete_account_warning")

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

	// Clear authentication headers
	h.apiClient.ClearAuthHeaders()

	h.printSuccess("delete_account_successful")
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

		if err := validatePassword(password); err != nil {
			h.printWarning(err.Error())
		}

		fmt.Print(h.languagePack.Get("confirm_password"))
		password2, err := h.readPassword()
		if err != nil {
			h.printError("failed_to_read_password")
			return "", err
		}
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
