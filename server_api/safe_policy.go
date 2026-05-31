package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

// check client if it is safe - Refactored to use services
func safe_check() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP address
		clientIP := c.ClientIP()

		// Check if IP is blocked using rate limit repository
		if dbService != nil {
			rateLimitRepo := dbService.GetRateLimitRepository()
			isBlocked, _ := rateLimitRepo.CheckIPBlocked(clientIP)
			if isBlocked {
				log.WithFields(log.Fields{
					"ip": clientIP,
				}).Warning("Blocked IP attempted access")
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
		}

		// Track IP visit count in Redis
		if dbService != nil {
			rateLimitRepo := dbService.GetRateLimitRepository()
			visitCount, err := rateLimitRepo.TrackIPVisit(clientIP)
			if err == nil && visitCount > int64(IP_LIMIT_VISIT_TIMES) {
				// Block IP
				reason := fmt.Sprintf("Exceeded rate limit: %d visits in %v (limit: %d)",
					visitCount, IP_LIMIT_TIME, IP_LIMIT_VISIT_TIMES)
				_ = rateLimitRepo.BlockIP(clientIP, reason, IP_LIMIT_LOCKOUT_DURATION)
				
				log.WithFields(log.Fields{
					"ip":     clientIP,
					"visits": visitCount,
				}).Error("IP blocked due to rate limiting")
				
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
		}

		// Check User-Agent
		if !strings.HasPrefix(c.GetHeader("User-Agent"), ALLOW_USER_AGENT) {
			log.WithFields(log.Fields{
				"user_agent": c.GetHeader("User-Agent"),
			}).Warning("User-Agent is not allowed")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		
		client_version := c.GetHeader("User-Agent")[len(ALLOW_USER_AGENT)+1:]
		if client_version != VERSION {
			log.WithFields(log.Fields{
				"expected": VERSION,
				"actual":   client_version,
			}).Warning("Version mismatch")
			c.AbortWithStatus(http.StatusExpectationFailed)
			return
		}

		// Check if user token is valid
		usr_token := c.GetHeader("user_token")
		if usr_token != "" {
			usr_id := c.GetHeader("user_id")
			if usr_id == "" {
				log.Warning("user_id is empty")
				c.AbortWithStatus(http.StatusExpectationFailed)
				return
			}
			
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(usr_token, claims,
				func(token *jwt.Token) (interface{}, error) {
					// Validate the signing method to prevent "alg: none" attack
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return []byte(JWT_SECRET), nil
				})
			
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Warning("Token parse error")
				c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("invalid_token")})
				c.Abort()
				return
			}
			
			user_id, err := strconv.Atoi(usr_id)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Warning("Invalid user_id format")
				c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("invalid_token")})
				c.Abort()
				return
			}
			
			if !token.Valid || claims.UserID != user_id {
				log.Warning("Token is invalid or user_id mismatch")
				c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("invalid_token")})
				c.Abort()
				return
			}
			
			// Check if user exists using UserService
			if services != nil {
				exists, _ := services.UserService.CheckUserExists(user_id)
				if !exists {
					log.WithFields(log.Fields{
						"user_id": user_id,
					}).Warning("User does not exist")
					c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("user_not_exists")})
					c.Abort()
					return
				}
			}
		}
		
		c.Next()
	}
}

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

// Generate a random verification token
func generateVerifyToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
