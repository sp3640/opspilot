// Package rbac is the single source of truth for OpsPilot's role and
// permission model. Role/permission string values are defined here once and
// re-exported (not redefined) by other packages that need them.
package rbac

import "strings"

// Role identifies one of OpsPilot's four platform-wide roles.
type Role string

const (
	RolePlatformAdmin  Role = "Platform Admin"
	RoleDevOpsEngineer Role = "DevOps Engineer"
	RoleDeveloper      Role = "Developer"
	RoleViewer         Role = "Viewer"
)

// AllRoles lists every valid role value, in privilege order.
var AllRoles = []Role{RolePlatformAdmin, RoleDevOpsEngineer, RoleDeveloper, RoleViewer}

// IsValidRole reports whether value (trimmed, case-insensitive) matches one
// of the four canonical roles.
func IsValidRole(value string) bool {
	normalized := strings.TrimSpace(value)
	for _, role := range AllRoles {
		if strings.EqualFold(normalized, string(role)) {
			return true
		}
	}

	return false
}
