package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"rich_chat/server_api/service"
)

// Index page
func (api *WebServerAPI) Index(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to Rich Chat!")
}

// HealthCheck returns basic health status
func (api *WebServerAPI) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": VERSION,
	})
}

// ReadinessCheck checks if all dependencies are ready
func (api *WebServerAPI) ReadinessCheck(c *gin.Context) {
	// Simple readiness check - verify services are initialized
	ready := api.authService != nil &&
		api.userService != nil &&
		api.chatService != nil &&
		api.tokenService != nil

	if ready {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
		})
	}
}

// Get verification token for sensitive operations - Refactored to use TokenService
func (api *WebServerAPI) GetVerifyToken(c *gin.Context) {
	// Get language pack for this request
	lp := getLanguagePackFromContext(c)

	// Generate verification token using service
	token, err := api.tokenService.GenerateVerificationToken()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to generate verification token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": lp.G("internal_server_error"),
		})
		return
	}

	// Store the verification token in Redis
	err = api.tokenService.StoreVerificationToken(token, VERIFY_TOKEN_EXPIRE_TIME)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to store verification token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": lp.G("internal_server_error"),
		})
		return
	}

	log.WithFields(log.Fields{
		"token": token,
	}).Info("Verification token generated and stored")

	c.JSON(http.StatusOK, gin.H{
		"verify_token": token,
	})
}

// Login user - Refactored to use AuthService
func (api *WebServerAPI) Login(c *gin.Context) {
	// Parse request
	req := &service.LoginRequest{
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
	}

	// Call service
	resp, err := api.authService.Login(c.Request.Context(), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Return success response
	log.WithFields(log.Fields{
		"user_id":  resp.UserID,
		"username": req.Username,
	}).Info("User login successful")

	c.JSON(http.StatusOK, gin.H{
		"user_token": resp.UserToken,
		"user_id":    resp.UserID,
	})
}

// Register user - Refactored to use AuthService
func (api *WebServerAPI) Register(c *gin.Context) {
	// Parse request
	req := &service.RegisterRequest{
		Username: c.PostForm("username"),
		Password: c.PostForm("password"),
	}

	// Call service
	resp, err := api.authService.Register(c.Request.Context(), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Return success response
	log.WithFields(log.Fields{
		"user_id":  resp.UserID,
		"username": req.Username,
	}).Info("User registration successful")

	c.JSON(http.StatusOK, gin.H{
		"user_token": resp.UserToken,
		"user_id":    resp.UserID,
	})
}

// Delete user account - Refactored to use UserService
func (api *WebServerAPI) DeleteUser(c *gin.Context) {
	// Get language pack for this request
	lp := getLanguagePackFromContext(c)

	// Parse user_id from URL parameter
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_user_id_format"),
		})
		return
	}

	// Parse request
	req := &service.DeleteUserRequest{
		UserID:   userID,
		Password: c.PostForm("user_password"),
	}

	// Call service
	err = api.userService.DeleteUser(c.Request.Context(), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Return success response
	log.WithFields(log.Fields{
		"user_id": userID,
	}).Info("User deleted successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": lp.G("user_deleted_successfully"),
	})
}

// Get user info - Refactored to use UserService
func (api *WebServerAPI) GetUserProfile(c *gin.Context) {
	// Get language pack for this request
	lp := getLanguagePackFromContext(c)

	// Parse user_id from URL parameter
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_user_id_format"),
		})
		return
	}

	// Call service
	profile, err := api.userService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": lp.G("user_info"),
		"data":    profile,
	})
}

// Change user info - Refactored to use UserService
func (api *WebServerAPI) ChangeUserProfile(c *gin.Context) {
	// Get language pack for this request
	lp := getLanguagePackFromContext(c)

	// Parse user_id from URL parameter
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_user_id_format"),
		})
		return
	}

	// Parse request
	req := &service.UserProfileUpdateRequest{
		UserID: userID,
		Key:    c.PostForm("user_info_key"),
		Value:  c.PostForm("user_info_value"),
	}

	// Call service
	err = api.userService.UpdateUserProfile(c.Request.Context(), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Return success response
	log.WithFields(log.Fields{
		"user_id": userID,
		"key":     req.Key,
	}).Info("User profile updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": lp.G("user_info_changed_successfully"),
	})
}

// Change user password
func (api *WebServerAPI) ChangeUserPassword(c *gin.Context) {
	// Get language pack for this request
	lp := getLanguagePackFromContext(c)

	// Parse user_id from URL parameter
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_user_id_format"),
		})
		return
	}

	old_password := c.PostForm("old_password")
	new_password := c.PostForm("new_password")

	// Parse request
	req := &service.ChangePasswordRequest{
		UserID:      userID,
		OldPassword: old_password,
		NewPassword: new_password,
	}

	// Call service
	err = api.userService.ChangeUserPassword(c.Request.Context(), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Return success response
	log.WithFields(log.Fields{
		"user_id": userID,
	}).Info("User password changed successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": lp.G("user_password_changed_successfully"),
	})
}

// GetUserBasicInfo retrieves basic user info for chat
func (api *WebServerAPI) GetUserBasicInfo(c *gin.Context) {
	lp := getLanguagePackFromContext(c)

	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_user_id_format"),
		})
		return
	}

	userInfo, err := api.userService.GetUserBasicInfo(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": userInfo,
	})
}

// CreateChatSession creates a new direct chat session
func (api *WebServerAPI) CreateChatSession(c *gin.Context) {
	lp := getLanguagePackFromContext(c)

	// Parse user_id from URL parameter (the authenticated user)
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_user_id_format"),
		})
		return
	}

	// Parse request body
	var req struct {
		RecipientID int `json:"recipient_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_request_body"),
		})
		return
	}

	if req.RecipientID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("recipient_id_required"),
		})
		return
	}

	// Check if recipient exists
	exists, err := api.userService.CheckUserExists(req.RecipientID)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"message": lp.G("user_not_found"),
		})
		return
	}

	sessionResp, err := api.chatService.CreateSession(c.Request.Context(), &service.CreateSessionRequest{
		CreatorID:   userID,
		SessionType: "direct",
		MemberIDs:   []int{req.RecipientID},
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionResp.SessionID,
		"message":    lp.G("chat_session_created"),
	})
}

// GetUserSessions retrieves all chat sessions for a user
func (api *WebServerAPI) GetUserSessions(c *gin.Context) {
	lp := getLanguagePackFromContext(c)

	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_user_id_format"),
		})
		return
	}

	sessions, err := api.chatService.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// GetUserPublicKeyHandler retrieves a user's public encryption key
func (api *WebServerAPI) GetUserPublicKeyHandler(c *gin.Context) {
	lp := getLanguagePackFromContext(c)

	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_user_id_format"),
		})
		return
	}

	keyResp, err := api.chatService.GetUserPublicKey(c.Request.Context(), &service.GetUserKeyRequest{
		UserID: userID,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    keyResp.UserID,
		"public_key": keyResp.PublicKey,
		"algorithm":  keyResp.Algorithm,
	})
}

// StoreUserKeyHandler stores a user's encryption key
func (api *WebServerAPI) StoreUserKeyHandler(c *gin.Context) {
	lp := getLanguagePackFromContext(c)

	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_user_id_format"),
		})
		return
	}

	var req struct {
		PublicKey    string `json:"public_key"`
		KeyAlgorithm string `json:"key_algorithm"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_request_body"),
		})
		return
	}

	if req.PublicKey == "" || req.KeyAlgorithm == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("public_key_and_algorithm_required"),
		})
		return
	}

	err = api.chatService.StoreUserKey(c.Request.Context(), &service.StoreUserKeyRequest{
		UserID:       userID,
		PublicKey:    req.PublicKey,
		KeyAlgorithm: req.KeyAlgorithm,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": lp.G("key_stored_successfully"),
	})
}
