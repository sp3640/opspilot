package authorization

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/response"
)

func HasRole(role string, allowedRoles ...string) bool {
	normalized := strings.TrimSpace(role)
	for _, allowed := range allowedRoles {
		if strings.EqualFold(normalized, strings.TrimSpace(allowed)) {
			return true
		}
	}

	return false
}

func RequireRole(c *gin.Context, allowedRoles ...string) bool {
	if !HasRole(c.GetString("role"), allowedRoles...) {
		response.Error(c, 403, "forbidden")
		return false
	}

	return true
}

func RequirePlatformAdmin(c *gin.Context) bool {
	return RequireRole(c, models.RolePlatformAdmin)
}

func RequireOrganizationMember(c *gin.Context) bool {
	raw := strings.TrimSpace(c.GetString("organizationID"))
	if raw == "" {
		response.Error(c, 403, "forbidden")
		return false
	}

	parsed, err := uuid.Parse(raw)
	if err != nil || parsed == uuid.Nil {
		response.Error(c, 403, "forbidden")
		return false
	}

	return true
}
