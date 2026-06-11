package ui_handler

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
)

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

	h.printUserInfoCard(userData)
}

// printUserInfoCard displays user info in a styled card
func (h *UIHandler) printUserInfoCard(userData *UserData) {
	title := h.menuRenderer.RenderTitle(h.languagePack.Get("user_info_title"))

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#32CD32")).
		Padding(1, 2)

	lines := []string{
		fmt.Sprintf(h.languagePack.Get("user_info_username"), userData.Username),
		fmt.Sprintf(h.languagePack.Get("user_info_email"), userData.Email),
		fmt.Sprintf(h.languagePack.Get("user_info_nickname"), userData.Nickname),
		fmt.Sprintf(h.languagePack.Get("user_info_bio"), userData.Bio),
	}

	content := strings.Join(lines, "\n")
	fmt.Println(title)
	fmt.Println(cardStyle.Render(content))
	fmt.Println()
}

// modifyUserInfo allows user to modify profile fields
func (h *UIHandler) modifyUserInfo() {
	for {
		h.printModifyUserInfoMenu()

		choice, err := input(h.languagePack.Get("choose_info_to_modify"))
		if err != nil {
			h.printError("failed_to_read_input")
			return
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			h.modifyField("nickname", "enter_new_nickname")
		case "2":
			h.modifyField("email", "enter_new_email")
		case "3":
			h.modifyField("bio", "enter_new_bio")
		case "4":
			return
		default:
			h.printError("invalid_choice")
		}
	}
}

// printModifyUserInfoMenu displays the modify user info menu
func (h *UIHandler) printModifyUserInfoMenu() {
	items := []string{
		h.languagePack.Get("menu_modify_nickname"),
		h.languagePack.Get("menu_modify_email"),
		h.languagePack.Get("menu_modify_bio"),
		h.languagePack.Get("menu_back"),
	}
	menu := h.menuRenderer.RenderMenu(items)
	fmt.Println(menu)
}

// modifyField handles modification of a single field
func (h *UIHandler) modifyField(field, promptKey string) {
	newValue, err := input(h.languagePack.Get(promptKey))
	if err != nil {
		h.printError("failed_to_read_input")
		return
	}
	newValue = strings.TrimSpace(newValue)

	if err := h.userService.UpdateProfile(field, newValue); err != nil {
		h.printError(err)
		return
	}
	h.printSuccess("user_info_changed_successfully")
}

// handleChangePassword manages the password change flow
func (h *UIHandler) handleChangePassword() {
	oldPassword, err := h.readPasswordSecure(h.languagePack.Get("enter_old_password"))
	if err != nil {
		h.printError("failed_to_read_password")
		return
	}

	if oldPassword == "" {
		h.printError("old_password_cannot_be_empty")
		return
	}

	newPassword, err := h.getNewPasswordWithConfirmation()
	if err != nil {
		return
	}

	h.printInfo("changing_password")
	if err := h.userService.ChangePassword(oldPassword, newPassword); err != nil {
		h.printError(err)
		return
	}

	h.printSuccess("password_changed_successfully")
}

// readPasswordSecure reads a password with a custom prompt
func (h *UIHandler) readPasswordSecure(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(os.Stdin.Fd())
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(bytePassword), nil
}

// getNewPasswordWithConfirmation reads and confirms a new password
func (h *UIHandler) getNewPasswordWithConfirmation() (string, error) {
	var newPassword string
	for {
		fmt.Print(h.languagePack.Get("enter_new_password"))
		newPasswordBytes, err := term.ReadPassword(os.Stdin.Fd())
		if err != nil {
			h.printError("failed_to_read_password")
			return "", err
		}
		fmt.Println()
		newPassword = string(newPasswordBytes)

		if newPassword == "" {
			h.printError("new_password_cannot_be_empty")
			continue
		}

		if err := validatePassword(newPassword); err != nil {
			h.printWarning(err.Error())
			continue
		}

		fmt.Print(h.languagePack.Get("confirm_new_password"))
		confirmPasswordBytes, err := term.ReadPassword(os.Stdin.Fd())
		if err != nil {
			h.printError("failed_to_read_password")
			return "", err
		}
		fmt.Println()
		confirmPassword := string(confirmPasswordBytes)

		if confirmPassword == "" {
			h.printError("new_password_cannot_be_empty")
			continue
		}

		if newPassword != confirmPassword {
			h.printError("new_passwords_do_not_match")
			continue
		}

		break
	}
	return newPassword, nil
}
