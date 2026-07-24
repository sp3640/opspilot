package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type AuthHandler struct {
	userService *services.UserService
}

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register handles user registration.
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userService.Register(req.Name, req.Email, req.Password); err != nil {

		switch err {

		case apperrors.ErrEmailAlreadyExists:
			response.Conflict(c, err.Error())

		default:
			response.InternalServerError(c)
		}

		return
	}

	response.Created(
		c,
		"User registered successfully",
		nil,
	)
}

// Login handles user authentication.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	token, err := h.userService.Login(
		req.Email,
		req.Password,
	)

	if err != nil {

		switch err {

		case apperrors.ErrInvalidCredentials:
			response.Unauthorized(c, err.Error())

		default:
			response.InternalServerError(c)
		}

		return
	}

	response.OK(
		c,
		"Login successful",
		gin.H{
			"access_token": token,
			"token_type":   "Bearer",
		},
	)
}