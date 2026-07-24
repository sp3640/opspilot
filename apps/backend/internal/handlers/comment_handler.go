package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sp3640/opspilot/backend/internal/apperrors"
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

	comment, err := h.service.CreateComment(req.Content, uint(incidentID), userID)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidCommentContent:
			response.Error(c, http.StatusBadRequest, err.Error())
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		case apperrors.ErrIncidentNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		default:
			response.InternalServerError(c)
		}
		return
	}

	response.Created(c, "Comment created successfully", comment)
}

func (h *CommentHandler) List(c *gin.Context) {
	incidentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid incident id")
		return
	}

	userID := c.MustGet("userID").(uint)

	req, ok := parsePagination(c, "created_at", "updated_at")
	if !ok {
		return
	}
	result, err := h.service.ListCommentsByIncidentID(uint(incidentID), userID, req)
	if err != nil {
		switch err {
		case apperrors.ErrProjectForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		case apperrors.ErrIncidentNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		default:
			response.InternalServerError(c)
		}
		return
	}

	response.OK(c, "Comments fetched successfully", result)
}

func (h *CommentHandler) Update(c *gin.Context) {
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

	comment, err := h.service.UpdateComment(uint(commentID), userID, req.Content)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidCommentContent:
			response.Error(c, http.StatusBadRequest, err.Error())
		case apperrors.ErrCommentNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case apperrors.ErrCommentForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c)
		}
		return
	}

	response.OK(c, "Comment updated successfully", comment)
}

func (h *CommentHandler) Delete(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid comment id")
		return
	}

	userID := c.MustGet("userID").(uint)

	err = h.service.DeleteComment(uint(commentID), userID)
	if err != nil {
		switch err {
		case apperrors.ErrCommentNotFound:
			response.Error(c, http.StatusNotFound, err.Error())
		case apperrors.ErrCommentForbidden:
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.InternalServerError(c)
		}
		return
	}

	response.OK(c, "Comment deleted successfully", nil)
}
