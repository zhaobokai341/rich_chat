package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

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

