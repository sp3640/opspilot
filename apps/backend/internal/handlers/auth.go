package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type AuthHandler struct {
	userService  *services.UserService
	verification *services.EmailVerificationService
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

// VerifyEmailRequest carries the raw token from a verification link. It is
// sent in the request body, not the URL, so it never lands in access logs.
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required,max=256"`
}

// ResendVerificationRequest asks for a new verification email.
type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Stable machine-readable codes the frontend branches on.
const (
	codeEmailNotVerified         = "email_not_verified"
	codeEmailDeliveryUnavailable = "email_delivery_unavailable"
	codeVerificationInvalid      = "verification_token_invalid"
	codeVerificationExpired      = "verification_token_expired"
)

const emailDeliveryUnavailableMessage = "We couldn't send the verification email right now. Please try again in a few minutes."

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// WithEmailVerification enables the verify and resend endpoints.
func (h *AuthHandler) WithEmailVerification(verification *services.EmailVerificationService) *AuthHandler {
	h.verification = verification
	return h
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

		case apperrors.ErrEmailDeliveryUnavailable:
			response.ErrorWithCode(c, http.StatusServiceUnavailable, codeEmailDeliveryUnavailable, emailDeliveryUnavailableMessage)

		default:
			response.InternalServerError(c, err)
		}

		return
	}

	// The same response whether or not the address was already registered
	// (see UserService.Register), so this endpoint cannot be used to
	// discover accounts. The account exists but is unverified; there is
	// deliberately no token in the response.
	response.Created(
		c,
		"Check your email for a link to verify your account.",
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

		case apperrors.ErrEmailNotVerified:
			response.ErrorWithCode(c, http.StatusForbidden, codeEmailNotVerified, "Please verify your email before signing in.")

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
		case apperrors.ErrEmailNotVerified:
			response.ErrorWithCode(c, http.StatusForbidden, codeEmailNotVerified, "Please verify your email before signing in.")
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

// VerifyEmail redeems a verification token. It is a POST (not a GET link
// hit directly) on purpose: the emailed link opens a frontend page that
// submits the token, so mail scanners that pre-fetch links cannot consume
// it, and the token stays out of URLs the API logs.
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithCode(c, http.StatusBadRequest, codeVerificationInvalid, "This verification link is invalid.")
		return
	}
	if h.verification == nil {
		response.ErrorWithCode(c, http.StatusServiceUnavailable, codeEmailDeliveryUnavailable, emailDeliveryUnavailableMessage)
		return
	}

	outcome, err := h.verification.Verify(req.Token, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		switch err {
		case apperrors.ErrVerificationTokenInvalid:
			response.ErrorWithCode(c, http.StatusBadRequest, codeVerificationInvalid, "This verification link is invalid or has been replaced by a newer one.")
		case apperrors.ErrVerificationTokenExpired:
			response.ErrorWithCode(c, http.StatusGone, codeVerificationExpired, "This verification link has expired. Request a new one to continue.")
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	message := "Email verified successfully."
	if outcome == services.VerificationAlreadyVerified {
		message = "This email address is already verified."
	}
	response.OK(c, message, gin.H{"status": string(outcome)})
}

// ResendVerification sends a new verification link. Every outcome that
// depends on the address (unknown, already verified, inside the cooldown)
// gets the same success response, so it cannot be used to find out who has
// an account.
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if h.verification == nil {
		response.ErrorWithCode(c, http.StatusServiceUnavailable, codeEmailDeliveryUnavailable, emailDeliveryUnavailableMessage)
		return
	}

	if err := h.verification.Resend(c.Request.Context(), req.Email, c.ClientIP(), c.Request.UserAgent()); err != nil {
		switch err {
		case apperrors.ErrEmailDeliveryUnavailable:
			response.ErrorWithCode(c, http.StatusServiceUnavailable, codeEmailDeliveryUnavailable, emailDeliveryUnavailableMessage)
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "If that account is awaiting verification, a new verification link is on its way.", nil)
}
