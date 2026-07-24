package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (h *UserHandler) Me(c *gin.Context) {

	userID := c.MustGet("userID").(uint)

	user, err := h.service.GetCurrentUser(userID)
	if err != nil {

		switch err {

		case apperrors.ErrUserNotFound:
			response.Error(c, http.StatusNotFound, err.Error())

		default:
			response.InternalServerError(c)
		}

		return
	}

	response.OK(c, "User fetched successfully", gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}