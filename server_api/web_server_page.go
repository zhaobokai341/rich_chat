package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"rich_chat/server_api/service"
)

// Index page
func (api *WebServerApi) Index(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to Rich Chat!")
}

// Get verification token for sensitive operations - Refactored to use TokenService
func (api *WebServerApi) GetVerifyToken(c *gin.Context) {
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
func (api *WebServerApi) Login(c *gin.Context) {
	// Parse request
	req := &service.LoginRequest{
		Username:    c.PostForm("username"),
		Password:    c.PostForm("password"),
		VerifyToken: c.PostForm("verify_token"),
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
func (api *WebServerApi) Register(c *gin.Context) {
	// Parse request
	req := &service.RegisterRequest{
		Username:    c.PostForm("username"),
		Password:    c.PostForm("password"),
		VerifyToken: c.PostForm("verify_token"),
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
func (api *WebServerApi) DeleteUser(c *gin.Context) {
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
		UserID:      userID,
		Password:    c.PostForm("user_password"),
		VerifyToken: c.PostForm("verify_token"),
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
func (api *WebServerApi) GetUserProfile(c *gin.Context) {
	// Parse user_id from URL parameter
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_user_id_format"),
		})
		return
	}

	// Validate verification token
	verifyToken := c.Query("verify_token")
	if err := api.tokenService.ValidateAndConsumeToken(verifyToken); err != nil {
		handleServiceError(c, err)
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
func (api *WebServerApi) ChangeUserProfile(c *gin.Context) {
	// Parse user_id from URL parameter
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_user_id_format"),
		})
		return
	}

	// Validate verification token
	verifyToken := c.PostForm("verify_token")
	if err := api.tokenService.ValidateAndConsumeToken(verifyToken); err != nil {
		handleServiceError(c, err)
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
