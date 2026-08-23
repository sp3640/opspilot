package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type ApplicationHealthHandler struct {
	service *services.ApplicationHealthService
}

func NewApplicationHealthHandler(service *services.ApplicationHealthService) *ApplicationHealthHandler {
	return &ApplicationHealthHandler{service: service}
}

// GetByApplication returns the application's unified health score. Read-only,
// so it is gated the same as every other application-scoped read endpoint
// (org membership only) rather than requiring PermissionApplicationManage.
func (h *ApplicationHealthHandler) GetByApplication(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application id")
		return
	}

	result, err := h.service.GetApplicationHealth(c.Request.Context(), applicationID, organizationID)
	if err != nil {
		switch err {
		case apperrors.ErrApplicationForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		case apperrors.ErrApplicationNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Application health fetched successfully", result)
}
