package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouterForMiddleware() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestRequestIDMiddleware(t *testing.T) {
	router := setupRouterForMiddleware()
	router.Use(request_id())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	t.Run("generates new request ID when not provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
	})

	t.Run("uses client-provided request ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "custom-id-123")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "custom-id-123", w.Header().Get("X-Request-ID"))
	})

	t.Run("sets request ID in context", func(t *testing.T) {
		var capturedID string
		router := setupRouterForMiddleware()
		router.Use(request_id())
		router.GET("/test", func(c *gin.Context) {
			id, exists := c.Get("request_id")
			if exists {
				capturedID = id.(string)
			}
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.NotEmpty(t, capturedID)
	})
}

func TestForceHTTPSMiddleware(t *testing.T) {
	t.Run("does not redirect when HTTPS_FORCE is false", func(t *testing.T) {
		router := setupRouterForMiddleware()
		router.Use(force_https())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCheckClientUserAgent(t *testing.T) {
	t.Run("allows valid user agent", func(t *testing.T) {
		router := setupRouterForMiddleware()
		router.Use(check_client_user_agent())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "rich_chat 1.0.0")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("rejects invalid user agent", func(t *testing.T) {
		router := setupRouterForMiddleware()
		router.Use(check_client_user_agent())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "invalid-agent")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("rejects version mismatch", func(t *testing.T) {
		router := setupRouterForMiddleware()
		router.Use(check_client_user_agent())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "rich_chat 0.9.0")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusExpectationFailed, w.Code)
	})
}

func TestLanguageMiddleware(t *testing.T) {
	t.Run("uses default language when not specified", func(t *testing.T) {
		router := setupRouterForMiddleware()
		router.Use(languageMiddleware())
		router.GET("/test", func(c *gin.Context) {
			lang, exists := c.Get("language")
			if exists {
				c.String(http.StatusOK, lang.(string))
			} else {
				c.String(http.StatusOK, "no-language")
			}
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, DEFAULT_LANGUAGE, w.Body.String())
	})

	t.Run("uses language from query parameter", func(t *testing.T) {
		router := setupRouterForMiddleware()
		router.Use(languageMiddleware())
		router.GET("/test", func(c *gin.Context) {
			lang, exists := c.Get("language")
			if exists {
				c.String(http.StatusOK, lang.(string))
			}
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test?language=en", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "en", w.Body.String())
	})

	t.Run("uses language from header", func(t *testing.T) {
		router := setupRouterForMiddleware()
		router.Use(languageMiddleware())
		router.GET("/test", func(c *gin.Context) {
			lang, exists := c.Get("language")
			if exists {
				c.String(http.StatusOK, lang.(string))
			}
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Language", "en")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "en", w.Body.String())
	})
}

func TestCheckIPInBlock(t *testing.T) {
	t.Run("continues when DBService is nil", func(t *testing.T) {
		router := setupRouterForMiddleware()
		deps := &MiddlewareDependencies{DBService: nil}
		router.Use(check_ip_in_block(deps))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTrackIPVisit(t *testing.T) {
	t.Run("continues when DBService is nil", func(t *testing.T) {
		router := setupRouterForMiddleware()
		deps := &MiddlewareDependencies{DBService: nil}
		router.Use(track_ip_visit(deps))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCheckUserIsExists(t *testing.T) {
	t.Run("rejects invalid user_id format", func(t *testing.T) {
		router := setupRouterForMiddleware()
		deps := &MiddlewareDependencies{Services: nil}
		router.Use(check_user_is_exists(deps))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("user_id", "invalid")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "Invalid user ID format", response["error"])
	})

	t.Run("continues when Services is nil", func(t *testing.T) {
		router := setupRouterForMiddleware()
		deps := &MiddlewareDependencies{Services: nil}
		router.Use(check_user_is_exists(deps))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("user_id", "123")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCheckClientToken(t *testing.T) {
	t.Run("rejects empty user_token", func(t *testing.T) {
		router := setupRouterForMiddleware()
		router.Use(check_client_token())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("user_id", "123")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("rejects empty user_id", func(t *testing.T) {
		router := setupRouterForMiddleware()
		router.Use(check_client_token())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("user_token", "some-token")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		router := setupRouterForMiddleware()
		router.Use(check_client_token())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("user_token", "invalid-token")
		req.Header.Set("user_id", "123")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
