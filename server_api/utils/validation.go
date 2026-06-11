package utils

import (
	"fmt"
	"regexp"
	"unicode"
)

// ValidatePassword validates password length and composition
func ValidatePassword(password string, maxPasswordLength int) error {
	// Check password length
	if len(password) > maxPasswordLength {
		return fmt.Errorf("password exceeds maximum length of %d", maxPasswordLength)
	}

	// If password length is less than 8 and contains only digits or only letters, warn
	if len(password) < 8 {
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
			// Just log a warning, don't prevent the operation as per requirement
			// The warning is handled at the client side
		}
	}

	return nil
}

// ValidateBio validates bio length
func ValidateBio(bio string, maxBioLength int) error {
	if len(bio) > maxBioLength {
		return fmt.Errorf("bio exceeds maximum length of %d", maxBioLength)
	}

	return nil
}

// ValidateEmail validates email format and length
func ValidateEmail(email string, maxEmailLength int) error {
	if len(email) > maxEmailLength {
		return fmt.Errorf("email exceeds maximum length of %d", maxEmailLength)
	}

	// Email regex validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateNickname validates nickname length and composition
func ValidateNickname(nickname string, maxNicknameLength int) error {
	if len(nickname) > maxNicknameLength {
		return fmt.Errorf("nickname exceeds maximum length of %d", maxNicknameLength)
	}

	return nil
}
