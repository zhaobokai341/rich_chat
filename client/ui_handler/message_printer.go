package ui_handler

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// MessagePrinter handles styled message output
type MessagePrinter struct {
	infoStyle    lipgloss.Style
	warnStyle    lipgloss.Style
	errorStyle   lipgloss.Style
	successStyle lipgloss.Style
}

// NewMessagePrinter creates a new message printer with lipgloss styles
func NewMessagePrinter() *MessagePrinter {
	return &MessagePrinter{
		infoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5CACEE")).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("#5CACEE")).
			PaddingLeft(1),

		warnStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("#FFD700")).
			PaddingLeft(1),

		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4500")).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("#FF4500")).
			PaddingLeft(1),

		successStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#32CD32")).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("#32CD32")).
			PaddingLeft(1),
	}
}

// PrintInfo prints an informational message
func (mp *MessagePrinter) PrintInfo(message string) {
	fmt.Println(mp.infoStyle.Render("[i] " + message))
}

// PrintWarning prints a warning message
func (mp *MessagePrinter) PrintWarning(message string) {
	fmt.Println(mp.warnStyle.Render("[!] " + message))
}

// PrintError prints an error message
func (mp *MessagePrinter) PrintError(message string) {
	fmt.Println(mp.errorStyle.Render("[x] " + message))
}

// PrintSuccess prints a success message
func (mp *MessagePrinter) PrintSuccess(message string) {
	fmt.Println(mp.successStyle.Render("[ok] " + message))
}
