package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/models"
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
			response.InternalServerError(c, err)
		}

		return
	}

	response.OK(c, "User fetched successfully", gin.H{
		"id":               user.ID,
		"name":             user.Name,
		"email":            user.Email,
		"role":             resolvedRole(user.Role),
		"organizationId":   resolvedOrganizationID(user.OrganizationID),
		"organizationName": resolvedOrganizationName(user.Organization),
		"organizationSlug": resolvedOrganizationSlug(user.Organization),
	})
}

func resolvedRole(role string) string {
	if strings.TrimSpace(role) == "" {
		return models.RoleViewer
	}

	return role
}

func resolvedOrganizationID(organizationID *uuid.UUID) string {
	if organizationID == nil {
		return ""
	}

	return organizationID.String()
}

func resolvedOrganizationName(organization *models.Organization) string {
	if organization == nil {
		return ""
	}

	return organization.Name
}

func resolvedOrganizationSlug(organization *models.Organization) string {
	if organization == nil {
		return ""
	}

	return organization.Slug
}
