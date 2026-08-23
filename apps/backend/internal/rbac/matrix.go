package rbac

import "strings"

var allPermissions = []Permission{
	PermissionProjectRead, PermissionProjectCreate, PermissionProjectUpdate, PermissionProjectDelete,
	PermissionTeamRead, PermissionTeamManage,
	PermissionInvitationRead, PermissionInvitationCreate, PermissionInvitationRevoke,
	PermissionProjectTeamRead, PermissionProjectTeamManage,
	PermissionClusterRead, PermissionClusterManage, PermissionClusterValidate, PermissionClusterSetDefault,
	PermissionApplicationRead, PermissionApplicationManage,
	PermissionApplicationTeamRead, PermissionApplicationTeamManage,
	PermissionDeploymentRead, PermissionDeploymentCreate, PermissionDeploymentUpdate, PermissionDeploymentCancel, PermissionDeploymentRollback, PermissionDeploymentDelete,
	PermissionIncidentRead, PermissionIncidentManage, PermissionIncidentComment,
	PermissionAlertRead, PermissionAlertManage, PermissionAlertAcknowledge, PermissionAlertResolve, PermissionAlertReopen,
	PermissionKubernetesRead, PermissionKubernetesOperate,
	PermissionAuditRead,
	PermissionOrganizationRead, PermissionOrganizationManage,
	PermissionUserManage,
	PermissionNotificationRead, PermissionNotificationManage,
	PermissionSLORead, PermissionSLOManage,
}

// rolePermissions is the authoritative role→permission matrix. Platform Admin
// always holds every permission; the other three roles are enumerated
// explicitly per the approved RBAC design rather than derived, so the matrix
// stays readable and reviewable as a single block.
var rolePermissions = map[Role]map[Permission]bool{
	RolePlatformAdmin: toSet(allPermissions),

	RoleDevOpsEngineer: toSet([]Permission{
		PermissionProjectRead,
		PermissionTeamRead,
		PermissionProjectTeamRead,
		PermissionClusterRead, PermissionClusterManage, PermissionClusterValidate, PermissionClusterSetDefault,
		PermissionApplicationRead, PermissionApplicationManage,
		PermissionApplicationTeamRead,
		PermissionDeploymentRead, PermissionDeploymentCreate, PermissionDeploymentUpdate, PermissionDeploymentCancel, PermissionDeploymentRollback, PermissionDeploymentDelete,
		PermissionIncidentRead, PermissionIncidentManage, PermissionIncidentComment,
		PermissionAlertRead, PermissionAlertManage, PermissionAlertAcknowledge, PermissionAlertResolve, PermissionAlertReopen,
		PermissionKubernetesRead, PermissionKubernetesOperate,
		PermissionAuditRead,
		PermissionOrganizationRead,
		PermissionNotificationRead, PermissionNotificationManage,
		PermissionSLORead, PermissionSLOManage,
	}),

	RoleDeveloper: toSet([]Permission{
		PermissionProjectRead,
		PermissionTeamRead,
		PermissionProjectTeamRead,
		PermissionClusterRead,
		PermissionApplicationRead,
		PermissionApplicationTeamRead,
		PermissionDeploymentRead,
		PermissionIncidentRead, PermissionIncidentManage, PermissionIncidentComment,
		PermissionAlertRead, PermissionAlertManage, PermissionAlertAcknowledge, PermissionAlertResolve, PermissionAlertReopen,
		PermissionKubernetesRead,
		PermissionAuditRead,
		PermissionOrganizationRead,
		PermissionNotificationRead,
		PermissionSLORead,
	}),

	RoleViewer: toSet([]Permission{
		PermissionProjectRead,
		PermissionTeamRead,
		PermissionProjectTeamRead,
		PermissionClusterRead,
		PermissionApplicationRead,
		PermissionApplicationTeamRead,
		PermissionDeploymentRead,
		PermissionIncidentRead,
		PermissionAlertRead,
		PermissionKubernetesRead,
		PermissionAuditRead,
		PermissionOrganizationRead,
		PermissionNotificationRead,
		PermissionSLORead,
	}),
}

func toSet(permissions []Permission) map[Permission]bool {
	set := make(map[Permission]bool, len(permissions))
	for _, permission := range permissions {
		set[permission] = true
	}

	return set
}

// HasPermission reports whether role (any casing/whitespace) grants
// permission per the matrix above. An unrecognized role grants nothing.
func HasPermission(role string, permission Permission) bool {
	normalized := strings.TrimSpace(role)
	for candidate, permissions := range rolePermissions {
		if strings.EqualFold(normalized, string(candidate)) {
			return permissions[permission]
		}
	}

	return false
}
