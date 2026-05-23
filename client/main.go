package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"rich_chat/lang_pack_load"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-resty/resty/v2"
)

var lp *lang_pack_load.LanguagePack
var requests *resty.Client

const (
	LANGUAGE = "zh"                                   // Language，你可以修改为zh
	URL_ROOT = "http://admin:password@localhost:2316" // URL root
)

// output, user input and do something
type UserInput struct {
	user_data map[string]interface{}
}

// beautiful print
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
	case "debug":
		style = style.Foreground(lipgloss.Color("#8b7e7e")) // Grey
		character = "[DEBUG]"
	}
	format_text := style.Render(fmt.Sprintf("%s %s", character, text))
	fmt.Println(format_text)
	if style_type == "critical" {
		panic(text)
	}
}

// simplify user input
func input(text string) (string, error) {
	fmt.Print(text)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return input, nil
}

// Initialize
func initialize() {
	lp = lang_pack_load.NewLanguagePack("client/main.json", LANGUAGE)
	lp.Load()
	requests = resty.New()
	requests.SetHeader("User-Agent", "rich_chat 1.0.0")
}

// extractUserIDFromToken extracts the user_id from a JWT token
func extractUserIDFromToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode token payload: %v", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse token claims: %v", err)
	}

	userID, ok := claims["user_id"]
	if !ok {
		return "", fmt.Errorf("user_id not found in token")
	}

	// Convert to string
	switch v := userID.(type) {
	case float64:
		return fmt.Sprintf("%.0f", v), nil
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func main() {
	initialize()
	title_style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d0f112")).
		Bold(true).
		Underline(true)
	fmt.Println(title_style.Render(lp.G("welcome")))
	usr_data := UserInput{
		make(map[string]interface{}),
	}
	usr_data.start()
	print("info", lp.G("exit"))
}
