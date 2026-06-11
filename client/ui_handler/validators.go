package ui_handler

import (
	"errors"
	"unicode"
)

const (
	ALLOW_MAX_LENGTH_OF_USERNAME = 50
	MIN_PASSWORD_LENGTH          = 8
)

// validatePassword checks if a password meets complexity requirements
func validatePassword(password string) error {
	if len(password) < MIN_PASSWORD_LENGTH {
		isOnlyDigits := true
		isOnlyLetters := true

		for _, char := range password {
			if !unicode.IsDigit(char) {
				isOnlyDigits = false
			}
			if !unicode.IsLetter(char) {
				isOnlyLetters = false
			}
		}

		if isOnlyDigits || isOnlyLetters {
			return errors.New("password_too_short")
		}
	}
	return nil
}

// validateUsername checks if a username is valid
func validateUsername(username string, maxLength int) error {
	if username == "" {
		return errors.New("username_cannot_be_empty")
	}
	if len(username) > maxLength {
		return errors.New("username_too_long")
	}
	return nil
}

// validateEmail checks if an email format is valid
func validateEmail(email string) error {
	if email == "" {
		return errors.New("email_cannot_be_empty")
	}
	return nil
}

// validateRequired checks if a string is not empty
func validateRequired(value, fieldName string) error {
	if value == "" {
		return errors.New(fieldName + "_cannot_be_empty")
	}
	return nil
}
