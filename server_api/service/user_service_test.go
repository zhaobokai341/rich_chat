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

	userService := NewUserService(mockUserRepo, mockRateLimitRepo)

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
			expectedError: ErrInvalidPassword, // Changed to prevent user enumeration
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
			name: "successful profile update - nickname",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "nickname",
				Value:  "NewNickname",
			},
			expectedError: nil,
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 1).Return(&database.User{ID: 1, Username: "testuser"}, nil)
				mockUserRepo.On("UpdateProfile", 1, "nickname", "NewNickname").Return(nil)
			},
		},
		{
			name: "profile update fails - user not found",
			req: &UserProfileUpdateRequest{
				UserID: 2,
				Key:    "nickname",
				Value:  "Invalid Name",
			},
			expectedError: ErrInvalidPassword, // Changed to prevent user enumeration
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 2).Return(nil, errors.New("user not found"))
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
			name: "profile update fails - empty value",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "nickname",
				Value:  "",
			},
			expectedError: nil, // Empty value is now allowed, will be handled by database layer
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 1).Return(&database.User{ID: 1, Username: "testuser"}, nil)
				mockUserRepo.On("UpdateProfile", 1, "nickname", "").Return(nil)
			},
		},
		{
			name: "profile update fails - nickname exceeds max length",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "nickname",
				Value:  string(make([]byte, ALLOW_MAX_LENGTH_OF_NICKNAME+1)),
			},
			expectedError: ErrInvalidNicknameFormat,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
		{
			name: "profile update fails - bio exceeds max length",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "bio",
				Value:  string(make([]byte, ALLOW_MAX_LENGTH_OF_BIO+1)),
			},
			expectedError: ErrBioExceedsMaxLength,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
		{
			name: "profile update fails - email exceeds max length",
			req: &UserProfileUpdateRequest{
				UserID: 1,
				Key:    "email",
				Value:  string(make([]byte, ALLOW_MAX_LENGTH_OF_EMAIL+1)),
			},
			expectedError: ErrEmailExceedsMaxLength,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
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
				mockUserRepo.On("FindByID", 1).Return(&database.User{ID: 1, Username: "testuser"}, nil)
				mockUserRepo.On("UpdateProfile", 1, "nickname", "New Nickname").Return(errors.New("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(database.MockUserRepository)
			mockRateLimitRepo := new(database.MockRateLimitRepository)
			userService := NewUserService(mockUserRepo, mockRateLimitRepo)

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
				mockUserRepo.On("FindByID", 1).Return(&database.User{PasswordHash: string(hashedPassword)}, nil)
				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(false, nil)
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
				mockUserRepo.On("FindByID", 1).Return(&database.User{PasswordHash: string(hashedPassword)}, nil)
				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(false, nil)
				mockRateLimitRepo.On("TrackLoginAttempt", "1", false).Return(nil)
			},
		},
		{
			name: "deletion fails - user not found",
			req: &DeleteUserRequest{
				UserID:   999,
				Password: "password123",
			},
			expectedError: ErrInvalidPassword, // Changed to prevent user enumeration
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 999).Return(nil, errors.New("user not found"))
				mockRateLimitRepo.On("TrackLoginAttempt", "999", false).Return(nil)
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
		{
			name: "deletion fails - delete error",
			req: &DeleteUserRequest{
				UserID:   1,
				Password: "password123",
			},
			expectedError: errors.New("failed to delete user: db error"),
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 1).Return(&database.User{PasswordHash: string(hashedPassword)}, nil)
				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(false, nil)
				mockUserRepo.On("DeleteUser", 1).Return(errors.New("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(database.MockUserRepository)
			mockRateLimitRepo := new(database.MockRateLimitRepository)
			userService := NewUserService(mockUserRepo, mockRateLimitRepo)

			tt.setupMocks(mockUserRepo, mockRateLimitRepo)

			err := userService.DeleteUser(context.Background(), tt.req)

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

// TestUserServiceImpl_CheckAccountLocked tests the CheckAccountLocked method
func TestUserServiceImpl_CheckAccountLocked(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)

	userService := NewUserService(mockUserRepo, mockRateLimitRepo)

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
		userServiceWithNilRepo := NewUserService(mockUserRepo, nil)
		isLocked := userServiceWithNilRepo.CheckAccountLocked("anyuser")
		assert.False(t, isLocked)
	})
}

// TestUserServiceImpl_CheckUserExists tests the CheckUserExists method
func TestUserServiceImpl_CheckUserExists(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)

	userService := NewUserService(mockUserRepo, mockRateLimitRepo)

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
				mockRateLimitRepo.On("TrackLoginAttempt", "1", true).Return(nil)
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
			expectedError: ErrInvalidPassword, // Changed to prevent user enumeration
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 999).Return(nil, errors.New("user not found"))
				mockRateLimitRepo.On("TrackLoginAttempt", "999", false).Return(nil)
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
				NewPassword: string(make([]byte, ALLOW_MAX_LENGTH_OF_PASSWORD+1)),
			},
			expectedError: ErrPasswordExceedsMaxLength,
			setupMocks:    func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {},
		},
		{
			name: "password change fails - update password error",
			req: &ChangePasswordRequest{
				UserID:      1,
				OldPassword: "oldpassword",
				NewPassword: "newpassword123",
			},
			expectedError: errors.New("failed to update password: db error"),
			setupMocks: func(mockUserRepo *database.MockUserRepository, mockRateLimitRepo *database.MockRateLimitRepository) {
				mockUserRepo.On("FindByID", 1).Return(&database.User{PasswordHash: string(hashedPassword)}, nil)
				mockRateLimitRepo.On("CheckAccountLocked", "1").Return(false, nil)
				mockUserRepo.On("UpdatePassword", 1, mock.AnythingOfType("string")).Return(errors.New("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(database.MockUserRepository)
			mockRateLimitRepo := new(database.MockRateLimitRepository)
			userService := NewUserService(mockUserRepo, mockRateLimitRepo)

			tt.setupMocks(mockUserRepo, mockRateLimitRepo)

			err := userService.ChangeUserPassword(context.Background(), tt.req)

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

// TestUserServiceImpl_GetUserBasicInfo tests the GetUserBasicInfo method
func TestUserServiceImpl_GetUserBasicInfo(t *testing.T) {
	mockUserRepo := new(database.MockUserRepository)
	mockRateLimitRepo := new(database.MockRateLimitRepository)

	userService := NewUserService(mockUserRepo, mockRateLimitRepo)

	tests := []struct {
		name          string
		userID        int
		expectedInfo  *database.UserBasicInfo
		expectedError error
		setupMocks    func()
	}{
		{
			name:   "successful basic info retrieval",
			userID: 1,
			expectedInfo: &database.UserBasicInfo{
				Username: "testuser",
				Nickname: "Test User",
			},
			expectedError: nil,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 1).Return(true, nil)
				mockUserRepo.On("GetUserBasicInfo", 1).Return(&database.UserBasicInfo{
					Username: "testuser",
					Nickname: "Test User",
				}, nil)
			},
		},
		{
			name:          "user not found",
			userID:        999,
			expectedInfo:  nil,
			expectedError: ErrUserNotFound,
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 999).Return(false, nil)
			},
		},
		{
			name:          "error checking user existence",
			userID:        2,
			expectedInfo:  nil,
			expectedError: errors.New("failed to check user existence: db error"),
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 2).Return(false, errors.New("db error"))
			},
		},
		{
			name:          "error retrieving basic info",
			userID:        3,
			expectedInfo:  nil,
			expectedError: errors.New("failed to get user basic info: db error"),
			setupMocks: func() {
				mockUserRepo.On("ExistsByID", 3).Return(true, nil)
				mockUserRepo.On("GetUserBasicInfo", 3).Return(nil, errors.New("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			info, err := userService.GetUserBasicInfo(context.Background(), tt.userID)

			if tt.expectedInfo != nil {
				assert.NotNil(t, info)
				assert.Equal(t, tt.expectedInfo.Username, info.Username)
				assert.Equal(t, tt.expectedInfo.Nickname, info.Nickname)
			} else {
				assert.Nil(t, info)
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
