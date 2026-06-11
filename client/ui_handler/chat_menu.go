package ui_handler

import (
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// handleChat manages the chat flow
func (h *UIHandler) handleChat(token, userIDStr string) {
	userID, err := parseUserID(userIDStr)
	if err != nil {
		h.printError("invalid_user_id")
		return
	}

	for {
		h.printChatMenu()

		choice, err := input(h.languagePack.Get("choose_chat_action"))
		if err != nil {
			h.printError("failed_to_read_choice")
			return
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			h.handleChatSessions(userID, token)
		case "2":
			h.handleNewChat(userID, token)
		case "3":
			h.handleGenerateKeys(userID, token)
		case "4":
			return
		default:
			h.printError("invalid_choice")
		}
	}
}

// parseUserID converts user ID string to int
func parseUserID(userIDStr string) (int, error) {
	var userID int
	_, err := fmt.Sscanf(userIDStr, "%d", &userID)
	if err != nil {
		return 0, err
	}
	if userID <= 0 {
		return 0, fmt.Errorf("invalid user ID: %d", userID)
	}
	return userID, nil
}

// printChatMenu displays the chat menu
func (h *UIHandler) printChatMenu() {
	title := h.menuRenderer.RenderTitle(h.languagePack.Get("chat_menu_title"))
	items := []string{
		h.languagePack.Get("chat_my_sessions"),
		h.languagePack.Get("chat_new_chat"),
		h.languagePack.Get("chat_generate_keys"),
		h.languagePack.Get("chat_back_to_main"),
	}
	menu := h.menuRenderer.RenderMenu(items)
	fmt.Println(title)
	fmt.Println(menu)
}

// handleNewChat creates a new one-on-one chat by user ID
func (h *UIHandler) handleNewChat(userID int, token string) {
	fmt.Println()
	fmt.Println(h.menuRenderer.RenderTitle(h.languagePack.Get("new_chat_title")))

	recipientIDStr, err := input(h.languagePack.Get("enter_user_id"))
	if err != nil {
		h.printError("failed_to_read_input")
		return
	}
	recipientIDStr = strings.TrimSpace(recipientIDStr)

	recipientID, err := parseUserID(recipientIDStr)
	if err != nil || recipientID <= 0 {
		h.printError("invalid_user_id")
		return
	}

	if recipientID == userID {
		h.printWarning("cannot_chat_yourself")
		return
	}

	h.printInfo("getting_user_info")
	verifyToken, err := h.apiClient.GetVerifyToken()
	if err != nil {
		h.printError(fmt.Errorf("%s: %w", h.languagePack.Get("getting_verify_token_failed"), err))
		return
	}

	recipientInfo, err := h.chatAPIClient.GetUserBasicInfo(userID, token, verifyToken, recipientID)
	if err != nil {
		h.printError("user_not_found")
		return
	}

	recipientName := recipientInfo.Nickname
	if recipientName == "" {
		recipientName = recipientInfo.Username
	}

	fmt.Println()
	fmt.Printf("%s: %s (ID: %d)\n", h.languagePack.Get("user_to_chat"), recipientName, recipientID)
	fmt.Println()

	confirm, err := input(h.languagePack.Get("confirm_create_chat"))
	if err != nil {
		h.printError("failed_to_read_input")
		return
	}
	confirm = strings.TrimSpace(confirm)
	if confirm != "y" && confirm != "Y" {
		h.printInfo("chat_creation_cancelled")
		return
	}

	h.printInfo("creating_chat_session")
	verifyToken2, err := h.apiClient.GetVerifyToken()
	if err != nil {
		h.printError(fmt.Errorf("%s: %w", h.languagePack.Get("getting_verify_token_failed"), err))
		return
	}

	sessionID, err := h.chatAPIClient.CreateChatSession(userID, token, verifyToken2, recipientID)
	if err != nil {
		h.printError(fmt.Errorf("%s: %w", h.languagePack.Get("create_session_failed"), err))
		return
	}

	h.printSuccess("chat_session_created")
	fmt.Printf(h.languagePack.Get("session_id_display"), sessionID)
	fmt.Println()

	h.startOneOnOneChat(userID, token, sessionID, recipientID, recipientName)
}

// handleChatSessions displays and allows selection of existing chat sessions
func (h *UIHandler) handleChatSessions(userID int, token string) {
	h.printInfo("loading_sessions")
	verifyToken, err := h.apiClient.GetVerifyToken()
	if err != nil {
		h.printError(fmt.Errorf("%s: %w", h.languagePack.Get("getting_verify_token_failed"), err))
		return
	}

	sessions, err := h.chatAPIClient.GetUserSessions(userID, token, verifyToken)
	if err != nil {
		h.printError(err)
		return
	}

	if len(sessions) == 0 {
		h.printInfo("no_sessions_found")
		return
	}

	h.printSessionList(sessions)

	choice, err := input(h.languagePack.Get("select_session"))
	if err != nil {
		h.printError("failed_to_read_input")
		return
	}
	choice = strings.TrimSpace(choice)

	var selectedIdx int
	_, err = fmt.Sscanf(choice, "%d", &selectedIdx)
	if err != nil || selectedIdx < 1 || selectedIdx > len(sessions) {
		h.printError("invalid_selection")
		return
	}

	selectedSession := sessions[selectedIdx-1]

	partnerName := h.getPartnerName(userID, token, selectedSession.PartnerID)
	h.startOneOnOneChat(userID, token, selectedSession.SessionID, selectedSession.PartnerID, partnerName)
}

// printSessionList displays chat sessions in a styled list
func (h *UIHandler) printSessionList(sessions []*SessionInfo) {
	title := h.menuRenderer.RenderTitle(h.languagePack.Get("my_sessions_title"))

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5CACEE")).
		Padding(1, 2)

	var lines []string
	for i, session := range sessions {
		partnerName := session.PartnerName
		if partnerName == "" {
			partnerName = fmt.Sprintf("User %d", session.PartnerID)
		}
		lines = append(lines, fmt.Sprintf("%d. %s (Session ID: %d)", i+1, partnerName, session.SessionID))
	}

	fmt.Println(title)
	fmt.Println(cardStyle.Render(strings.Join(lines, "\n")))
	fmt.Println()
}

// getPartnerName retrieves partner name by ID
func (h *UIHandler) getPartnerName(currentUserID int, token string, partnerID int) string {
	verifyToken, err := h.apiClient.GetVerifyToken()
	if err != nil {
		return fmt.Sprintf("User %d", partnerID)
	}

	partnerInfo, err := h.chatAPIClient.GetUserBasicInfo(currentUserID, token, verifyToken, partnerID)
	if err != nil || partnerInfo == nil {
		return fmt.Sprintf("User %d", partnerID)
	}

	if partnerInfo.Nickname != "" {
		return partnerInfo.Nickname
	}
	return partnerInfo.Username
}

// handleGenerateKeys generates and uploads E2EE keys for the user
func (h *UIHandler) handleGenerateKeys(userID int, token string) {
	h.printInfo("generating_keys")

	verifyToken, err := h.apiClient.GetVerifyToken()
	if err != nil {
		h.printError(fmt.Errorf("%s: %w", h.languagePack.Get("getting_verify_token_failed"), err))
		return
	}

	err = h.chatService.GenerateAndUploadKeys(userID, token, verifyToken)
	if err != nil {
		h.printError(err)
		return
	}

	h.printSuccess("keys_generated_successfully")
}

// startOneOnOneChat starts a one-on-one chat session
func (h *UIHandler) startOneOnOneChat(userID int, token string, sessionID, partnerID int, partnerName string) {
	h.printInfo("connecting_chat")

	err := h.chatService.Connect(token, userID)
	if err != nil {
		h.printError(err)
		return
	}
	defer h.chatService.Disconnect()

	hasKey, err := h.chatService.HasEncryptionKey(userID)
	if err != nil {
		h.printWarning(fmt.Sprintf("failed_to_check_encryption_key: %v", err))
	}

	keyPath := fmt.Sprintf("%s/keys/%d.pem", h.configDir, userID)

	if !hasKey {
		if !h.handleNewDeviceScenario(userID, token, keyPath) {
			return
		}
	} else {
		err = h.chatService.LoadEncryptionKey(userID)
		if err != nil {
			h.printWarning(fmt.Sprintf("key_load_failed: %v", err))
		} else {
			log.Printf("Private key loaded into memory for user %d", userID)
		}

		verifyTokenCheck, errCheck := h.apiClient.GetVerifyToken()
		if errCheck == nil {
			_, _, errGet := h.chatAPIClient.GetUserPublicKey(userID, token, verifyTokenCheck, userID)
			if errGet != nil {
				h.printInfo("uploading_public_key_to_server")
				verifyToken4, err4 := h.apiClient.GetVerifyToken()
				if err4 == nil {
					if err := h.chatService.UploadPublicKeyOnly(userID, token, verifyToken4); err != nil {
						h.printWarning("key_upload_failed_continue_anyway")
					}
				}
			}
		}
	}

	h.printInfo("fetching_partner_key")
	partnerPubKey := ""
	verifyToken3, err3 := h.apiClient.GetVerifyToken()
	if err3 == nil {
		partnerPubKey, _, err3 = h.chatAPIClient.GetUserPublicKey(userID, token, verifyToken3, partnerID)
	}
	if err3 != nil {
		h.printWarning("partner_key_not_found_continue_anyway")
		partnerPubKey = ""
	}

	fmt.Println()
	fmt.Println(h.languagePack.Get("chat_started"))
	fmt.Printf(h.languagePack.Get("chatting_with")+"\n", partnerName)
	fmt.Println(h.languagePack.Get("chat_instructions"))
	fmt.Println()

	go h.receiveMessages(partnerName)

	for {
		msg, err := input("")
		if err != nil {
			continue
		}
		msg = strings.TrimSpace(msg)

		if msg == "" {
			continue
		}

		if msg == "/quit" || msg == "/exit" {
			h.printInfo("leaving_chat")
			return
		}

		if msg == "/status" {
			h.printInfo("chat_status_online")
			continue
		}

		if partnerPubKey == "" {
			h.printError("cannot_send_without_partner_key")
			continue
		}

		err = h.chatService.SendMessage(sessionID, partnerID, msg, partnerPubKey)
		if err != nil {
			h.printError(fmt.Errorf("%s: %w", h.languagePack.Get("message_send_failed"), err))
		}
	}
}

// handleNewDeviceScenario handles the new device key import flow
func (h *UIHandler) handleNewDeviceScenario(userID int, token, keyPath string) bool {
	fmt.Println()
	fmt.Printf(h.languagePack.Get("new_device_detected"), keyPath)
	fmt.Println()

	choice, err := input("")
	if err != nil {
		h.printError(err)
		return false
	}
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return h.handleImportExistingKey(userID, token, keyPath)
	case "2":
		return h.handleGenerateNewKey(userID, token, keyPath)
	case "3":
		h.printInfo("leaving_chat")
		return false
	default:
		h.printError("invalid_choice")
		return false
	}
}

// handleImportExistingKey handles importing an existing private key file
func (h *UIHandler) handleImportExistingKey(userID int, token, _ string) bool {
	fmt.Println()
	fmt.Print(h.languagePack.Get("please_select_private_key_file"))
	keyFilePath, err := input("")
	if err != nil {
		h.printError(err)
		return false
	}
	keyFilePath = strings.TrimSpace(keyFilePath)

	keyData, err := os.ReadFile(keyFilePath)
	if err != nil {
		h.printError(fmt.Errorf("%s: %s", h.languagePack.Get("private_key_file_not_found"), keyFilePath))
		return false
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		h.printError(fmt.Errorf("%s", h.languagePack.Get("invalid_private_key_file")))
		return false
	}

	err = h.chatService.SaveEncryptionKey(userID, string(keyData))
	if err != nil {
		h.printError(fmt.Errorf("%s: %v", h.languagePack.Get("key_save_failed"), err))
		return false
	}

	verifyToken, err := h.apiClient.GetVerifyToken()
	if err != nil {
		h.printError(fmt.Errorf("%s: %v", h.languagePack.Get("getting_verify_token_failed"), err))
		return false
	}

	err = h.chatService.UploadPublicKeyOnly(userID, token, verifyToken)
	if err != nil {
		h.printError(fmt.Errorf("%s: %v", h.languagePack.Get("private_key_upload_failed"), err))
		return false
	}

	err = h.chatService.LoadEncryptionKey(userID)
	if err != nil {
		h.printError(fmt.Errorf("%s: %v", h.languagePack.Get("key_load_failed"), err))
		return false
	}

	h.printSuccess("key_import_successful")
	return true
}

// handleGenerateNewKey handles generating a new key pair
func (h *UIHandler) handleGenerateNewKey(userID int, token, keyPath string) bool {
	h.printInfo("generating_keys_for_chat")

	verifyToken, err := h.apiClient.GetVerifyToken()
	if err != nil {
		h.printError(fmt.Errorf("%s: %v", h.languagePack.Get("getting_verify_token_failed"), err))
		return false
	}

	err = h.chatService.GenerateAndUploadKeys(userID, token, verifyToken)
	if err != nil {
		h.printWarning("key_generation_failed_continue_anyway")
	} else {
		err = h.chatService.LoadEncryptionKey(userID)
		if err != nil {
			h.printWarning(fmt.Sprintf("key_load_failed: %v", err))
		}

		fmt.Println()
		fmt.Printf(h.languagePack.Get("private_key_security_warning"), keyPath)
		fmt.Println()
		_, _ = input("")
	}

	return true
}

// receiveMessages listens for incoming messages
func (h *UIHandler) receiveMessages(partnerName string) {
	messages := h.chatService.GetMessages()
	for msg := range messages {
		decrypted, err := h.chatService.DecryptMessage(msg)
		if err != nil {
			log.Printf("Failed to decrypt message: %v", err)
			fmt.Printf("\n[%s] [Failed to decrypt message]\n> ", partnerName)
			continue
		}
		fmt.Printf("\n[%s] %s\n> ", partnerName, decrypted)
	}
}

// printChatConnected displays chat connected message
func (h *UIHandler) printChatConnected(partnerName string) {
	title := h.menuRenderer.RenderTitle(h.languagePack.Get("chat_connected_title"))
	message := fmt.Sprintf(h.languagePack.Get("chat_connected_message"), partnerName)
	fmt.Println(title)
	fmt.Println(message)
}
