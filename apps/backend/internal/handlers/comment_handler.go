package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
	"github.com/sp3640/opspilot/backend/internal/authorization"
	"github.com/sp3640/opspilot/backend/internal/rbac"
	"github.com/sp3640/opspilot/backend/internal/response"
	"github.com/sp3640/opspilot/backend/internal/services"
)

type CommentHandler struct {
	service *services.CommentService
}

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

func NewCommentHandler(service *services.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

func (h *CommentHandler) Create(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) || !authorization.RequirePermission(c, rbac.PermissionIncidentComment) {
		return
	}

	incidentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid incident id")
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	comment, err := h.service.CreateComment(c.Request.Context(), req.Content, uint(incidentID), userID, organizationID)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidCommentContent:
			response.Error(c, http.StatusBadRequest, err.Error())
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		case apperrors.ErrIncidentNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.Created(c, "Comment created successfully", comment)
}

func (h *CommentHandler) List(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	incidentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid incident id")
		return
	}

	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	req, ok := parsePagination(c, "created_at", "updated_at")
	if !ok {
		return
	}
	result, err := h.service.ListCommentsByIncidentID(uint(incidentID), organizationID, req)
	if err != nil {
		switch err {
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		case apperrors.ErrIncidentNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Comments fetched successfully", result)
}

func (h *CommentHandler) Update(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid comment id")
		return
	}

	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	comment, err := h.service.UpdateComment(c.Request.Context(), uint(commentID), userID, organizationID, req.Content)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidCommentContent:
			response.Error(c, http.StatusBadRequest, err.Error())
		case apperrors.ErrCommentNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case apperrors.ErrCommentForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Comment updated successfully", comment)
}

func (h *CommentHandler) Delete(c *gin.Context) {
	if !authorization.RequireOrganizationMember(c) {
		return
	}

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid comment id")
		return
	}

	userID := c.MustGet("userID").(uint)
	organizationID, ok := parseOrganizationIDFromContext(c)
	if !ok {
		return
	}

	err = h.service.DeleteComment(c.Request.Context(), uint(commentID), userID, organizationID)
	if err != nil {
		switch err {
		case apperrors.ErrCommentNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case apperrors.ErrCommentForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Comment deleted successfully", nil)
}
