package main

import (
	"fmt"
	"net/http"
	"rich_chat/lang_pack_load"
	"rich_chat/server_api/database"
	"rich_chat/server_api/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

// MiddlewareDependencies holds dependencies for middleware functions
type MiddlewareDependencies struct {
	DBService *database.DatabaseService
	Services  *service.Services
}

// getLanguagePackFromContext retrieves the language pack from gin context
func getLanguagePackFromContext(c *gin.Context) *lang_pack_load.LanguagePack {
	lang, exists := c.Get("language")
	if !exists {
		lang = DEFAULT_LANGUAGE
	}

	lp := lang_pack_load.NewLanguagePack("lang.json", lang.(string))
	lp.Load()
	return lp
}

// languageMiddleware extracts language from request and stores it in context
func languageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Query("language")
		if lang == "" {
			lang = c.PostForm("language")
		}
		if lang == "" {
			lang = c.GetHeader("X-Language")
		}
		if lang == "" {
			lang = DEFAULT_LANGUAGE
		}

		c.Set("language", lang)
		c.Next()
	}
}

// check_ip_in_block creates middleware that checks if IP is blocked
func check_ip_in_block(deps *MiddlewareDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if deps.DBService == nil {
			c.Next()
			return
		}
		rateLimitRepo := deps.DBService.GetRateLimitRepository()
		isBlocked, _ := rateLimitRepo.CheckIPBlocked(clientIP)
		if isBlocked {
			log.WithFields(log.Fields{
				"ip": clientIP,
			}).Warning("Blocked IP attempted access")
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
	}
}

// track_ip_visit creates middleware that tracks IP visit count
func track_ip_visit(deps *MiddlewareDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if deps.DBService == nil {
			c.Next()
			return
		}

		rateLimitRepo := deps.DBService.GetRateLimitRepository()
		visitCount, err := rateLimitRepo.TrackIPVisit(clientIP)
		if err == nil && visitCount > int64(IP_LIMIT_VISIT_TIMES) {
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
}

// check_client_user_agent validates client user agent
func check_client_user_agent() gin.HandlerFunc {
	return func(c *gin.Context) {
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
	}
}

// check_verify_token validates verification token
func check_verify_token(deps *MiddlewareDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		verify_token := c.PostForm("verify_token")
		if verify_token == "" {
			verify_token = c.Query("verify_token")
		}
		if err := deps.Services.TokenService.ValidateAndConsumeToken(verify_token); err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

// check_client_token validates client JWT token
func check_client_token() gin.HandlerFunc {
	return func(c *gin.Context) {
		usr_token := c.GetHeader("user_token")
		if usr_token == "" {
			log.Warning("user_token is empty")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		usr_id := c.GetHeader("user_id")
		if usr_id == "" {
			log.Warning("user_id is empty")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(usr_token, claims,
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(JWT_SECRET), nil
			})

		if err != nil {
			lp := getLanguagePackFromContext(c)
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Warning("Token parse error")
			c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("invalid_token")})
			c.Abort()
			return
		}

		user_id, err := strconv.Atoi(usr_id)
		if err != nil {
			lp := getLanguagePackFromContext(c)
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Warning("Invalid user_id format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("invalid_token")})
			c.Abort()
			return
		}

		if !token.Valid || claims.UserID != user_id {
			lp := getLanguagePackFromContext(c)
			log.Warning("Token is invalid or user_id mismatch")
			c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("invalid_token")})
			c.Abort()
			return
		}
	}
}

// check_user_is_exists validates user exists
func check_user_is_exists(deps *MiddlewareDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		usr_id := c.GetHeader("user_id")
		user_id, _ := strconv.Atoi(usr_id)

		if deps.Services == nil {
			c.Next()
			return
		}
		exists, _ := deps.Services.UserService.CheckUserExists(user_id)
		if !exists {
			lp := getLanguagePackFromContext(c)
			log.WithFields(log.Fields{
				"user_id": user_id,
			}).Warning("User does not exist")
			c.JSON(http.StatusUnauthorized, gin.H{"error": lp.G("authentication_failed")})
			c.Abort()
			return
		}
	}
}

// force_https forces HTTPS if configured
func force_https() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HTTPS_FORCE {
			return
		}
		if c.Request.Header.Get("X-Forwarded-Proto") != "https" ||
			c.Request.TLS == nil {
			target := "https://" + c.Request.Host + c.Request.URL.Path
			if len(c.Request.URL.RawQuery) > 0 {
				target += "?" + c.Request.URL.RawQuery
			}
			c.Redirect(http.StatusPermanentRedirect, target)
			c.Abort()
			return
		}
	}
}
