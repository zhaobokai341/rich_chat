package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Check server if it's available
func (usr_data *UserInput) check_server() bool {
	print("info", lp.G("connecting_server"))
	resp, err := requests.R().Get(URL_ROOT)
	if err != nil {
		print("error", lp.G("connection_error"))
		return false
	}
	if resp.StatusCode() != 200 {
		print("error", fmt.Sprintf("%s: %d", lp.G("http_status_code_error"), resp.StatusCode()))
		return false
	}
	if resp.String() != "Welcome to Rich Chat!" {
		print("error", lp.G("is_not_rich_chat_server"))
		return false
	}
	print("success", lp.G("connected_server"))
	return true
}

// Load user config file
func (usr_data *UserInput) read_config() {
	print("info", lp.G("reading_config"))
	dir_info, err := os.Stat("config")
	if err != nil {
		err = os.Mkdir("config", 0755)
		if err != nil {
			print("error", lp.G("create_config_dir_error"))
			os.Exit(1)
		}
		print("info", lp.G("created_config_dir"))
	} else if !dir_info.IsDir() {
		print("error", lp.G("config_dir_is_not_dir"))
		os.Exit(1)
	}
	_, err = os.Stat("config/config.json")
	if err != nil {
		file, err := os.Create("config/config.json")
		if err != nil {
			print("error", lp.G("create_config_file_error"))
			os.Exit(1)
		}
		print("info", lp.G("created_config_file"))
		_, err = file.WriteString("{}")
		if err != nil {
			print("error", lp.G("write_config_file_error"))
			os.Exit(1)
		}
		defer file.Close()
	}
	file_data, err := os.ReadFile("config/config.json")
	if err != nil {
		print("error", lp.G("read_config_file_error"))
		os.Exit(1)
	}
	err = json.Unmarshal(file_data, &usr_data.user_data)
	if err != nil {
		print("error", lp.G("parse_config_file_error"))
		os.Exit(1)
	}
	print("success", lp.G("read_successfully"))
}

// Update user config file
func (usr_data *UserInput) update_config() {
	json_bytes, err := json.Marshal(usr_data.user_data)
	if err != nil {
		print("error", lp.G("map_convert_to_str_error"))
		os.Exit(1)
	}
	err = os.WriteFile("config/config.json", json_bytes, 0644)
	if err != nil {
		print("error", lp.G("write_config_file_error"))
		os.Exit(1)
	}
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
	password, err := input(lp.G("password_prompt"))
	if err != nil {
		print("error", lp.G("failed_to_read_password"))
		return
	}
	password = strings.TrimSpace(password)
	if password == "" {
		print("error", lp.G("password_cannot_be_empty"))
		return
	}

	// Send login request
	print("info", lp.G("logging_in"))
	resp, err := requests.R().
		SetFormData(map[string]string{
			"username": username,
			"password": password,
		}).
		Post(fmt.Sprintf("%s/login", URL_ROOT))

	if err != nil {
		print("error", fmt.Sprintf(lp.G("login_request_failed"), err))
		return
	}

	if resp.StatusCode() != 200 {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &errorResp); err == nil {
			if msg, ok := errorResp["message"].(string); ok {
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
	if len(username) > 50 {
		print("error", lp.G("username_too_long"))
		return
	}

	// Get password
	password, err := input(lp.G("password_prompt"))
	if err != nil {
		print("error", lp.G("failed_to_read_password"))
		return
	}
	password = strings.TrimSpace(password)
	if password == "" {
		print("error", lp.G("password_cannot_be_empty"))
		return
	}

	// Validate password length
	if len(password) < 8 {
		print("warning", lp.G("password_too_short"))
	}

	// Send register request
	print("info", lp.G("registering"))
	resp, err := requests.R().
		SetFormData(map[string]string{
			"username": username,
			"password": password,
		}).
		Post(fmt.Sprintf("%s/register", URL_ROOT))

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

// Main function
func (usr_data *UserInput) start() {
	if !usr_data.check_server() {
		return
	}
	usr_data.read_config()
	usr_token, exists := usr_data.user_data["token"]
	if !exists {
		// Show menu for login/register choice
		print("info", lp.G("not_logged_in"))
		choice, err := input(lp.G("choose_login_method"))
		if err != nil {
			print("error", "Failed to read choice")
			return
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			usr_data.login()
		case "2":
			usr_data.register()
		case "3":
			print("info", lp.G("exit"))
			os.Exit(0)
		default:
			print("error", "Invalid choice")
			return
		}
	} else {
		usr_id, exists := usr_data.user_data["user_id"]
		if !exists {
			delete(usr_data.user_data, "token")
			usr_data.update_config()
			print("warning", lp.G("user_id_not_found"))
			usr_data.login()
			return
		}
		usr_token_str, usr_token_ok := usr_token.(string)
		usr_id_str, usr_id_ok := usr_id.(string)
		if !usr_id_ok || !usr_token_ok {
			delete(usr_data.user_data, "token")
			delete(usr_data.user_data, "user_id")
			usr_data.update_config()
			print("warning", lp.G("user_id_or_token_not_string"))
			usr_data.login()
			return
		}
		requests.SetHeader("user_token", usr_token_str)
		requests.SetHeader("user_id", usr_id_str)
		requests.R().Post(fmt.Sprintf("%s/api/login", URL_ROOT))
	}
}
