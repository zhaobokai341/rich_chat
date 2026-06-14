package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"rich_chat/server_api/service"
)

// handleServiceError converts service errors to HTTP responses
func handleServiceError(c *gin.Context, err error) {
	// Get language pack for this request
	lp := getLanguagePackFromContext(c)

	switch err {
	case service.ErrInvalidInput:
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("username_and_password_required"),
		})
	case service.ErrInvalidToken:
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": lp.G("invalid_or_expired_verification_token"),
		})
	case service.ErrAccountLocked:
		c.JSON(http.StatusTooManyRequests, gin.H{
			"message": lp.G("account_locked_try_later"),
		})
	case service.ErrInvalidPassword:
		fallthrough
	case service.ErrUserNotFound:
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": lp.G("authentication_failed"),
		})
	case service.ErrUsernameAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{
			"message": lp.G("username_already_exists"),
		})
	case service.ErrInvalidNicknameFormat:
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("nickname_invalid_format"),
		})
	case service.ErrInvalidEmailFormat:
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warning("Invalid email format")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_input"),
		})
	case service.ErrEmailExceedsMaxLength:
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warning("Email length validation error")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_input"),
		})
	case service.ErrBioExceedsMaxLength:
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warning("Bio length validation error")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_input"),
		})
	case service.ErrPasswordExceedsMaxLength:
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Warning("Password length validation error")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("invalid_input"),
		})
	default:
		// Log as error for genuine server-side issues
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Service error occurred")
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": lp.G("internal_server_error"),
		})
	}
}
