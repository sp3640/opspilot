package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type AuthHandler struct {
	userService *services.UserService
}

// RegisterRequest represents the request body for user registration.
// OrganizationName is optional: it only names the new workspace created for
// an uninvited registrant. It is ignored when the email matches a pending
// invitation, since that flow joins the invitation's existing organization.
type RegisterRequest struct {
	Name             string `json:"name" binding:"required"`
	Email            string `json:"email" binding:"required,email"`
	Password         string `json:"password" binding:"required,min=8"`
	OrganizationName string `json:"organizationName" binding:"omitempty,max=100"`
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

	if err := h.userService.Register(req.Name, req.Email, req.Password, req.OrganizationName); err != nil {

		switch err {

		case apperrors.ErrEmailAlreadyExists:
			response.Conflict(c, err.Error())

		default:
			response.InternalServerError(c, err)
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
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	if err != nil {

		switch err {

		case apperrors.ErrInvalidCredentials:
			response.Unauthorized(c, err.Error())

		default:
			response.InternalServerError(c, err)
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

// Reissue mints a fresh access token for the already-authenticated caller,
// reflecting their CURRENT database role and organization. It requires a
// currently-valid access token (via AuthMiddleware) rather than a separate
// refresh token, so a session started before a role/organization change —
// most commonly, accepting an invitation — can be brought up to date
// without a full re-login and without introducing a second token type.
func (h *AuthHandler) Reissue(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	token, err := h.userService.ReissueToken(userID)
	if err != nil {
		switch err {
		case apperrors.ErrUserNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(
		c,
		"Session refreshed successfully",
		gin.H{
			"access_token": token,
			"token_type":   "Bearer",
		},
	)
}
