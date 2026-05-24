package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
)

// Get verification token from server
func (usr_data *UserInput) getVerifyToken() (string, error) {
	print("info", lp.G("getting_verify_token"))
	resp, err := requests.R().Post(fmt.Sprintf("%s/api/get_verify_token", url_root))
	if err != nil {
		return "", fmt.Errorf(lp.G("verify_token_request_failed"), err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode())
	}

	var tokenResp map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &tokenResp); err != nil {
		return "", fmt.Errorf("%s", lp.G("verify_token_parse_failed"))
	}

	verifyToken, exists := tokenResp["verify_token"]
	if !exists {
		return "", fmt.Errorf("no verify_token in response")
	}

	verifyTokenStr, ok := verifyToken.(string)
	if !ok {
		return "", fmt.Errorf("invalid verify_token format")
	}

	return verifyTokenStr, nil
}

// Login to server
func (usr_data *UserInput) login() {
	// Get username
	username, err := input(lp.G("username_prompt"))
	if err != nil {
		print("error", lp.G("failed_to_read_username"))
		return
	}
	username = strings.TrimSpace(username)
	if username == "" {
		print("error", lp.G("username_cannot_be_empty"))
		return
	}

	// Get password
	fmt.Print(lp.G("password_prompt"))
	bytepassword, err := term.ReadPassword(os.Stdin.Fd())
	if err != nil {
		print("error", lp.G("failed_to_read_password"))
		return
	}
	password := string(bytepassword)
	password = strings.TrimSpace(password)
	if password == "" {
		print("error", lp.G("password_cannot_be_empty"))
		return
	}

	// Get verification token
	verifyToken, err := usr_data.getVerifyToken()
	if err != nil {
		print("error", err.Error())
		return
	}

	// Send login request
	print("info", lp.G("logging_in"))
	resp, err := requests.R().
		SetFormData(map[string]string{
			"username":     username,
			"password":     password,
			"verify_token": verifyToken,
		}).
		Post(fmt.Sprintf("%s/api/login", url_root))

	if err != nil {
		print("error", fmt.Sprintf(lp.G("login_request_failed"), err))
		return
	}

	if resp.StatusCode() != 200 {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			if msg, ok := errorResp["message"].(string); ok {
				// Check for account locked message
				if resp.StatusCode() == 429 {
					print("error", lp.G("account_locked_try_later"))
					return
				}

				// Show remaining attempts if available
				if remaining, ok := errorResp["remaining_attempts"]; ok {
					switch v := remaining.(type) {
					case float64:
						print("warning", fmt.Sprintf(lp.G("remaining_attempts"), int(v)))
					}
				}

				print("error", fmt.Sprintf(lp.G("login_failed"), msg))
			} else {
				print("error", fmt.Sprintf(lp.G("login_failed_with_status"), resp.StatusCode()))
			}
		} else {
			print("error", fmt.Sprintf(lp.G("login_failed_with_status"), resp.StatusCode()))
		}
		return
	}

	// Parse response
	var loginResp map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &loginResp); err != nil {
		print("error", lp.G("failed_to_parse_login_response"))
		return
	}

	userToken, tokenExists := loginResp["user_token"]
	if !tokenExists {
		print("error", lp.G("no_token_received"))
		return
	}

	userTokenStr, tokenOk := userToken.(string)
	if !tokenOk {
		print("error", lp.G("invalid_token_format"))
		return
	}

	// Extract user_id from JWT token or use from response
	var userID string
	if userIDFromResp, exists := loginResp["user_id"]; exists {
		switch v := userIDFromResp.(type) {
		case float64:
			userID = fmt.Sprintf("%.0f", v)
		case string:
			userID = v
		default:
			userID = fmt.Sprintf("%v", v)
		}
	} else {
		// Fallback to extracting from token
		userID, err = extractUserIDFromToken(userTokenStr)
		if err != nil {
			print("warning", fmt.Sprintf(lp.G("failed_to_extract_user_id"), err))
			userID = "unknown"
		}
	}

	// Save token and user_id to config
	usr_data.user_data["token"] = userTokenStr
	usr_data.user_data["user_id"] = userID

	usr_data.update_config()
	print("success", lp.G("login_successful"))
}

// Register new user
func (usr_data *UserInput) register() {
	// Get username
	username, err := input(lp.G("username_prompt"))
	if err != nil {
		print("error", lp.G("failed_to_read_username"))
		return
	}
	username = strings.TrimSpace(username)
	if username == "" {
		print("error", lp.G("username_cannot_be_empty"))
		return
	}

	// Validate username length
	if len(username) > ALLOW_MAX_LENGTH_OF_USERNAME {
		print("error", lp.G("username_too_long"))
		return
	}

	var password string
	for {
		// Get password
		fmt.Print(lp.G("password_prompt"))
		bytepassword, err := term.ReadPassword(os.Stdin.Fd())
		if err != nil {
			print("error", lp.G("failed_to_read_password"))
			return
		}
		fmt.Println()
		password = string(bytepassword)
		if password == "" {
			print("error", lp.G("password_cannot_be_empty"))
			continue
		}

		// Validate password length
		if len(password) < 8 {
			print("warning", lp.G("password_too_short"))
		}

		// Confirm password
		fmt.Print(lp.G("confirm_password"))
		bytepassword2, err := term.ReadPassword(os.Stdin.Fd())
		if err != nil {
			print("error", lp.G("failed_to_read_password"))
			return
		}
		fmt.Println()
		password2 := string(bytepassword2)
		if password2 == "" {
			print("error", lp.G("password_cannot_be_empty"))
			continue
		}

		if password == password2 {
			break
		}
		print("error", lp.G("passwords_do_not_match"))
	}

	// Get verification token
	verifyToken, err := usr_data.getVerifyToken()
	if err != nil {
		print("error", err.Error())
		return
	}

	// Send register request
	print("info", lp.G("registering"))
	resp, err := requests.R().
		SetFormData(map[string]string{
			"username":     username,
			"password":     password,
			"verify_token": verifyToken,
		}).
		Post(fmt.Sprintf("%s/api/register", url_root))

	if err != nil {
		print("error", fmt.Sprintf(lp.G("register_request_failed"), err))
		return
	}

	if resp.StatusCode() != 200 {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			if msg, ok := errorResp["message"].(string); ok {
				// Check for specific error types
				if resp.StatusCode() == 409 {
					print("error", lp.G("username_already_exists"))
				} else if resp.StatusCode() == 429 {
					print("error", lp.G("account_locked_try_later"))
				} else {
					print("error", fmt.Sprintf(lp.G("register_failed"), msg))
				}
			} else {
				print("error", fmt.Sprintf(lp.G("register_failed_with_status"), resp.StatusCode()))
			}
		} else {
			print("error", fmt.Sprintf(lp.G("register_failed_with_status"), resp.StatusCode()))
		}
		return
	}

	// Parse response
	var registerResp map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &registerResp); err != nil {
		print("error", lp.G("failed_to_parse_register_response"))
		return
	}

	userToken, tokenExists := registerResp["user_token"]
	if !tokenExists {
		print("error", lp.G("no_token_received"))
		return
	}

	userTokenStr, tokenOk := userToken.(string)
	if !tokenOk {
		print("error", lp.G("invalid_token_format"))
		return
	}

	// Extract user_id from response or token
	var userID string
	if userIDFromResp, exists := registerResp["user_id"]; exists {
		switch v := userIDFromResp.(type) {
		case float64:
			userID = fmt.Sprintf("%.0f", v)
		case string:
			userID = v
		default:
			userID = fmt.Sprintf("%v", v)
		}
	} else {
		// Fallback to extracting from token
		userID, err = extractUserIDFromToken(userTokenStr)
		if err != nil {
			print("warning", fmt.Sprintf(lp.G("failed_to_extract_user_id"), err))
			userID = "unknown"
		}
	}

	// Save token and user_id to config
	usr_data.user_data["token"] = userTokenStr
	usr_data.user_data["user_id"] = userID

	usr_data.update_config()
	print("success", lp.G("register_successful"))
}

// Logout
func (usr_data *UserInput) logout() {
	delete(usr_data.user_data, "token")
	delete(usr_data.user_data, "user_id")
	usr_data.update_config()
	requests.SetHeader("user_token", "")
	requests.SetHeader("user_id", "")
	print("success", lp.G("logout_successful"))
}

// Delete account
func (usr_data *UserInput) deleteAccount() {
	// Confirm deletion
	confirm, err := input(lp.G("delete_account_confirm"))
	if err != nil {
		print("error", lp.G("failed_to_read_password"))
		return
	}
	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm != "y" && confirm != "yes" {
		print("info", lp.G("exit"))
		return
	}

	// Get password for verification
	fmt.Print(lp.G("password_prompt_for_delete"))
	bytepassword, err := term.ReadPassword(os.Stdin.Fd())
	if err != nil {
		print("error", lp.G("failed_to_read_password"))
		return
	}
	password := string(bytepassword)
	password = strings.TrimSpace(password)
	if password == "" {
		print("error", lp.G("password_cannot_be_empty"))
		return
	}

	// Get user_id from config
	usr_id, exists := usr_data.user_data["user_id"]
	if !exists {
		print("error", lp.G("user_id_or_token_not_string"))
		return
	}
	usr_id_str, ok := usr_id.(string)
	if !ok {
		print("error", lp.G("user_id_or_token_not_string"))
		return
	}

	// Get verification token
	verifyToken, err := usr_data.getVerifyToken()
	if err != nil {
		print("error", err.Error())
		return
	}

	// Send delete account request
	print("info", lp.G("deleting_account"))
	resp, err := requests.R().
		SetFormData(map[string]string{
			"user_id":       usr_id_str,
			"user_password": password,
			"verify_token":  verifyToken,
		}).
		Post(fmt.Sprintf("%s/api/delete_user", url_root))

	if err != nil {
		print("error", fmt.Sprintf(lp.G("delete_account_request_failed"), err))
		return
	}

	if resp.StatusCode() != 200 {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			if msg, ok := errorResp["message"].(string); ok {
				if resp.StatusCode() == 429 {
					print("error", lp.G("account_locked_try_later"))
					return
				}
				print("error", fmt.Sprintf(lp.G("delete_account_failed"), msg))
			} else if errMsg, ok := errorResp["error"].(string); ok {
				if resp.StatusCode() == 429 {
					print("error", lp.G("account_locked_try_later"))
					return
				}
				print("error", fmt.Sprintf(lp.G("delete_account_failed"), errMsg))
			} else {
				print("error", fmt.Sprintf(lp.G("delete_account_failed_with_status"), resp.StatusCode()))
			}
		} else {
			print("error", fmt.Sprintf(lp.G("delete_account_failed_with_status"), resp.StatusCode()))
		}
		return
	}

	// Clear local data
	delete(usr_data.user_data, "token")
	delete(usr_data.user_data, "user_id")
	usr_data.update_config()
	requests.SetHeader("user_token", "")
	requests.SetHeader("user_id", "")

	print("success", lp.G("delete_account_successful"))
}
