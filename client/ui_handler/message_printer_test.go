package ui_handler

import (
	"testing"
)

func TestNewMessagePrinter(t *testing.T) {
	mp := NewMessagePrinter()
	if mp == nil {
		t.Fatal("NewMessagePrinter() returned nil")
	}
	if mp.infoStyle.GetForeground() == nil {
		t.Error("infoStyle foreground not set")
	}
	if mp.warnStyle.GetForeground() == nil {
		t.Error("warnStyle foreground not set")
	}
	if mp.errorStyle.GetForeground() == nil {
		t.Error("errorStyle foreground not set")
	}
	if mp.successStyle.GetForeground() == nil {
		t.Error("successStyle foreground not set")
	}
}

func TestMessagePrinter_PrintInfo(t *testing.T) {
	mp := NewMessagePrinter()
	mp.PrintInfo("test message")
}

func TestMessagePrinter_PrintWarning(t *testing.T) {
	mp := NewMessagePrinter()
	mp.PrintWarning("test warning")
}

func TestMessagePrinter_PrintError(t *testing.T) {
	mp := NewMessagePrinter()
	mp.PrintError("test error")
}

func TestMessagePrinter_PrintSuccess(t *testing.T) {
	mp := NewMessagePrinter()
	mp.PrintSuccess("test success")
}
