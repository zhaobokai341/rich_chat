package main

import (
	"fmt"
	"rich_chat/lang_pack_load"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-resty/resty/v2"
)

var lp *lang_pack_load.LanguagePack
var requests *resty.Client

const (
	LANGUAGE = "zh"
	URL_ROOT = "http://localhost:2316"
)

type UserInput struct {
	Cookie string
}

func print(style_type string, text string) {
	var character string
	style := lipgloss.NewStyle()
	switch style_type {
	case "info":
		style = style.Foreground(lipgloss.Color("#3B82F6")) // Blue
		character = "[*]"
	case "warning":
		style = style.Foreground(lipgloss.Color("#EAB308")) // Yellow
		character = "[!]"
	case "error":
		style = style.Foreground(lipgloss.Color("#EF4444")) // Red
		character = "[-]"
	case "success":
		style = style.Foreground(lipgloss.Color("#22C55E")) // Green
		character = "[+]"
	case "critical":
		style = style.Foreground(lipgloss.Color("#EF4444")) // Red
		character = "[!!!]"
	}
	format_text := style.Render(fmt.Sprintf("%s %s", character, text))
	fmt.Println(format_text)
	if style_type == "critical" {
		panic(text)
	}
}

func initialize() {
	lp = lang_pack_load.NewLanguagePack("client/main.json", LANGUAGE)
	lp.Load()
	requests = resty.New()
	requests.SetHeader("User-Agent", "rich_chat 1.0.0")
}

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

func (usr_data *UserInput) start() {
	if !usr_data.check_server() {
		print("info", lp.G("exit"))
		return
	}
}

func main() {
	initialize()
	title_style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d0f112")).
		Bold(true).
		Underline(true)
	fmt.Println(title_style.Render(lp.G("welcome")))
	usr_data := UserInput{}
	usr_data.start()
}
