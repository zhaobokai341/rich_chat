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

func (usr_data *UserInput) get_user_token_and_id() (string, string) {
	usr_token, exists := usr_data.user_data["token"]
	usr_id, exists := usr_data.user_data["user_id"]
	usr_token_str, usr_token_ok := usr_token.(string)
	if !exists {
		var err error
		usr_id, err = extractUserIDFromToken(usr_token_str)
		if err != nil {
			print("warning", fmt.Sprintf(lp.G("failed_to_extract_user_id"), err))
			usr_id = "unknown"
		}
		usr_data.user_data["user_id"] = usr_id
		usr_data.update_config()
	}
	usr_id_str, usr_id_ok := usr_id.(string)
	if !usr_id_ok || !usr_token_ok {
		delete(usr_data.user_data, "token")
		delete(usr_data.user_data, "user_id")
		usr_data.update_config()
		print("warning", lp.G("user_id_or_token_not_string"))
		os.Exit(1)
	}
	return usr_token_str, usr_id_str
}

func (usr_data *UserInput) choose_login_method() {
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
}

func (usr_data *UserInput) index_panel() {
	for {
		choice, err := input(lp.G("choose_action"))
		if err != nil {
			print("error", "Failed to read choice")
			return
		}
		choice = strings.TrimSpace(choice)
		switch choice {
		case "1":
			usr_data.logout()
			return
		case "2":
			print("info", lp.G("exit"))
			os.Exit(0)
		case "3":
			usr_data.deleteAccount()
			return
		default:
			print("error", "Invalid choice")
		}
	}
}

// Main function
func (usr_data *UserInput) start() {
	if !usr_data.check_server() {
		return
	}
	for {
		usr_data.read_config()
		_, exists := usr_data.user_data["token"]
		if !exists {
			usr_data.choose_login_method()
		} else {
			usr_token, usr_id := usr_data.get_user_token_and_id()
			requests.SetHeader("user_token", usr_token)
			requests.SetHeader("user_id", usr_id)
			usr_data.index_panel()
		}
	}
}
