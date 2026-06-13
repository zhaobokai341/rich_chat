package service

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"rich_chat/server_api/database"
	"rich_chat/server_api/utils"
)

// UserServiceImpl implements UserService
type UserServiceImpl struct {
	userRepo      database.UserRepository
	rateLimitRepo database.RateLimitRepository
}

// NewUserService creates a new user service
func NewUserService(
	userRepo database.UserRepository,
	rateLimitRepo database.RateLimitRepository,
) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo:      userRepo,
		rateLimitRepo: rateLimitRepo,
	}
}

// GetUserProfile retrieves user profile information
func (s *UserServiceImpl) GetUserProfile(ctx context.Context, userID int) (*database.UserInfo, error) {
	// Check if user exists
	exists, err := s.userRepo.ExistsByID(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to check user existence")
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		// To prevent user enumeration, return the same error type as for auth failures
		return nil, ErrInvalidPassword
	}

	// Get user profile
	userInfo, err := s.userRepo.GetUserProfile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get user profile")
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return userInfo, nil
}

// UpdateUserProfile updates user profile information
func (s *UserServiceImpl) UpdateUserProfile(ctx context.Context, req *UserProfileUpdateRequest) error {
	// Validate input
	if req.Key == "" {
		return ErrInvalidInput
	}

	// Validate value based on the key
	switch req.Key {
	case "email":
		if err := utils.ValidateEmail(req.Value, ALLOW_MAX_LENGTH_OF_EMAIL); err != nil {
			log.WithFields(log.Fields{
				"user_id": req.UserID,
				"key":     req.Key,
				"value":   req.Value,
				"error":   err.Error(),
			}).Warning("Invalid email format or length")

			// Determine specific error type
			if err.Error() == "invalid email format" {
				return ErrInvalidEmailFormat
			} else if err.Error() == fmt.Sprintf("email exceeds maximum length of %d", ALLOW_MAX_LENGTH_OF_EMAIL) {
				return ErrEmailExceedsMaxLength
			}
			return ErrInvalidInput
		}
	case "bio":
		if err := utils.ValidateBio(req.Value, ALLOW_MAX_LENGTH_OF_BIO); err != nil {
			log.WithFields(log.Fields{
				"user_id": req.UserID,
				"key":     req.Key,
				"value":   req.Value,
				"error":   err.Error(),
			}).Warning("Bio exceeds maximum length")
			return ErrBioExceedsMaxLength
		}
	case "nickname":
		if err := utils.ValidateNickname(req.Value, ALLOW_MAX_LENGTH_OF_NICKNAME); err != nil {
			log.WithFields(log.Fields{
				"user_id": req.UserID,
				"key":     req.Key,
				"value":   req.Value,
				"error":   err.Error(),
			}).Warning("Invalid nickname format")
			return ErrInvalidNicknameFormat
		}
	default:
		// Allow other fields that might be added in the future
	}

	// Check if user exists
	_, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Warning("User not found during profile update attempt")

		identifier := fmt.Sprintf("%d", req.UserID)
		// Track failed attempt - treat non-existent user as invalid attempt
		if s.rateLimitRepo != nil {
			_ = s.rateLimitRepo.TrackLoginAttempt(identifier, false)
		}

		// Return ErrInvalidPassword to prevent user enumeration
		return ErrInvalidPassword
	}

	// Update profile
	err = s.userRepo.UpdateProfile(req.UserID, req.Key, req.Value)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"key":     req.Key,
			"error":   err.Error(),
		}).Error("Failed to update user profile")
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id": req.UserID,
		"key":     req.Key,
	}).Info("User profile updated successfully")

	return nil
}

// ChangeUserPassword changes user password
func (s *UserServiceImpl) ChangeUserPassword(ctx context.Context, req *ChangePasswordRequest) error {
	// Validate input
	if req.OldPassword == "" || req.NewPassword == "" {
		return ErrInvalidInput
	}

	// Validate new password length and composition
	if err := utils.ValidatePassword(req.NewPassword, ALLOW_MAX_LENGTH_OF_PASSWORD); err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Warning("Invalid new password format")
		return ErrPasswordExceedsMaxLength
	}

	// Check if user exists and get user info
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Warning("User not found during password change attempt")

		identifier := fmt.Sprintf("%d", req.UserID)
		// Track failed attempt
		if s.rateLimitRepo != nil {
			_ = s.rateLimitRepo.TrackLoginAttempt(identifier, false)
		}

		// Return ErrInvalidPassword to prevent user enumeration
		return ErrInvalidPassword
	}

	// Check if account is locked
	identifier := fmt.Sprintf("%d", req.UserID)
	if s.CheckAccountLocked(identifier) {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
		}).Warning("Account is locked during password change attempt")
		return ErrAccountLocked
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
		}).Warning("Invalid old password during password change")

		// Track failed attempt
		if s.rateLimitRepo != nil {
			_ = s.rateLimitRepo.TrackLoginAttempt(identifier, false)
		}

		return ErrInvalidPassword
	}

	// Hash new password
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to hash new password")
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	err = s.userRepo.UpdatePassword(req.UserID, string(newPasswordHash))
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to update password")
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Track successful attempt (reset counter)
	if s.rateLimitRepo != nil {
		_ = s.rateLimitRepo.TrackLoginAttempt(identifier, true)
	}

	log.WithFields(log.Fields{
		"user_id": req.UserID,
	}).Info("User password changed successfully")

	return nil
}

// DeleteUser deletes a user account
func (s *UserServiceImpl) DeleteUser(ctx context.Context, req *DeleteUserRequest) error {
	// Validate input
	if req.Password == "" {
		return ErrInvalidInput
	}

	// Check if user exists and get user info
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Warning("User not found during account deletion attempt")

		identifier := fmt.Sprintf("%d", req.UserID)
		// Track failed attempt
		if s.rateLimitRepo != nil {
			_ = s.rateLimitRepo.TrackLoginAttempt(identifier, false)
		}

		// Return ErrInvalidPassword to prevent user enumeration
		return ErrInvalidPassword
	}

	// Check if account is locked
	identifier := fmt.Sprintf("%d", req.UserID)
	if s.CheckAccountLocked(identifier) {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
		}).Warning("Account is locked during deletion attempt")
		return ErrAccountLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
		}).Warning("Invalid password during account deletion")

		// Track failed attempt
		if s.rateLimitRepo != nil {
			_ = s.rateLimitRepo.TrackLoginAttempt(identifier, false)
		}

		return ErrInvalidPassword
	}

	// Delete user
	err = s.userRepo.DeleteUser(req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to delete user")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id": req.UserID,
	}).Info("User deleted successfully")

	return nil
}

// CheckAccountLocked checks if an account is locked
func (s *UserServiceImpl) CheckAccountLocked(identifier string) bool {
	if s.rateLimitRepo == nil {
		return false
	}
	locked, _ := s.rateLimitRepo.CheckAccountLocked(identifier)
	return locked
}

// CheckUserExists checks if a user exists by ID
func (s *UserServiceImpl) CheckUserExists(userID int) (bool, error) {
	return s.userRepo.ExistsByID(userID)
}

// GetUserBasicInfo retrieves basic user information
func (s *UserServiceImpl) GetUserBasicInfo(ctx context.Context, userID int) (*database.UserBasicInfo, error) {
	// Get user basic info
	userInfo, err := s.userRepo.GetUserBasicInfo(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get user basic info")
		return nil, fmt.Errorf("failed to get user basic info: %w", err)
	}

	return userInfo, nil
}
