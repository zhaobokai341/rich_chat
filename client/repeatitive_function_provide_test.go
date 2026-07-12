package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintFunction(t *testing.T) {
	tests := []struct {
		name        string
		styleType   string
		text        string
		shouldPanic bool
	}{
		{
			name:        "print info",
			styleType:   "info",
			text:        "test info message",
			shouldPanic: false,
		},
		{
			name:        "print warning",
			styleType:   "warning",
			text:        "test warning message",
			shouldPanic: false,
		},
		{
			name:        "print error",
			styleType:   "error",
			text:        "test error message",
			shouldPanic: false,
		},
		{
			name:        "print success",
			styleType:   "success",
			text:        "test success message",
			shouldPanic: false,
		},
		{
			name:        "print debug",
			styleType:   "debug",
			text:        "test debug message",
			shouldPanic: false,
		},
		{
			name:        "print unknown style",
			styleType:   "unknown",
			text:        "test unknown message",
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					print(tt.styleType, tt.text)
				})
			} else {
				assert.NotPanics(t, func() {
					print(tt.styleType, tt.text)
				})
			}
		})
	}
}
