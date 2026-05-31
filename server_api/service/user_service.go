package service

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"rich_chat/server_api/database"
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
		return nil, ErrUserNotFound
	}

	// Get user profile
	profile, err := s.userRepo.GetUserProfile(userID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to get user profile")
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return profile, nil
}

// UpdateUserProfile updates user profile information
func (s *UserServiceImpl) UpdateUserProfile(ctx context.Context, req *UserProfileUpdateRequest) error {
	// Validate input
	if req.Key == "" || req.Value == "" {
		return ErrInvalidInput
	}

	// Check if user exists
	exists, err := s.userRepo.ExistsByID(req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to check user existence")
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return ErrUserNotFound
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

// ChangeUserPassword changes a user's password
func (s *UserServiceImpl) ChangeUserPassword(ctx context.Context, req *ChangePasswordRequest) error {
	// Validate input
	if req.NewPassword == "" || req.OldPassword == "" {
		return ErrInvalidInput
	}

	// Check if user exists
	exists, err := s.userRepo.ExistsByID(req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to check user existence")
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	// Check if account is locked
	identifier := fmt.Sprintf("%d", req.UserID)
	if s.rateLimitRepo != nil {
		locked, _ := s.rateLimitRepo.CheckAccountLocked(identifier)
		if locked {
			log.WithFields(log.Fields{
				"user_id": req.UserID,
			}).Warning("Change password attempt on locked account")
			return ErrAccountLocked
		}
	}

	// Find user
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to find user")
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.OldPassword),
	); err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
		}).Warning("Invalid old password")

		// Track failed attempt
		if s.rateLimitRepo != nil {
			_ = s.rateLimitRepo.TrackLoginAttempt(identifier, false)
		}
		return ErrInvalidPassword
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to hash password")
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Hash new password
	err = s.userRepo.UpdatePassword(req.UserID, string(passwordHash))
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Error("Failed to update user password")

		// Track failed attempt
		if s.rateLimitRepo != nil {
			_ = s.rateLimitRepo.TrackLoginAttempt(identifier, false)
		}
		return fmt.Errorf("failed to update user password: %w", err)
	}

	log.WithFields(log.Fields{
		"user_id": req.UserID,
	}).Info("User password changed successfully")
	return nil
}

// DeleteUser deletes a user account after verifying password
func (s *UserServiceImpl) DeleteUser(ctx context.Context, req *DeleteUserRequest) error {
	// Validate input
	if req.Password == "" {
		return ErrInvalidInput
	}

	identifier := fmt.Sprintf("%d", req.UserID)

	// Check if account is locked
	if s.rateLimitRepo != nil {
		locked, _ := s.rateLimitRepo.CheckAccountLocked(identifier)
		if locked {
			log.WithFields(log.Fields{
				"user_id": req.UserID,
			}).Warning("Delete attempt on locked account")
			return ErrAccountLocked
		}
	}

	// Find user
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
			"error":   err.Error(),
		}).Warning("User not found during delete attempt")

		// Track failed attempt
		if s.rateLimitRepo != nil {
			_ = s.rateLimitRepo.TrackLoginAttempt(identifier, false)
		}

		return ErrUserNotFound
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		log.WithFields(log.Fields{
			"user_id": req.UserID,
		}).Warning("Invalid password during delete attempt")

		// Track failed attempt
		if s.rateLimitRepo != nil {
			_ = s.rateLimitRepo.TrackLoginAttempt(identifier, false)
		}

		return ErrInvalidPassword
	}

	// Track successful verification
	if s.rateLimitRepo != nil {
		_ = s.rateLimitRepo.TrackLoginAttempt(identifier, true)
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
