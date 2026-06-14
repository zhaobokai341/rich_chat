package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"rich_chat/server_api/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "ErrUserNotFound",
			err:            service.ErrUserNotFound,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "authentication_failed",
		},
		{
			name:           "ErrInvalidInput",
			err:            service.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "username_and_password_required",
		},
		{
			name:           "ErrInvalidEmailFormat",
			err:            service.ErrInvalidEmailFormat,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid_input",
		},
		{
			name:           "ErrEmailExceedsMaxLength",
			err:            service.ErrEmailExceedsMaxLength,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid_input",
		},
		{
			name:           "ErrBioExceedsMaxLength",
			err:            service.ErrBioExceedsMaxLength,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid_input",
		},
		{
			name:           "ErrPasswordExceedsMaxLength",
			err:            service.ErrPasswordExceedsMaxLength,
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "invalid_input",
		},
		{
			name:           "ErrAccountLocked",
			err:            service.ErrAccountLocked,
			expectedStatus: http.StatusTooManyRequests,
			expectedMsg:    "account_locked_try_later",
		},
		{
			name:           "ErrInvalidPassword",
			err:            service.ErrInvalidPassword,
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "authentication_failed",
		},
		{
			name:           "generic error",
			err:            errors.New("unknown error"),
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "internal_server_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			c.Set("language", "en")

			handleServiceError(c, tt.err)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response["message"], tt.expectedMsg)
		})
	}
}
