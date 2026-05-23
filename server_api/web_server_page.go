package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// Generate a JWT token for a given user ID
func generateToken(user_id int, valid_time time.Duration) (string, error) {
	expirationTime := time.Now().Add(valid_time)

	claims := &Claims{
		UserID: user_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "rich_chat",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWT_SECRET))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

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
			log_with_ctx(c, "warn", "Version mismatch:", VERSION, "!=", client_version)
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
				log_with_ctx(c, "warn", "Token parse error:", err.Error())
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
			user_id, err := strconv.Atoi(usr_id)
			if err != nil {
				log_with_ctx(c, "warn", "Invalid user_id format:", err.Error())
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
			if !token.Valid || claims.UserID != user_id {
				log_with_ctx(c, "warn", "Token is invalid or user_id mismatch")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
			if !user_database.check_user_is_exist(user_id) {
				log_with_ctx(c, "warn", "User does not exist")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "The user is not exist"})
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

func (api *WebServerApi) Login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Username and password are required",
		})
		log_with_ctx(c, "warn", "Username and password are required")
		return
	}

	user_id, pass := user_database.select_user_is_exist(username, password)
	if !pass {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid username or password",
		})
		log_with_ctx(c, "warn", "Invalid username or password")
		return
	}

	jwt_token, err := generateToken(user_id, JWT_EXPIRE_TIME)
	if err != nil {
		log_with_ctx(c, "error", "Error generating token:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
		return
	}

	log_with_ctx(c, "info", "User id", user_id, "login successful")
	c.JSON(http.StatusOK, gin.H{
		"user_token": jwt_token,
		"user_id":    user_id,
	})
}

func (api *WebServerApi) Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Username and password are required",
		})
		log_with_ctx(c, "warn", "Username and password are required")
		return
	}

	// Validate username length
	if len(username) > ALLOW_MAX_LENGTH_OF_USERNAME {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Username must less than " + strconv.Itoa(ALLOW_MAX_LENGTH_OF_USERNAME) + " characters",
		})
		log_with_ctx(c, "warn", "Username must be less than", ALLOW_MAX_LENGTH_OF_USERNAME, "characters")
		return
	}

	user_id, err := user_database.insert_user(username, password)
	if err != nil {
		// Check if it's a duplicate username error
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{
				"message": "Username already exists",
			})
			return
		}
		log_with_ctx(c, "error", "Error registering user:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
		return
	}

	jwt_token, err := generateToken(user_id, JWT_EXPIRE_TIME)
	if err != nil {
		log_with_ctx(c, "error", "Error generating token:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
		return
	}

	log_with_ctx(c, "info", "User id", user_id, "registration successful")
	c.JSON(http.StatusOK, gin.H{
		"user_token": jwt_token,
		"user_id":    user_id,
		"message":    "Registration successful",
	})
}

func (api *WebServerApi) DeleteUser(c *gin.Context) {
	user_id, err := strconv.Atoi(c.PostForm("user_id"))
	if err != nil {
		log_with_ctx(c, "warn", "Invalid user_id format:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}
	user_password := c.PostForm("user_password")
	if user_password == "" {
		log_with_ctx(c, "warn", "User password is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "User_password is required"})
		return
	}
	if !user_database.select_user_id_is_exist(user_id, user_password) {
		log_with_ctx(c, "warn", "Invalid user_id or user_password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user_id or user_password"})
		return
	}
	err = user_database.delete_user(user_id)
	if err != nil {
		log_with_ctx(c, "warn", "Can't delete the account: ", err.Error())
	}
	log_with_ctx(c, "info", "User id", user_id, "delete successful")
	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}
