package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/dto"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/services"

	"github.com/sp3640/opspilot/backend/internal/response"
)

type OrganizationHandler struct {
	service *services.OrganizationService
}

func NewOrganizationHandler(service *services.OrganizationService) *OrganizationHandler {
	return &OrganizationHandler{service: service}
}

func (h *OrganizationHandler) Create(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	userID := c.MustGet("userID").(uint)

	var req dto.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	organization, err := h.service.Create(userID, req)
	if err != nil {
		switch err {
		case apperrors.ErrOrganizationAlreadyExists:
			response.Conflict(c, err.Error())
		case apperrors.ErrInvalidOrganizationName, apperrors.ErrInvalidOrganizationDescription, apperrors.ErrInvalidOrganizationSlug:
			response.BadRequest(c, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.Created(c, "Organization created successfully", organization)
}

func (h *OrganizationHandler) List(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	req, ok := parsePagination(c, "name", "created_at", "updated_at")
	if !ok {
		return
	}

	organization, err := h.service.GetByID(organizationID, organizationID)
	if err != nil {
		switch err {
		case apperrors.ErrOrganizationNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case apperrors.ErrOrganizationForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	organizations := &dto.OrganizationListResponse{
		Items:      []dto.OrganizationResponse{*organization},
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      1,
		TotalPages: 1,
	}

	response.OK(c, "Organizations fetched successfully", organizations)
}

func (h *OrganizationHandler) GetByID(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	callerOrganizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	organizationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid organization id")
		return
	}

	organization, err := h.service.GetByID(organizationID, callerOrganizationID)
	if err != nil {
		switch err {
		case apperrors.ErrOrganizationNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case apperrors.ErrOrganizationForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Organization fetched successfully", organization)
}

func (h *OrganizationHandler) Update(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	callerOrganizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	organizationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid organization id")
		return
	}

	var req dto.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	organization, err := h.service.Update(organizationID, callerOrganizationID, req)
	if err != nil {
		switch err {
		case apperrors.ErrOrganizationNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case apperrors.ErrOrganizationForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		case apperrors.ErrOrganizationAlreadyExists:
			response.Conflict(c, err.Error())
		case apperrors.ErrInvalidOrganizationName, apperrors.ErrInvalidOrganizationDescription, apperrors.ErrInvalidOrganizationSlug:
			response.BadRequest(c, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Organization updated successfully", organization)
}

func (h *OrganizationHandler) Delete(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionOrganizationManage) {
		return
	}

	callerOrganizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	organizationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid organization id")
		return
	}

	if err := h.service.Delete(organizationID, callerOrganizationID); err != nil {
		switch err {
		case apperrors.ErrOrganizationNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case apperrors.ErrOrganizationForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Organization deleted successfully", nil)
}
