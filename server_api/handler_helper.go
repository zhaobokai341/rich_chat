package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"rich_chat/server_api/service"
)

// handleServiceError converts service errors to HTTP responses
func handleServiceError(c *gin.Context, err error) {
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": lp.G("username_or_password_is_invalid"),
		})
	case service.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"message": lp.G("user_not_found"),
		})
	case service.ErrUsernameAlreadyExists:
		c.JSON(http.StatusConflict, gin.H{
			"message": lp.G("username_already_exists"),
		})
	default:
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Service error occurred")
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": lp.G("internal_server_error"),
		})
	}
}
