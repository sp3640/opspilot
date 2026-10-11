package response

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/logger"
)

type APIResponse struct {
	Success bool `json:"success"`
	// Code is an optional stable, machine-readable reason (for example
	// "email_not_verified") that clients branch on instead of parsing
	// Message. Omitted from every response that does not set one.
	Code    string      `json:"code,omitempty"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(
	c *gin.Context,
	status int,
	message string,
	data interface{},
) {
	c.JSON(status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(
	c *gin.Context,
	status int,
	message string,
) {
	if status >= http.StatusInternalServerError {
		logger.Error(
			c.Request.Context(),
			"HTTP error response",
			slog.Int("status", status),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)
	}

	c.JSON(status, APIResponse{
		Success: false,
		Message: message,
	})
}

// ErrorWithCode is Error plus a machine-readable Code.
func ErrorWithCode(c *gin.Context, status int, code, message string) {
	if status >= http.StatusInternalServerError {
		logger.Error(
			c.Request.Context(),
			"HTTP error response",
			slog.Int("status", status),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("code", code),
		)
	}

	c.JSON(status, APIResponse{
		Success: false,
		Code:    code,
		Message: message,
	})
}

func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message)
}

func InternalServerError(c *gin.Context, causes ...error) {
	attributes := []slog.Attr{
		slog.Int("status", http.StatusInternalServerError),
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.Path),
	}
	if len(causes) > 0 && causes[0] != nil {
		attributes = append(attributes, slog.Any("error", causes[0]))
	}
	logger.Error(c.Request.Context(), "internal server error", attributes...)

	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Message: "Internal server error",
	})
}
