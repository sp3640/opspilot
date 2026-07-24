package handlers

import (
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
