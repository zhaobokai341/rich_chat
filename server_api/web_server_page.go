package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

// check client if it is safe
func safe_check() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.GetHeader("User-Agent"), "rich_chat ") {
			log.Warning(
				fmt.Sprintf(
					"From %s, User-Agent is %s. Forbidden",
					c.Request.RemoteAddr,
					c.GetHeader("User-Agent"),
				),
			)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		client_version := c.GetHeader("User-Agent")[len("rich_chat "):]
		if client_version != VERSION {
			log.Warning(
				fmt.Sprintf(
					"From %s, User-Agent is %s. version mismatch",
					c.Request.RemoteAddr,
					c.GetHeader("User-Agent"),
				),
			)
			c.AbortWithStatus(http.StatusExpectationFailed)
			return
		}
		usr_token := c.GetHeader("user_token")
		if usr_token != "" {
			usr_id := c.GetHeader("user_id")
			if usr_id == "" {
				log.Warning(
					fmt.Sprintf(
						"From %s, User-Agent is %s. user_id is empty",
						c.Request.RemoteAddr,
					),
				)
				c.AbortWithStatus(http.StatusExpectationFailed)
				return
			}
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(usr_token, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(JWT_SECRET), nil
			})
			if err != nil {
				log.Warning(
					fmt.Sprintf(
						"From %s, User-Agent is %s. error is %s",
						c.Request.RemoteAddr,
						c.GetHeader("User-Agent"),
						err.Error(),
					),
				)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
			user_id, err := strconv.Atoi(usr_id)
			if err != nil {
				log.Warning(
					fmt.Sprintf(
						"From %s, User-Agent is %s. error is %s",
						c.Request.RemoteAddr,
						c.GetHeader("User-Agent"),
						err.Error(),
					),
				)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
			if !token.Valid || claims.UserID != user_id {
				log.Warning(
					fmt.Sprintf(
						"From %s, User-Agent is %s. the token is invalid",
						c.Request.RemoteAddr,
						c.GetHeader("User-Agent"),
					),
				)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
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
		return
	}

	user_id, pass := user_database.select_user_is_exist(username, password)
	if !pass {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid username or password",
		})
		return
	}

	jwt_token, err := generateToken(user_id)
	if err != nil {
		log.Error("Error generating token: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
		return
	}

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
		return
	}

	// Validate username length
	if len(username) < 1 || len(username) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Username must be between 3 and 50 characters",
		})
		return
	}

	user_id, err := user_database.insert_user(username, password)
	if err != nil {
		// Check if it's a duplicate username error
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			c.JSON(http.StatusConflict, gin.H{
				"message": "Username already exists",
			})
			return
		}
		log.Error("Error registering user: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
		return
	}

	jwt_token, err := generateToken(user_id)
	if err != nil {
		log.Error("Error generating token: ", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_token": jwt_token,
		"user_id":    user_id,
		"message":    "Registration successful",
	})
}
