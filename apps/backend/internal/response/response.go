package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
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
	c.JSON(status, APIResponse{
		Success: false,
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

func InternalServerError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "Internal server error")
}