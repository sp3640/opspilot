package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/response"
)

func parsePagination(c *gin.Context, allowedSortFields ...string) (*models.PaginationRequest, bool) {
	req, err := models.ParsePagination(c.Request.URL.Query(), allowedSortFields...)
	if err != nil {
		response.BadRequest(c, err.Error())
		return nil, false
	}

	return req, true
}

func parseOptionalUUID(c *gin.Context, key string) (uuid.UUID, bool) {
	rawValue, provided := c.GetQuery(key)
	if !provided || strings.TrimSpace(rawValue) == "" {
		return uuid.Nil, true
	}

	parsed, err := uuid.Parse(strings.TrimSpace(rawValue))
	if err != nil {
		response.BadRequest(c, "invalid "+key)
		return uuid.Nil, false
	}

	return parsed, true
}

func parseOptionalUint(c *gin.Context, key string) (uint, bool) {
	rawValue, provided := c.GetQuery(key)
	if !provided || strings.TrimSpace(rawValue) == "" {
		return 0, true
	}

	parsed, err := strconv.ParseUint(strings.TrimSpace(rawValue), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid "+key)
		return 0, false
	}

	return uint(parsed), true
}

func parseOrganizationIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	rawValue, exists := c.Get("organizationID")
	if !exists {
		response.Unauthorized(c, "missing organization context")
		return uuid.Nil, false
	}

	value, ok := rawValue.(string)
	if !ok {
		response.Unauthorized(c, "invalid organization context")
		return uuid.Nil, false
	}

	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		response.Unauthorized(c, fmt.Sprintf("invalid organization context: %v", err))
		return uuid.Nil, false
	}

	return parsed, true
}
