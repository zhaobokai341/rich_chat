package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// check client if it is safe
func safe_check() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.GetHeader("User-Agent"), ALLOW_USER_AGENT) {
			log_with_ctx(c, "warn", "User-Agent is not allowed")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		client_version := c.GetHeader("User-Agent")[len(ALLOW_USER_AGENT)+1:]
		if client_version != VERSION {
			log_with_ctx(c, "warn", "Version mismatch: %s != %s", VERSION, client_version)
			c.AbortWithStatus(http.StatusExpectationFailed)
			return
		}
		usr_token := c.GetHeader("user_token")
		if usr_token != "" {
			usr_id := c.GetHeader("user_id")
			if usr_id == "" {
				log_with_ctx(c, "warn", "user_id is empty")
				c.AbortWithStatus(http.StatusExpectationFailed)
				return
			}
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(usr_token, claims,
				func(token *jwt.Token) (interface{}, error) {
					return []byte(JWT_SECRET), nil
				})
			if err != nil {
				log_with_ctx(c, "warn", "Token parse error: %v", err)
				c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("invalid_token")})
				c.Abort()
				return
			}
			user_id, err := strconv.Atoi(usr_id)
			if err != nil {
				log_with_ctx(c, "warn", "Invalid user_id format: %v", err)
				c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("invalid_token")})
				c.Abort()
				return
			}
			if !token.Valid || claims.UserID != user_id {
				log_with_ctx(c, "warn", "Token is invalid or user_id mismatch")
				c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("invalid_token")})
				c.Abort()
				return
			}
			if !user_database.check_user_is_exist(user_id) {
				log_with_ctx(c, "warn", "User does not exist")
				c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("user_not_exists")})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// Index page
func (api *WebServerApi) Index(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to Rich Chat!")
}

// Get verification token for sensitive operations
func (api *WebServerApi) GetVerifyToken(c *gin.Context) {
	token, err := generateVerifyToken()
	if err != nil {
		log_with_ctx(c, "error", "Error generating verification token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": lp.G("internal_server_error"),
		})
		return
	}

	err = user_database.storeVerifyToken(token)
	if err != nil {
		log_with_ctx(c, "error", "Error storing verification token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": lp.G("internal_server_error"),
		})
		return
	}

	log_with_ctx(c, "info", "Verification token generated")
	c.JSON(http.StatusOK, gin.H{
		"verify_token": token,
	})
}

// Login user
func (api *WebServerApi) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	verifyToken := c.PostForm("verify_token")

	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("username_and_password_required"),
		})
		log_with_ctx(c, "warn", "Username and password are required")
		return
	}

	// Verify the verification token
	if verifyToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("verification_token_required"),
		})
		log_with_ctx(c, "warn", "Verification token is required")
		return
	}

	if !user_database.verifyAndConsumeToken(verifyToken) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": lp.G("invalid_or_expired_verification_token"),
		})
		log_with_ctx(c, "warn", "Invalid or expired verification token")
		return
	}

	if !user_database.check_username_is_exist(username) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": lp.G("username_or_password_is_invalid"),
		})
		log_with_ctx(c, "warn", "Invalid username or password for: %s", username)
		return
	}

	// Check if account is locked
	if user_database.isAccountLocked(username) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"message": lp.G("account_locked_try_later"),
		})
		log_with_ctx(c, "warn", "Login attempt on locked account: %s", username)
		return
	}

	user_id, pass := user_database.select_user_is_exist(username, password)
	if !pass {
		// Track failed login attempt
		user_database.trackLoginAttempt(username, false)

		remaining := user_database.getRemainingAttempts(username)
		c.JSON(http.StatusUnauthorized, gin.H{
			"message":            lp.G("username_or_password_is_invalid"),
			"remaining_attempts": remaining,
		})
		log_with_ctx(c, "warn", "Invalid username or password for: %s", username)
		return
	}

	// Track successful login
	user_database.trackLoginAttempt(username, true)

	jwt_token, err := generateToken(user_id, JWT_EXPIRE_TIME)
	if err != nil {
		log_with_ctx(c, "error", "Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": lp.G("internal_server_error"),
		})
		return
	}

	log_with_ctx(c, "info", "User id %d login successful", user_id)
	c.JSON(http.StatusOK, gin.H{
		"user_token": jwt_token,
		"user_id":    user_id,
	})
}

// Register user
func (api *WebServerApi) Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	verifyToken := c.PostForm("verify_token")

	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("username_and_password_required"),
		})
		log_with_ctx(c, "warn", "Username and password are required")
		return
	}

	// Verify the verification token
	if verifyToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("verification_token_required"),
		})
		log_with_ctx(c, "warn", "Verification token is required")
		return
	}

	if !user_database.verifyAndConsumeToken(verifyToken) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": lp.G("invalid_or_expired_verification_token"),
		})
		log_with_ctx(c, "warn", "Invalid or expired verification token")
		return
	}

	// Validate username length
	if len(username) > ALLOW_MAX_LENGTH_OF_USERNAME {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("length_of_username_must_less_than") + strconv.Itoa(ALLOW_MAX_LENGTH_OF_USERNAME) + lp.G("characters"),
		})
		log_with_ctx(c, "warn", "Username must be less than %d characters", ALLOW_MAX_LENGTH_OF_USERNAME)
		return
	}

	user_id, err := user_database.insert_user(username, password)
	if err != nil {
		// Check if it's a duplicate username error
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{
				"message": lp.G("username_already_exists"),
			})
			return
		}
		log_with_ctx(c, "error", "Error registering user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": lp.G("internal_server_error"),
		})
		return
	}

	jwt_token, err := generateToken(user_id, JWT_EXPIRE_TIME)
	if err != nil {
		log_with_ctx(c, "error", "Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": lp.G("internal_server_error"),
		})
		return
	}

	log_with_ctx(c, "info", "User id %d registration successful", user_id)
	c.JSON(http.StatusOK, gin.H{
		"user_token": jwt_token,
		"user_id":    user_id,
		"message":    lp.G("registration_successful"),
	})
}

// delete user
func (api *WebServerApi) DeleteUser(c *gin.Context) {
	user_id, err := strconv.Atoi(c.PostForm("user_id"))
	if err != nil {
		log_with_ctx(c, "warn", "Invalid user_id format: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": lp.G("invalid_user_id_format")})
		return
	}
	user_password := c.PostForm("user_password")
	verifyToken := c.PostForm("verify_token")

	if user_password == "" {
		log_with_ctx(c, "warn", "User password is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": lp.G("user_password_required")})
		return
	}

	// Verify the verification token
	if verifyToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": lp.G("verification_token_required"),
		})
		log_with_ctx(c, "warn", "Verification token is required")
		return
	}

	if !user_database.verifyAndConsumeToken(verifyToken) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": lp.G("invalid_or_expired_verification_token"),
		})
		log_with_ctx(c, "warn", "Invalid or expired verification token")
		return
	}

	// Check if account is locked
	identifier := fmt.Sprintf("%d", user_id)
	if user_database.isAccountLocked(identifier) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"message": lp.G("account_locked_try_later"),
		})
		log_with_ctx(c, "warn", "Delete attempt on locked account: %s", identifier)
		return
	}

	if !user_database.select_user_id_is_exist(user_id, user_password) {
		// Track failed attempt
		user_database.trackLoginAttempt(identifier, false)

		log_with_ctx(c, "warn", "Invalid user_id or user_password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("invalid_user_id_or_password")})
		return
	}

	// Track successful verification
	user_database.trackLoginAttempt(identifier, true)

	err = user_database.delete_user(user_id)
	if err != nil {
		log_with_ctx(c, "warn", "Can't delete the account: %v", err)
	}
	log_with_ctx(c, "info", "User id %d delete successful", user_id)
	c.JSON(http.StatusOK, gin.H{
		"message": lp.G("user_deleted_successfully"),
	})
}
