package authorization

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sp3640/opspilot/backend/internal/models"
	"github.com/sp3640/opspilot/backend/internal/rbac"
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

// HasPermission reports whether role grants permission, per the centralized
// role→permission matrix in internal/rbac. This is the single mechanism
// handlers should use for capability checks going forward, in place of
// scattered role-string comparisons.
func HasPermission(role string, permission rbac.Permission) bool {
	return rbac.HasPermission(role, permission)
}

// RequirePermission aborts the request with 403 Forbidden unless the
// authenticated user's role grants permission. It is independent of
// RequireOrganizationMember — callers that need both must call both.
func RequirePermission(c *gin.Context, permission rbac.Permission) bool {
	if !HasPermission(c.GetString("role"), permission) {
		response.Error(c, 403, "forbidden")
		return false
	}

	return true
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
