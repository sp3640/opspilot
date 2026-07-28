package handlers

import (
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
