package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"rich_chat/server_api/database"
)

// TestUserServiceImpl_GetUserProfile tests the GetUserProfile method
func TestUserServiceImpl_GetUserProfile(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)

	userService := NewUserService(mockUserRepo, mockRateLimitRepo, 128, 500, 255, 1000)

	tests := []struct {
		name          string
		userID        int
		expectedUser  *database.UserInfo
		expectedError error
		setupMocks    func()
	}{
		{
			name:   "successful profile retrieval",
			userID: 1,
			expectedUser: &database.UserInfo{
				Username: "testuser",
				Email:    "test@example.com",
				Nickname: "Test User",
			},
			expectedError: nil,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 1).Return(true, nil)
				mockUserRepo.On("GetUserProfile", 1).Return(&database.UserInfo{
					Username: "testuser",
					Email:    "test@example.com",
					Nickname: "Test User",
				}, nil)
			},
		},
		{
			name:          "profile not found",
			userID:        999,
			expectedUser:  nil,
			expectedError: ErrInvalidPassword,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 999).Return(false, nil)
			},
		},
		{
			name:          "error checking user existence",
			userID:        2,
			expectedUser:  nil,
			expectedError: errors.New("failed to check user existence: db error"),
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 2).Return(false, errors.New("db error"))
			},
		},
		{
			name:          "error retrieving profile",
			userID:        3,
			expectedUser:  nil,
			expectedError: errors.New("failed to get user profile: db error"),
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 3).Return(true, nil)
				mockUserRepo.On("GetUserProfile", 3).Return(nil, errors.New("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			profile, err := userService.GetUserProfile(context.Background(), tt.userID)

			if tt.expectedUser != nil {
				assert.NotNil(t, profile)
				assert.Equal(t, tt.expectedUser.Username, profile.Username)
				assert.Equal(t, tt.expectedUser.Email, profile.Email)
				assert.Equal(t, tt.expectedUser.Nickname, profile.Nickname)
			} else {
				assert.Nil(t, profile)
			}

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

// TestUserServiceImpl_UpdateUserProfile tests the UpdateUserProfile method
func TestUserServiceImpl_UpdateUserProfile(t *testing.T) {
	tests := []struct {
		name          string
		req           *UserProfileUpdateRequest
		expectedError error
		setupMocks    func(*database.MockUserRepository, *database.MockRateLimitRepository)
	}{
		{
			name: "successful profile update",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "display_name",
				Value:  "New Display Name",
			},
			expectedError: nil,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 1).Return(&database.User{}, nil)
				mockUserRepo.On("UpdateProfile", 1, "display_name", "New Display Name").Return(nil)
			},
		},
		{
			name: "profile update fails - user not found",
			req: &UserProfileUpdateRequest{
				UserID: 2,
				Key:    "display_name",
				Value:  "Invalid Name",
			},
			expectedError: ErrInvalidPassword,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 2).Return(nil, errors.New("not found"))
				mockRateLimitRepo.On("TrackLoginAttempt", "2", false).Return(nil)
			},
		},
		{
			name: "profile update fails - empty key",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "",
				Value:  "Some value",
			},
			expectedError: ErrInvalidInput,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
		{
			name: "profile update fails - invalid email format",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "email",
				Value:  "invalid-email",
			},
			expectedError: ErrInvalidEmailFormat,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
		{
			name: "profile update fails - email exceeds max length",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "email",
				Value:  string(make([]byte, 300)) + "@example.com",
			},
			expectedError: ErrEmailExceedsMaxLength,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
		{
			name: "profile update fails - bio exceeds max length",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "bio",
				Value:  string(make([]byte, 600)),
			},
			expectedError: ErrBioExceedsMaxLength,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
		{
			name: "successful profile update - nickname",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "nickname",
				Value:  "NewNickname",
			},
			expectedError: nil,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 1).Return(&database.User{}, nil)
				mockUserRepo.On("UpdateProfile", 1, "nickname", "NewNickname").Return(nil)
			},
		},
		{
			name: "successful profile update - unknown field",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "unknown_field",
				Value:  "some value",
			},
			expectedError: nil,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 1).Return(&database.User{}, nil)
				mockUserRepo.On("UpdateProfile", 1, "unknown_field", "some value").Return(nil)
			},
		},
		{
			name: "profile update fails - update profile error",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "nickname",
				Value:  "New Nickname",
			},
			expectedError: errors.New("failed to update user profile: db error"),
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 1).Return(&database.User{}, nil)
				mockUserRepo.On("UpdateProfile", 1, "nickname", "New Nickname").Return(errors.New("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(database.MockUserRepository)
			mockRateLimitRepo := new(database.MockRateLimitRepository)
			userService := NewUserService(mockUserRepo, mockRateLimitRepo, 128, 500, 255, 255)

			tt.setupMocks(mockUserRepo, mockRateLimitRepo)

			err := userService.UpdateUserProfile(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
			mockRateLimitRepo.AssertExpectations(t)
		})
	}
}

// TestUserServiceImpl_DeleteUser tests the DeleteUser method
func TestUserServiceImpl_DeleteUser(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		req           *DeleteUserRequest
		expectedError error
		setupMocks    func(*database.MockUserRepository, *database.MockRateLimitRepository)
	}{
		{
			name: "successful deletion",
			req: &DeleteUserRequest{
				UserID:   1,
				Password: "password123",
			},
			expectedError: nil,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(false, nil)
				mockUserRepo.On("FindByID", 1).Return(&database.User{PasswordHash: string(hashedPassword)}, nil)
				mockRateLimitRepo.On("TrackLoginAttempt", "1", true).Return(nil)
				mockUserRepo.On("DeleteUser", 1).Return(nil)
			},
		},
		{
			name: "deletion fails - invalid password",
			req: &DeleteUserRequest{
				UserID:   1,
				Password: "wrongpassword",
			},
			expectedError: ErrInvalidPassword,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(false, nil)
				mockUserRepo.On("FindByID", 1).Return(&database.User{PasswordHash: string(hashedPassword)}, nil)
				mockRateLimitRepo.On("TrackLoginAttempt", "1", false).Return(nil)
			},
		},
		{
			name: "deletion fails - user not found",
			req: &DeleteUserRequest{
				UserID:   999,
				Password: "password123",
			},
			expectedError: ErrInvalidPassword,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockRateLimitRepo.On("CheckAccountLocked", "999").Return(false, nil)
				mockUserRepo.On("FindByID", 999).Return(nil, errors.New("not found"))
				mockRateLimitRepo.On("TrackLoginAttempt", "999", false).Return(nil)
			},
		},
		{
			name: "deletion fails - account locked",
			req: &DeleteUserRequest{
				UserID:   1,
				Password: "password123",
			},
			expectedError: ErrAccountLocked,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(true, nil)
			},
		},
		{
			name: "deletion fails - empty password",
			req: &DeleteUserRequest{
				UserID:   1,
				Password: "",
			},
			expectedError: ErrInvalidInput,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(database.MockUserRepository)
			mockRateLimitRepo := new(database.MockRateLimitRepository)
			userService := NewUserService(mockUserRepo, mockRateLimitRepo, 128, 500, 255, 255)

			tt.setupMocks(mockUserRepo, mockRateLimitRepo)

			err := userService.DeleteUser(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
			mockRateLimitRepo.AssertExpectations(t)
		})
	}
}

// TestUserServiceImpl_CheckAccountLocked tests the CheckAccountLocked method
func TestUserServiceImpl_CheckAccountLocked(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)

	userService := NewUserService(mockUserRepo, mockRateLimitRepo, 128, 500, 255, 255)

	tests := []struct {
		name             string
		identifier       string
		expectedIsLocked bool
		setupMocks       func()
	}{
		{
			name:             "account not locked",
			identifier:       "activeuser",
			expectedIsLocked: false,
			setupMocks: func() {
				mockRateLimitRepo.On("CheckAccountLocked", "activeuser").Return(false, nil)
			},
		},
		{
			name:             "account locked",
			identifier:       "lockeduser",
			expectedIsLocked: true,
			setupMocks: func() {
				mockRateLimitRepo.On("CheckAccountLocked", "lockeduser").Return(true, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			isLocked := userService.CheckAccountLocked(tt.identifier)

			assert.Equal(t, tt.expectedIsLocked, isLocked)

			mockRateLimitRepo.AssertExpectations(t)
		})
	}

	t.Run("rate limit repo is nil", func(t *testing.T) {
		userServiceWithNilRepo := NewUserService(mockUserRepo, nil, 128, 500, 255, 255)
		isLocked := userServiceWithNilRepo.CheckAccountLocked("anyuser")
		assert.False(t, isLocked)
	})
}

// TestUserServiceImpl_CheckUserExists tests the CheckUserExists method
func TestUserServiceImpl_CheckUserExists(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)

	userService := NewUserService(mockUserRepo, mockRateLimitRepo, 128, 500, 255, 255)

	tests := []struct {
		name          string
		userID        int
		expectedBool  bool
		expectedError error
		setupMocks    func()
	}{
		{
			name:          "user exists",
			userID:        1,
			expectedBool:  true,
			expectedError: nil,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 1).Return(true, nil)
			},
		},
		{
			name:          "user does not exist",
			userID:        999,
			expectedBool:  false,
			expectedError: nil,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 999).Return(false, nil)
			},
		},
		{
			name:          "check fails",
			userID:        500,
			expectedBool:  false,
			expectedError: errors.New("check failed"),
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 500).Return(false, errors.New("check failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			exists, err := userService.CheckUserExists(tt.userID)

			assert.Equal(t, tt.expectedBool, exists)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

// TestUserServiceImpl_ChangeUserPassword tests the ChangeUserPassword method
func TestUserServiceImpl_ChangeUserPassword(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		req           *ChangePasswordRequest
		expectedError error
		setupMocks    func(*database.MockUserRepository, *database.MockRateLimitRepository)
	}{
		{
			name: "successful password change",
			req: &ChangePasswordRequest{
				UserID:      1,
				OldPassword: "oldpassword",
				NewPassword: "newpassword123",
			},
			expectedError: nil,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 1).Return(&database.User{PasswordHash: string(hashedPassword)}, nil)
				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(false, nil)
				mockUserRepo.On("UpdatePassword", 1, mock.AnythingOfType("string")).Return(nil)
			},
		},
		{
			name: "password change fails - invalid old password",
			req: &ChangePasswordRequest{
				UserID:      1,
				OldPassword: "wrongpassword",
				NewPassword: "newpassword123",
			},
			expectedError: ErrInvalidPassword,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 1).Return(&database.User{PasswordHash: string(hashedPassword)}, nil)
				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(false, nil)
				mockRateLimitRepo.On("TrackLoginAttempt", "1", false).Return(nil)
			},
		},
		{
			name: "password change fails - user not found",
			req: &ChangePasswordRequest{
				UserID:      999,
				OldPassword: "oldpassword",
				NewPassword: "newpassword123",
			},
			expectedError: ErrInvalidPassword,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 999).Return(nil, errors.New("not found"))
				mockRateLimitRepo.On("TrackLoginAttempt", "999", false).Return(nil)
			},
		},
		{
			name: "password change fails - account locked",
			req: &ChangePasswordRequest{
				UserID:      1,
				OldPassword: "oldpassword",
				NewPassword: "newpassword123",
			},
			expectedError: ErrAccountLocked,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 1).Return(&database.User{PasswordHash: string(hashedPassword)}, nil)
				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(true, nil)
			},
		},
		{
			name: "password change fails - empty new password",
			req: &ChangePasswordRequest{
				UserID:      1,
				OldPassword: "oldpassword",
				NewPassword: "",
			},
			expectedError: ErrInvalidInput,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
		{
			name: "password change fails - empty old password",
			req: &ChangePasswordRequest{
				UserID:      1,
				OldPassword: "",
				NewPassword: "newpassword123",
			},
			expectedError: ErrInvalidInput,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
		{
			name: "password change fails - password exceeds max length",
			req: &ChangePasswordRequest{
				UserID:      1,
				OldPassword: "oldpassword",
				NewPassword: string(make([]byte, 200)),
			},
			expectedError: ErrPasswordExceedsMaxLength,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(database.MockUserRepository)
			mockRateLimitRepo := new(database.MockRateLimitRepository)
			userService := NewUserService(mockUserRepo, mockRateLimitRepo, 128, 500, 255, 10)

			tt.setupMocks(mockUserRepo, mockRateLimitRepo)

			err := userService.ChangeUserPassword(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
			mockRateLimitRepo.AssertExpectations(t)
		})
	}
}
