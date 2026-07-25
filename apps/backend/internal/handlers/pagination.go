package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

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

func parseOptionalPositiveUint(c *gin.Context, key string) (uint, bool) {
	rawValue, provided := c.GetQuery(key)
	if !provided || strings.TrimSpace(rawValue) == "" {
		return 0, true
	}

	parsed, err := strconv.ParseUint(strings.TrimSpace(rawValue), 10, 0)
	if err != nil || parsed == 0 {
		response.BadRequest(c, "invalid "+key)
		return 0, false
	}

	return uint(parsed), true
}
